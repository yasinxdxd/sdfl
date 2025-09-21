#include <window.hpp>
#include <texture2d.hpp>
#include <render_texture2d.hpp>
#include <input.hpp>
#include <shader.hpp>
#include <ssbo.hpp>
#include <shader_compiler.hpp>

#define SHAPE_QUAD
#include <renderables.hpp>

#include <stb_image_write.h>

#include <chrono>
#include <iostream>
#include <fstream>
#include <file_watcher.hpp>
#include <sdf_app.hpp>
#include <sdf_client.hpp>

#include <ui_widgets.hpp>

Texture2D* screenTexture;
RenderTexture2D screenRenderTexture;
ImVec2 screenSize;
#define DESCRIPTION_BUFF_SIZE 1024
#define NAME_BUFF_SIZE 256
char description_buff[DESCRIPTION_BUFF_SIZE] = {0};
char name_buff[NAME_BUFF_SIZE] = {0};

std::string sdfl_file_name;


int resolution = 256;
int workGroupsPerAxis = (resolution + 7) / 8; // ceil


void writeSDFToBinary(const std::vector<float>& sdfData, int resolution, 
                      const std::string& filename) {
    std::ofstream file(filename, std::ios::binary);
    
    // Write header info
    file.write(reinterpret_cast<const char*>(&resolution), sizeof(int));
    
    // Write SDF data
    file.write(reinterpret_cast<const char*>(sdfData.data()), 
               sdfData.size() * sizeof(float));
    
    file.close();
    std::cout << "Exported SDF data to " << filename << std::endl;
}

void exportSDFData(const char* filepath) {
    // create SSBO for SDF output
    SSBO sdfBuffer(0); // binding point 0 // inside glsl: layout(std430, binding = 0)
    size_t bufferSize = resolution * resolution * resolution * sizeof(float);
    sdfBuffer.initialize(bufferSize);
    
    Shader* computeShader = new Shader("out_compute.glsl", Shader::ShaderCodeType::COMPUTE_SHADER);
    glcompiler::compile_and_attach_shaders(computeShader);
    Shader::dispatch_compute(computeShader, workGroupsPerAxis, workGroupsPerAxis, workGroupsPerAxis, [&](Shader* shader) {
        computeShader->set<float, 3>("minBound", -8.0f, -8.0f, -8.0f);
        computeShader->set<float, 3>("maxBound", 8.0f, 8.0f, 8.0f);
        computeShader->set<int>("resolution", resolution);
    });
    
    sdfBuffer.waitForComputeShader();
    
    // get SDF data from gpu
    std::vector<float> sdfData = sdfBuffer.downloadData<float>();
    
    // std::cout << "Downloaded " << sdfData.size() << " SDF values" << std::endl;
    
    writeSDFToBinary(sdfData, resolution, filepath);
    delete computeShader;
}

// callback to write bytes into vector
void write_to_vector(void* context, void* data, int size) {
    std::vector<unsigned char>* out = reinterpret_cast<std::vector<unsigned char>*>(context);
    unsigned char* bytes = reinterpret_cast<unsigned char*>(data);
    out->insert(out->end(), bytes, bytes + size);
}

std::vector<unsigned char> encode_texture_to_jpeg(Texture2D* tex, int quality = 90) {
    int width = tex->get_width();
    int height = tex->get_height();
    int channels = tex->get_channels();
    const unsigned char* data = tex->read_back_from_gpu();
    std::cout << width << std::endl;
    std::cout << height << std::endl;
    std::cout << channels << std::endl;

    std::vector<unsigned char> jpeg_buffer;
    if (!data || width == 0 || height == 0 || channels == 0) {
        std::cerr << "Invalid texture data\n";
        return jpeg_buffer;
    }

    // encode JPEG into memory
    if (!stbi_write_jpg_to_func(write_to_vector, &jpeg_buffer, width, height, channels, data, quality)) {
        std::cerr << "Failed to encode JPEG\n";
    }
    
    delete[] data;

    return jpeg_buffer; // this can be sent over HTTP
}

