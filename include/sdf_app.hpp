#ifndef SDF_APP_HPP
#define SDF_APP_HPP

#include <imgui.h>
#include <imgui_impl_glfw.h>
#include <imgui_impl_opengl3.h>
#include <window.hpp>
#include <iostream>
#include <imstyle.hpp>
#include <JetBrainsMonoNL-Light_ttf.hpp>

#if defined(_WIN32)
#include <windows.h>
#else
#include <unistd.h>
#include <sys/types.h>
#include <sys/wait.h>
#endif

#include <future>
#include <thread>


ImFont* fontBig;
ImFont* fontSml;

void InitImgui(yt2d::Window& window) {
    // Setup Dear ImGui context
    IMGUI_CHECKVERSION();
    ImGui::CreateContext();
    ImGuiIO& io = ImGui::GetIO(); (void)io;
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableKeyboard;     // Enable Keyboard Controls
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableGamepad;      // Enable Gamepad Controls
    io.ConfigFlags |= ImGuiConfigFlags_DockingEnable;


    // ImFontConfig cfg;
    // cfg.FontDataOwnedByAtlas = false;
    // fontBig = ImGui::GetIO().Fonts->AddFontFromMemoryTTF(JetBrainsMonoNL_Light_ttf, JetBrainsMonoNL_Light_ttf_len, 18.0f, &cfg);
    // ImGui::GetIO().Fonts->Build();

    // let imgui to free the font
    unsigned char* fontData = (unsigned char*)malloc(JetBrainsMonoNL_Light_ttf_len);
    memcpy(fontData, JetBrainsMonoNL_Light_ttf, JetBrainsMonoNL_Light_ttf_len);
    ImFontConfig cfg;
    cfg.FontDataOwnedByAtlas = true;  // imgui will free() safely
    fontBig = ImGui::GetIO().Fonts->AddFontFromMemoryTTF(fontData, JetBrainsMonoNL_Light_ttf_len, 18.0f, &cfg);


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

bool launch_process_blocking(const std::vector<std::string>& args) {
    if (args.empty()) return false;

#if defined(_WIN32)
    // Build command line
    std::string cmd;
    for (const auto& arg : args) {
        cmd += "\"" + arg + "\" ";
    }

    STARTUPINFOA si{};
    PROCESS_INFORMATION pi{};
    si.cb = sizeof(si);

    if (!CreateProcessA(
            NULL,
            cmd.data(),   // command line
            NULL,
            NULL,
            FALSE,
            0,
            NULL,
            NULL,
            &si,
            &pi))
    {
        std::cerr << "CreateProcess failed. Error: " << GetLastError() << "\n";
        return false;
    }

    // Wait until child process exits
    WaitForSingleObject(pi.hProcess, INFINITE);

    // Clean up
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);
    return true;

#else
    pid_t pid = fork();
    if (pid == 0) {
        // child
        std::vector<char*> cargs;
        for (const auto& arg : args) {
            cargs.push_back(const_cast<char*>(arg.c_str()));
        }
        cargs.push_back(nullptr);

        execvp(cargs[0], cargs.data());
        perror("execvp failed");
        _exit(1);
    } else if (pid > 0) {
        // parent
        int status;
        waitpid(pid, &status, 0);
        return WIFEXITED(status) && WEXITSTATUS(status) == 0;
    } else {
        perror("fork failed");
        return false;
    }
#endif
}

