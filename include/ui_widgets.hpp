#ifndef UI_WIDGETS_HPP
#define UI_WIDGETS_HPP

#include "imgui.h"
#include <vector>
#include <string>

std::vector<std::string> input_tags;
#define TAG_BUFF_SIZE 128
static char inputBuf[TAG_BUFF_SIZE] = "";
static bool shouldAddTag = false;
static bool shouldFocus = false;
static bool shouldClear = false;

namespace widget {

// Callback function to handle comma input
static int InputTextCallback(ImGuiInputTextCallbackData* data) {
    if (data->EventChar == ',') {
        // Signal that we should add a tag and clear
        shouldAddTag = true;
        shouldClear = true;
        // data->DeleteChars(0, data->BufTextLen);
        // Return 1 to filter out the comma character
        return 1;
    }
    return 0;
}

void DrawTagInput() {
    ImGui::Text("Tags");
    ImGui::NewLine();

    // Render existing input_tags
    for (size_t i = 0; i < input_tags.size(); i++) {
        ImGui::SameLine();
        ImGui::PushStyleColor(ImGuiCol_Button, ImVec4(0.3f, 0.6f, 1.0f, 0.8f));
        ImGui::PushStyleColor(ImGuiCol_ButtonHovered, ImVec4(0.2f, 0.5f, 0.9f, 1.0f));
        ImGui::PushStyleColor(ImGuiCol_ButtonActive, ImVec4(0.2f, 0.5f, 0.9f, 1.0f));
        if (ImGui::Button((input_tags[i] + " x").c_str())) {
            input_tags.erase(input_tags.begin() + i); // delete tag
            i--;
        }
        ImGui::PopStyleColor(3);
    }

    ImGui::SameLine();
    ImGui::PushItemWidth(200);

    // Set focus if requested
    if (shouldFocus) {
        ImGui::SetKeyboardFocusHere();
        shouldFocus = false;
    }

    // Clear buffer if requested (must be done before InputText)
    if (shouldClear) {
        memset(inputBuf, 0, TAG_BUFF_SIZE);
        shouldClear = false;
    }

    if (ImGui::InputText("##taginput", inputBuf, IM_ARRAYSIZE(inputBuf),
                         ImGuiInputTextFlags_EnterReturnsTrue | ImGuiInputTextFlags_CallbackCharFilter,
                         InputTextCallback)) {
        if (strlen(inputBuf) > 0) {
            input_tags.push_back(inputBuf);
            memset(inputBuf, 0, TAG_BUFF_SIZE);
        }
        // Request focus for next frame after Enter
        shouldFocus = true;
    }

    // handle comma-triggered tag addition
    if (shouldAddTag) {
        if (strlen(inputBuf) > 0) {
            input_tags.push_back(inputBuf);
        }
        shouldAddTag = false;
        // Request focus for next frame after comma
        shouldFocus = true;
        memset(inputBuf, 0, TAG_BUFF_SIZE);
    }

    ImGui::PopItemWidth();
}

}

#endif // UI_WIDGETS_HPP