std::string read_file_to_string(const std::string& filepath) {
    std::ifstream file(filepath, std::ios::in | std::ios::binary);
    if (!file) {
        throw std::runtime_error("Failed to open file: " + filepath);
    }

    std::ostringstream buffer;
    buffer << file.rdbuf();
    return buffer.str();
}

enum class PageState {
    RENDER_STATE,
    FILE_STATE,
};
PageState page_state = PageState::FILE_STATE;


void createRendererWindow() {
    ImGui::Begin("Renderer");
    {
        ImTextureID textureID = (ImTextureID)(intptr_t)((unsigned int)(*screenRenderTexture.get_texture()));
        ImVec2 size = ImGui::GetContentRegionAvail(); // Use available space instead of window size
        ImGui::Image(textureID, size, ImVec2(0, 1), ImVec2(1, 0)); // Flip UVs for correct orientation

        // get correct mouse coords:
        ImVec2 mouseScreen = ImGui::GetMousePos();
        ImVec2 imagePos = ImGui::GetItemRectMin(); // Position of top-left corner of ImGui::Image
        screenSize = ImGui::GetItemRectSize(); // Size of the image
        // mouseInImageDisplay = ImVec2(mouseScreen.x - imagePos.x, mouseScreen.y - imagePos.y);
    }
    ImGui::End();
}

void createInfoWindow() {
    ImGui::Begin("Info");
    {
        ImVec2 availableSize = ImGui::GetContentRegionAvail();
        ImGui::Text("Name");
        ImGui::InputText("##name_input", name_buff, NAME_BUFF_SIZE);
        ImGui::Text("Description");
        ImGui::InputTextMultiline("##description_input", description_buff, DESCRIPTION_BUFF_SIZE, 
                                 ImVec2(availableSize.x * 0.9f, 128), ImGuiInputTextFlags_WordWrap);
        widget::DrawTagInput();
        if (ImGui::Button("Publish", {64, 28})) {
            Texture2D* tex = screenRenderTexture.get_texture();
            std::vector<unsigned char> jpeg_data = encode_texture_to_jpeg(tex, 90);

            std::string code;
            std::string sequence;
            try {
                code = read_file_to_string(sdfl_file_name);
                sequence = read_file_to_string("ast_sequence.txt");

                std::cout << "Code length: " << code.size() << "\n";
                std::cout << "Sequence length: " << sequence.size() << "\n";
            } catch (const std::exception& e) {
                std::cerr << e.what() << std::endl;
            }

            publish_program(name_buff, description_buff, code, sequence, jpeg_data);
        }
    }
    ImGui::End();
}

void createExportWindow() {
    ImGui::Begin("Export");
    {
        if (ImGui::Button("Export SDF", {120, 28})) {
            // for now hardcoded path
            exportSDFData("testpy/test_sdf_data.bin");
        }
    }
    ImGui::End();
}

void createFileWindow(const yt2d::Window& window) {
    ImGui::Begin("Open File");
    {
        // ImGui::Text("Drag and Drop an .sdfl file");
        {
            const char* text = "Drag and Drop an .sdfl file";
            ImGui::SetWindowFontScale(2.0f); // 2x bigger
            // recalculate text size
            ImVec2 windowSize = ImGui::GetWindowSize();
            ImVec2 textSize = ImGui::CalcTextSize(text);
            
            // center
            ImGui::SetCursorPosX((windowSize.x - textSize.x) * 0.5f);
            ImVec2 availableSize = ImGui::GetContentRegionAvail();
            ImGui::SetCursorPosY(ImGui::GetCursorPosY() + availableSize.y * 0.4f);
            
            ImGui::Text("%s", text);
            // reset font scale
            ImGui::SetWindowFontScale(1.0f);

        }
        const std::vector<std::string> files = window.getDraggedPaths();
        if (files.size() == 1) {
            sdfl_file_name = files[0];
            launch_process_blocking({"./sdfl/sdflc", sdfl_file_name});
            std::thread([=]() {
                launch_process_blocking({"./sdfl/sdflc", sdfl_file_name, "--watch", "--interval=0"});
            }).detach();
            page_state = PageState::RENDER_STATE;
        }
    }
    ImGui::End();
}

