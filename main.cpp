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

#include <head_tracker.hpp>

#include <ui_widgets.hpp>

Texture2D* screenTexture;
RenderTexture2D screenRenderTexture;
ImVec2 screenSize;
#define DESCRIPTION_BUFF_SIZE 1024
#define NAME_BUFF_SIZE 256
char description_buff[DESCRIPTION_BUFF_SIZE] = {0};
char name_buff[NAME_BUFF_SIZE] = {0};

std::string sdfl_file_name;
std::vector<ProgramMetaData> sdfl_programs;
std::vector<Texture2D*> program_preview_textures;


constexpr int resolution = 64;
int workGroupsPerAxis = (resolution + 7) / 8; // ceil

std::vector<Texture2D*> create_preview_textures(std::vector<ProgramMetaData> programs) {
    std::vector<Texture2D*> textures;
    for (size_t i = 0; i < programs.size(); i++) {
        Texture2D* texture = new Texture2D();
        texture->load_texture_from_memory(programs[i].preview_image.data(), programs[i].preview_image.size());
        texture->generate_texture();
        textures.push_back(std::move(texture));
    }
    return textures;
}

void delete_preview_textures() {
    for (size_t i = 0; i < program_preview_textures.size(); i++) {
        delete program_preview_textures[i];
    }
}


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
    GENERATE_STATE,
};
PageState page_state = PageState::FILE_STATE;


