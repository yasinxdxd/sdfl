#include <window.hpp>
#include <shader.hpp>
#include <shader_compiler.hpp>

#define SHAPE_QUAD
#include <renderables.hpp>

#include <chrono>
#include <iostream>
#include <file_watcher.hpp>


int main(void) {
    yt2d::Window window("hello", 480, 340);

    /* shader */
    glcompiler::init();
    Shader* shader = new Shader("sdfl/out_frag.glsl", Shader::ShaderCodeType::FRAGMENT_SHADER);
    glcompiler::compile_and_attach_shaders(shader);

    FileWatcher fw("sdfl/out_frag.glsl");

    Quad quad;

    int frame_time = 0;
    float elapsed_time = 0.f;
    
    while (!window.isClose()) {
        std::chrono::steady_clock::time_point begin = std::chrono::steady_clock::now();
        window.pollEvent();


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
    glcompiler::destroy();
}