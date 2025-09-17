#include <window.hpp>
#include <texture2d.hpp>
#include <render_texture2d.hpp>
#include <input.hpp>
#include <shader.hpp>
#include <ssbo.hpp>
#include <shader_compiler.hpp>

#define SHAPE_QUAD
#include <renderables.hpp>

#include <chrono>
#include <iostream>
#include <fstream>
#include <file_watcher.hpp>
#include <sdf_app.hpp>

Texture2D* screenTexture;
RenderTexture2D screenRenderTexture;
ImVec2 screenSize;

void exportSDFToBinary(const std::vector<float>& sdfData, int resolution, 
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

int main(void) {
    yt2d::Window window("hello", 480, 340);
    InitImgui(window);

    screenTexture = new Texture2D(640, 480, nullptr);
    screenTexture->generate_texture();
    screenRenderTexture.set_texture(screenTexture);

    /* shader */
    glcompiler::init();
    Shader* shader = new Shader("sdfl/out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
    glcompiler::compile_and_attach_shaders(shader);


    int resolution = 256;
    int workGroupsPerAxis = (resolution + 7) / 8; // ceil
    
    // create SSBO for SDF output
    SSBO sdfBuffer(0); // binding point 0 // inside glsl: layout(std430, binding = 0)
    size_t bufferSize = resolution * resolution * resolution * sizeof(float);
    sdfBuffer.initialize(bufferSize);
    
    Shader* computeShader = new Shader("sdfl/out_compute.glsl", Shader::ShaderCodeType::COMPUTE_SHADER);
    glcompiler::compile_and_attach_shaders(computeShader);
    Shader::dispatch_compute(computeShader, workGroupsPerAxis, workGroupsPerAxis, workGroupsPerAxis, [&](Shader* shader) {
        computeShader->set<float, 3>("minBound", -8.0f, -8.0f, -8.0f);
        computeShader->set<float, 3>("maxBound", 8.0f, 8.0f, 8.0f);
        computeShader->set<int>("resolution", resolution);
    });
    
    sdfBuffer.waitForComputeShader();
    
    // get SDF data from gpu
    std::vector<float> sdfData = sdfBuffer.downloadData<float>();
    
    std::cout << "Downloaded " << sdfData.size() << " SDF values" << std::endl;
    
    exportSDFToBinary(sdfData, resolution, "testpy/test_sdf_data.bin");



    FileWatcher fw("sdfl/out_frag.glsl");

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
            shader = new Shader("sdfl/out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
            glcompiler::compile_and_attach_shaders(shader);
        }


        window.clear();
        ImGui_ImplOpenGL3_NewFrame();
        ImGui_ImplGlfw_NewFrame();
        ImGui::NewFrame();
        ImGui::Begin("window");
        ImGui::BeginChild("RenderSection");
        {
            ImTextureID textureID = (ImTextureID)(intptr_t)((unsigned int)(*screenRenderTexture.get_texture()));            
            ImVec2 size = ImVec2(ImGui::GetWindowSize());
            ImGui::Image(textureID, size, ImVec2(0, 1), ImVec2(1, 0)); // Flip UVs for correct orientation

            // get correct mouse coords:
            ImVec2 mouseScreen = ImGui::GetMousePos();
            ImVec2 imagePos = ImGui::GetItemRectMin(); // Position of top-left corner of ImGui::Image
            screenSize = ImGui::GetItemRectSize(); // Size of the image
            // mouseInImageDisplay = ImVec2(mouseScreen.x - imagePos.x, mouseScreen.y - imagePos.y);
        }
        ImGui::EndChild();
        ImGui::End();


        screenRenderTexture.bind();
        window.clear();
        window.setViewport(0, 0, screenRenderTexture.get_texture()->getWidth(), screenRenderTexture.get_texture()->getHeight());

        // render
        render(quad, 6, shader, [&](Shader* shader) {
            // send uniforms
            shader->set<int, 2>("window_size", screenSize.x, screenSize.y);
            shader->set<float, 1>("elapsed_time", elapsed_time);
        });

        screenRenderTexture.unbind();


        ImGui::Render();
        ImGui_ImplOpenGL3_RenderDrawData(ImGui::GetDrawData());
        window.display();
        
        std::chrono::steady_clock::time_point end = std::chrono::steady_clock::now();

        frame_time = std::chrono::duration_cast<std::chrono::milliseconds>(end - begin).count();
        elapsed_time += frame_time / 1000.0;
    }

    delete shader;
    delete computeShader;
    delete screenTexture;
    glcompiler::destroy();
    DestroyImgui();
}