void renderMainUI(const yt2d::Window& window) {    
    ImGuiViewport* viewport = ImGui::GetMainViewport();
    ImGui::SetNextWindowPos(viewport->WorkPos);
    ImGui::SetNextWindowSize(viewport->WorkSize);
    ImGui::SetNextWindowViewport(viewport->ID);
    
    ImGuiWindowFlags window_flags = ImGuiWindowFlags_MenuBar | ImGuiWindowFlags_NoDocking;
    window_flags |= ImGuiWindowFlags_NoTitleBar | ImGuiWindowFlags_NoCollapse;
    window_flags |= ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoMove;
    window_flags |= ImGuiWindowFlags_NoBringToFrontOnFocus | ImGuiWindowFlags_NoNavFocus;
    
    // create the main dockspace
    ImGui::PushStyleVar(ImGuiStyleVar_WindowRounding, 0.0f);
    ImGui::PushStyleVar(ImGuiStyleVar_WindowBorderSize, 0.0f);
    ImGui::PushStyleVar(ImGuiStyleVar_WindowPadding, ImVec2(0.0f, 0.0f));
    
    ImGui::Begin("MainDockSpace", nullptr, window_flags);
    ImGui::PopStyleVar(3);
    
    ImGuiID dockspace_id = ImGui::GetID("MainDockSpace");
    ImGui::DockSpace(dockspace_id, ImVec2(0.0f, 0.0f), ImGuiDockNodeFlags_None);
    
    ImGui::End();
    
    switch (page_state)
    {
    case PageState::FILE_STATE:
        createFileWindow(window);
        break;
    case PageState::RENDER_STATE:
        createRendererWindow();
        createInfoWindow();
        createExportWindow();
        break;
    
    default:
        break;
    }
}

int main(void) {
    yt2d::Window window("SDF Renderer", 1280, 720);
    InitImgui(window);

    screenTexture = new Texture2D(640, 480, nullptr);
    screenTexture->generate_texture();
    screenRenderTexture.set_texture(screenTexture);

    /* shader */
    glcompiler::init();
    Shader* shader = new Shader("out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
    glcompiler::compile_and_attach_shaders(shader);

    // start the local bridge server
    std::thread([=]() {
        launch_process_blocking({"./bridge/bridge"});
    }).detach();

    FileWatcher fw("out_frag.glsl");

    Quad quad;

    int frame_time = 0;
    float elapsed_time = 0.f;
    
    while (!window.isClose()) {
        std::chrono::steady_clock::time_point begin = std::chrono::steady_clock::now();
        window.pollEvent();

        if (Input::isKeyPressed(KeyCode::KEY_ESCAPE)) {
            break;
        }

        if (fw.hasChanged()) {
            delete shader;
            shader = new Shader("out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
            glcompiler::compile_and_attach_shaders(shader);
            std::cout << "FILE CHANGED!!!!"<< std::endl;
        }


        window.clear();
        ImGui_ImplOpenGL3_NewFrame();
        ImGui_ImplGlfw_NewFrame();
        ImGui::NewFrame();

        // draw sdf to texture
        screenRenderTexture.bind();
        window.clear();
        window.setViewport(0, 0, screenRenderTexture.get_texture()->get_width(), screenRenderTexture.get_texture()->get_height());

        // render
        render(quad, 6, shader, [&](Shader* shader) {
            // send uniforms
            shader->set<int, 2>("window_size", screenSize.x, screenSize.y);
            shader->set<float, 1>("elapsed_time", elapsed_time);
        });
        screenRenderTexture.unbind();

        renderMainUI(window);
        
        ImGui::Render();
        ImGui_ImplOpenGL3_RenderDrawData(ImGui::GetDrawData());
        window.display();
        
        std::chrono::steady_clock::time_point end = std::chrono::steady_clock::now();

        frame_time = std::chrono::duration_cast<std::chrono::milliseconds>(end - begin).count();
        elapsed_time += frame_time / 1000.0;
    }

    // shutdown server
    shutdown_server();

    delete shader;
    delete screenTexture;
    glcompiler::destroy();
    DestroyImgui();

    return 0;
}