bool launch_process_with_interrupt(const std::vector<std::string>& args, 
                                 const std::atomic<bool>& stop_flag) {
    if (args.empty()) return false;

#if defined(_WIN32)
    // Build command line
    std::string cmd;
    for (const auto& arg : args) {
        cmd += "\"" + arg + "\" ";
    }

    STARTUPINFOA si{};
    PROCESS_INFORMATION pi{};
    si.cb = sizeof(si);

    if (!CreateProcessA(
            NULL,
            cmd.data(),
            NULL,
            NULL,
            FALSE,
            0,
            NULL,
            NULL,
            &si,
            &pi))
    {
        std::cerr << "CreateProcess failed. Error: " << GetLastError() << "\n";
        return false;
    }

    // Wait with periodic checks for stop flag
    bool process_finished = false;
    bool killed = false;
    
    while (!process_finished && !killed) {
        DWORD wait_result = WaitForSingleObject(pi.hProcess, 100); // 100ms timeout
        
        if (wait_result == WAIT_OBJECT_0) {
            // Process finished naturally
            process_finished = true;
        } else if (wait_result == WAIT_TIMEOUT) {
            // Check if we should stop
            if (stop_flag.load()) {
                // Terminate the process
                TerminateProcess(pi.hProcess, 1);
                WaitForSingleObject(pi.hProcess, INFINITE); // Wait for termination
                killed = true;
            }
        } else {
            // Error occurred
            std::cerr << "WaitForSingleObject failed. Error: " << GetLastError() << "\n";
            break;
        }
    }

    // Clean up
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);
    return process_finished && !killed;

#else
    pid_t pid = fork();
    if (pid == 0) {
        // child process
        std::vector<char*> cargs;
        for (const auto& arg : args) {
            cargs.push_back(const_cast<char*>(arg.c_str()));
        }
        cargs.push_back(nullptr);

        execvp(cargs[0], cargs.data());
        perror("execvp failed");
        _exit(1);
    } else if (pid > 0) {
        // parent process
        int status;
        bool process_finished = false;
        bool killed = false;
        
        while (!process_finished && !killed) {
            // Non-blocking wait
            pid_t result = waitpid(pid, &status, WNOHANG);
            
            if (result == pid) {
                // Process finished
                process_finished = true;
            } else if (result == 0) {
                // Process still running, check stop flag
                if (stop_flag.load()) {
                    // Send SIGTERM first (graceful)
                    kill(pid, SIGTERM);
                    
                    // Give it a moment to terminate gracefully
                    std::this_thread::sleep_for(std::chrono::milliseconds(500));
                    
                    // Check if it's still running
                    result = waitpid(pid, &status, WNOHANG);
                    if (result == 0) {
                        // Still running, force kill
                        kill(pid, SIGKILL);
                        waitpid(pid, &status, 0); // Wait for it to die
                    }
                    killed = true;
                } else {
                    // Brief sleep to avoid busy waiting
                    std::this_thread::sleep_for(std::chrono::milliseconds(100));
                }
            } else {
                // Error in waitpid
                perror("waitpid failed");
                break;
            }
        }
        
        return process_finished && !killed && WIFEXITED(status) && WEXITSTATUS(status) == 0;
    } else {
        perror("fork failed");
        return false;
    }
#endif
}

class SDFLCManager {
private:
    std::thread watch_thread;
    std::atomic<bool> stop_flag{false};
    std::atomic<bool> is_watch_running{false};
    
public:
    void start_watch(const std::string& sdfl_file_name) {        
        // initial compilation
        launch_process_blocking({"./sdfl/sdflc", sdfl_file_name});
        
        // start watch process
        stop_flag = false;
        is_watch_running = true;
        
        watch_thread = std::thread([this, sdfl_file_name]() {
            launch_process_with_interrupt({"./sdfl/sdflc", sdfl_file_name, "--watch", "--interval=1000"}, 
                                        stop_flag);
            is_watch_running = false; // Mark as finished when thread exits
        });
    }
    
    void stop_watch() {
        if (!is_watch_running.load()) {
            return; // Already stopped or never started
        }
        
        stop_flag = true;
        if (watch_thread.joinable()) {
            watch_thread.join();
        }
        // is_watch_running will be set to false by the thread itself
    }
    
    bool is_running() const {
        return is_watch_running.load();
    }
    
    ~SDFLCManager() {
        stop_watch();
    }
};

SDFLCManager SDFLCompiler;

#endif // SDF_APP_HPP