void createRendererWindow() {
    ImGuiWindowFlags window_flags = ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoCollapse;
    ImGui::Begin("Renderer", nullptr, window_flags);
    {
        // Back button in top left corner
        ImGui::PushStyleVar(ImGuiStyleVar_FramePadding, ImVec2(4, 2)); // Make button smaller
        if (ImGui::Button("< Back")) {
            page_state = PageState::FILE_STATE;
        }
        ImGui::PopStyleVar();
        
        // Add some spacing between button and image
        ImGui::Spacing();

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
            // debug
            for (auto& t : input_tags) {
                std::cout << t << std::endl;
            }
            publish_program(name_buff, description_buff, code, sequence, jpeg_data, input_tags);
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

void createGenerateWindow() {
    static int selected = 0;
    ImGui::Begin("Generate");
    {
        ImVec2 size = ImGui::GetWindowSize();
        ImGui::BeginChild("##Radiobuttons", ImVec2(size.x, size.y * 0.2));
        if (ImGui::RadioButton("Random Generate", selected == 0)) selected = 0;
        if (ImGui::RadioButton("Random Mixed Generate", selected == 1)) selected = 1;
        if (ImGui::RadioButton("Text Base Generate", selected == 2)) selected = 2;
        ImGui::EndChild();
        ImGui::NewLine();
        if (selected == 0) {
            if (ImGui::Button("Generate Random Model", {256, 28})) {
                generate_random_program();
                launch_process_blocking({"./sdfl/sdflc", "--seq", "generated_random_sequence.seq"});
            }
        } else if (selected == 1) {
            if (ImGui::Button("Generate Mixed Random Model", {256, 28})) {
                generate_mixed_random_program();
                launch_process_blocking({"./sdfl/sdflc", "--seq", "generated_mixed_random_sequence.seq"});
            }
        } else {
            char promptBuff[256] = {0};
            ImGui::InputTextWithHint("##prompt", "Prompt", promptBuff, 256);
            ImGui::NewLine();
            if (ImGui::Button("Generate Text Based Model", {256, 28})) {
                
            }
        }
    }
    ImGui::End();
}

void createFileWindow(const yt2d::Window& window) {
    ImGuiWindowFlags window_flags = ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoCollapse;
    if (ImGui::Begin("Open File", nullptr, window_flags))
    {
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

            // generate page button
            float padding = 20.0f; // distance from window edges
            ImVec2 buttonSize = ImVec2(120, 40); // button size
            ImVec2 buttonPos = ImVec2(
                windowSize.x - buttonSize.x - padding,
                windowSize.y - buttonSize.y - padding
            );

            ImGui::SetCursorPos(buttonPos);
            if (ImGui::Button("Generate", buttonSize)) {
                page_state = PageState::GENERATE_STATE;
            }
        }
        const std::vector<std::string> files = window.getDraggedPaths();
        if (files.size() == 1) {
            sdfl_file_name = files[0];
            std::cout << "FILE_SIZE: " << files.size() << std::endl;
            if (SDFLCompiler.is_running()) {
                SDFLCompiler.stop_watch();
            }
            SDFLCompiler.start_watch(sdfl_file_name);
            page_state = PageState::RENDER_STATE;
        }
    }
    ImGui::End();
}

int programs_columns = 1;
void createProgramsWindow(const yt2d::Window& window) {
    ImGuiWindowFlags window_flags = ImGuiWindowFlags_NoMove | ImGuiWindowFlags_NoCollapse;
    if (ImGui::Begin("SDFL Programs", nullptr, window_flags))
    {
        ImGui::SliderInt("columns", &programs_columns, 1, 8);
        ImGui::SameLine();

        // put button at the end of row
        float buttonWidth = ImGui::CalcTextSize("Refresh").x + ImGui::GetStyle().FramePadding.x * 2;
        float fullWidth = ImGui::GetContentRegionAvail().x;
        ImGui::SetCursorPosX(ImGui::GetCursorPosX() + fullWidth - buttonWidth);
        if (ImGui::Button("Refresh")) {
            update_cache();
            delete_preview_textures();
            sdfl_programs = get_programs_from_cache();
            program_preview_textures = create_preview_textures(sdfl_programs);
        }

        int cols = programs_columns * 2;
        ImGui::Columns(cols);
        char name[128];
        for (size_t i = 0; i < sdfl_programs.size(); i++) {
            if (i % cols == 0) ImGui::Separator();
            ImGui::Text("Name: %s", sdfl_programs[i].name.c_str());
            ImGui::Text("Date: %s", sdfl_programs[i].created_at.c_str());            
            
            ImTextureID textureID = (ImTextureID)(intptr_t)((unsigned int)(*program_preview_textures[i]));
            float ratio = (float)program_preview_textures[i]->get_height() / program_preview_textures[i]->get_width();
            ImVec2 size = ImGui::GetContentRegionAvail();
            char programName[64];
            sprintf(programName, "##program%ld", i);
            if (ImGui::ImageButton(programName, textureID, {size.x, size.x * ratio}, ImVec2(0, 1), ImVec2(1, 0))) {

            }
            

            char tagLabels[256];
            ImGui::Text("Tags: ");
            for (size_t j = 0; j < sdfl_programs[i].tags.size(); j++) {
                sprintf(tagLabels, "%s##tags%ld%ld", (sdfl_programs[i].tags[j]).c_str(), i, j);
                ImGui::SameLine();
                ImGui::PushStyleColor(ImGuiCol_Button, ImVec4(0.4f, 0.6f, 1.0f, 0.8f));
                ImGui::PushStyleColor(ImGuiCol_ButtonHovered, ImVec4(0.1f, 0.8f, 0.9f, 1.0f));
                ImGui::PushStyleColor(ImGuiCol_ButtonActive, ImVec4(0.1f, 0.5f, 0.7f, 1.0f));
                // ImGui::LabelText(tagLabels, "%s", );
                ImGui::Button(tagLabels);
                ImGui::PopStyleColor(3);
            }
            
            
            ImGui::NextColumn();
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
        createProgramsWindow(window);
        createFileWindow(window);
        break;
    case PageState::RENDER_STATE:
        createRendererWindow();
        createExportWindow();
        createInfoWindow();
        ht::draw_cam_frame();
        break;
    case PageState::GENERATE_STATE:
        createRendererWindow();
        createGenerateWindow();
        ht::draw_cam_frame();
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

    sdfl_programs = get_programs_from_cache();
    program_preview_textures = create_preview_textures(sdfl_programs);

    // head_tracker
    if (!ht::init()) {
        std::cout << "ERROR: Head tracker cannot be initialized"<< std::endl;
    }

    
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
        if (page_state == PageState::RENDER_STATE || page_state == PageState::GENERATE_STATE) {    
            // render
            render(quad, 6, shader, [&](Shader* shader) {
                // send uniforms
                shader->set<int, 2>("window_size", screenSize.x, screenSize.y);
                shader->set<float, 1>("elapsed_time", elapsed_time);
                shader->set<bool, 1>("ht_tracking_enabled", ht::is_head_tracking_enabled());
                cv::Point3f hc = ht::get_head_center() * 32.0f; // TODO: effect can be
                // printf("head_center: %f, %f, %f\n", hc.x, hc.y, hc.z);
                float hcz = (2.5f * hc.z);
                shader->set<float, 3>("ht_head_center", hc.x, hc.y, hcz);
            });
        }
        screenRenderTexture.unbind();

        renderMainUI(window);
        
        ImGui::Render();
        ImGui_ImplOpenGL3_RenderDrawData(ImGui::GetDrawData());
        window.display();
        
        std::chrono::steady_clock::time_point end = std::chrono::steady_clock::now();

        frame_time = std::chrono::duration_cast<std::chrono::milliseconds>(end - begin).count();
        elapsed_time += frame_time / 1000.0;
    }

    if (SDFLCompiler.is_running()) {
        SDFLCompiler.stop_watch();
    }

    // shutdown server
    shutdown_server();

    delete_preview_textures();

    delete shader;
    delete screenTexture;
    glcompiler::destroy();
    DestroyImgui();

    ht::destroy();

    return 0;
}