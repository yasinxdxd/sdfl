#include <window.hpp>
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

    /* shader */
    glcompiler::init();
    Shader* shader = new Shader("sdfl/out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
    glcompiler::compile_and_attach_shaders(shader);


    int resolution = 64;
    
    // create SSBO for SDF output
    SSBO sdfBuffer(0); // binding point 0 // inside glsl: layout(std430, binding = 0)
    size_t bufferSize = resolution * resolution * resolution * sizeof(float);
    sdfBuffer.initialize(bufferSize);
    
    Shader* computeShader = new Shader("sdfl/out_compute.glsl", Shader::ShaderCodeType::COMPUTE_SHADER);
    glcompiler::compile_and_attach_shaders(computeShader);
    Shader::dispatch_compute(computeShader, 8, 8, 8, [&](Shader* shader) {
        computeShader->set<float, 3>("minBound", -10.0f, -10.0f, -10.0f);
        computeShader->set<float, 3>("maxBound", 10.0f, 10.0f, 10.0f);
        computeShader->set<int>("resolution", resolution);
    });
    
    sdfBuffer.waitForComputeShader();
    
    // get SDF data from gpu
    std::vector<float> sdfData = sdfBuffer.downloadData<float>();
    
    std::cout << "Downloaded " << sdfData.size() << " SDF values" << std::endl;
    
    exportSDFToBinary(sdfData, resolution, "testpy/test_sdf_data.bin");

    // for (int z = 0; z < resolution; z++) {
    //     for (int y = 0; y < resolution; y++) {
    //         for (int x = 0; x < resolution; x++) {
    //             int index = z * resolution * resolution + y * resolution + x;
    //             float sdfValue = sdfData[index];
    //         }
    //     }
    // }


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
        // render
        render(quad, 6, shader, [&](Shader* shader) {
            // send uniforms
            shader->set<int, 2>("window_size", window.getWindowWidth(), window.getWindowHeight());
            shader->set<float, 1>("elapsed_time", elapsed_time);
        });
        window.display();
        
        std::chrono::steady_clock::time_point end = std::chrono::steady_clock::now();

        frame_time = std::chrono::duration_cast<std::chrono::milliseconds>(end - begin).count();
        elapsed_time += frame_time / 1000.0;
    }

    delete shader;
    delete computeShader;
    glcompiler::destroy();
}