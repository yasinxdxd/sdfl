#ifndef SDF_APP_HPP
#define SDF_APP_HPP

#include <imgui.h>
#include <imgui_impl_glfw.h>
#include <imgui_impl_opengl3.h>
#include <window.hpp>
#include <iostream>
#include <imstyle.hpp>
#include <JetBrainsMonoNL-Light_ttf.hpp>


ImFont* fontBig;
ImFont* fontSml;

void InitImgui(yt2d::Window& window) {
    // Setup Dear ImGui context
    IMGUI_CHECKVERSION();
    ImGui::CreateContext();
    ImGuiIO& io = ImGui::GetIO(); (void)io;
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableKeyboard;     // Enable Keyboard Controls
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableGamepad;      // Enable Gamepad Controls

    fontBig = ImGui::GetIO().Fonts->AddFontFromMemoryTTF(JetBrainsMonoNL_Light_ttf, JetBrainsMonoNL_Light_ttf_len, 18.0f);
    // fontSml = ImGui::GetIO().Fonts->AddFontFromFileTTF("JetBrainsMonoNL-Light.ttf", 18.0f);
    // Setup Dear ImGui style
    ImguiStyle();
    // ImGui::StyleColorsDark();
    // ImGui::StyleColorsLight();

    // Setup Platform/Renderer backends
    if (!ImGui_ImplGlfw_InitForOpenGL(window, true)) {
        std::cout << "ERROR: ImGui_ImplGlfw_InitForOpenGL" << std::endl;
    }
    #ifdef __EMSCRIPTEN__
    ImGui_ImplGlfw_InstallEmscriptenCallbacks(window, "#canvas");
    #endif
    if (!ImGui_ImplOpenGL3_Init("#version 330 core")) {
        std::cout << "ERROR: ImGui_ImplOpenGL3_Init" << std::endl;
    }

    // glfwSetScrollCallback(window, _priv::callbacks::scroll_callback);
}

void DestroyImgui() {
    ImGui_ImplOpenGL3_Shutdown();
    ImGui_ImplGlfw_Shutdown();
    ImGui::DestroyContext();
}

#endif // SDF_APP_HPP