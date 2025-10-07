#ifndef HEAD_TRACKER_HPP
#define HEAD_TRACKER_HPP

#include <algorithm>

#include <opencv2/objdetect/objdetect.hpp>
#include <opencv2/highgui/highgui.hpp>
#include <opencv2/imgproc/imgproc.hpp>
#include <opencv2/calib3d/calib3d.hpp>

#include "imgui.h"
#include "face_detector.h"

namespace ht {

const char* model_path = "./assets/face_detection_front.tflite";
FaceDetector::Detector* detector = nullptr;
cv::VideoCapture *capture = NULL;
cv::Mat frame;

cv::Rect head_rect;
cv::Point2f head_center;
cv::Point3f smoothed_head;
bool ht_tracking_enabled = false;
float smoothing = 0.5f;

Texture2D* camFrameTexture;

int camWidth;
int camHeight;

    bool init() {
        // Initialize face detector
        detector = new FaceDetector::Detector();
        if (!detector->init(model_path, 0.75f, 0.3f)) {
            fprintf(stderr, "ERROR:HEAD_TRACKER: Failed to initialize face detector\n");
            delete detector;
            detector = nullptr;
            return false;
        }

        // start video capture from camera
        capture = new cv::VideoCapture(0);

        // lower resolution
        capture->set(cv::CAP_PROP_FRAME_WIDTH, 320);
        capture->set(cv::CAP_PROP_FRAME_HEIGHT, 240);

        // check that video is working
        if (capture == NULL || !capture->isOpened()) {
            fprintf(stderr, "ERROR:HEAD_TRACKER: Could not start video capture\n");
            detector->destroy();
            delete detector;
            detector = nullptr;
            return false;
        }

        camWidth = 320;
        camHeight = 240;
        printf("camWidth: %d, camHeight: %d\n", camWidth, camHeight);

        camFrameTexture = new Texture2D(camWidth, camHeight, nullptr);
        camFrameTexture->generate_texture();
        return true;
    }

    void draw_cam_frame() {
        (*capture) >> frame;
        if (frame.empty()) return;

        // Detect faces using TensorFlow Lite detector
        std::vector<FaceDetector::FaceRect> faces;
        if (detector && detector->detect(frame, faces, false)) {  // false = no mirror during detection
            
            if (!faces.empty()) {
                // pick largest face
                FaceDetector::FaceRect largest_face = faces[0];
                for (size_t i = 1; i < faces.size(); i++) {
                    float area1 = largest_face.width * largest_face.height;
                    float area2 = faces[i].width * faces[i].height;
                    if (area2 > area1) {
                        largest_face = faces[i];
                    }
                }

                // convert normalized coordinates to pixel coordinates
                head_rect = largest_face.toPixelRect(camWidth, camHeight);

                // compute head center
                head_center = cv::Point2f(head_rect.x + head_rect.width / 2.0f, 
                                         head_rect.y + head_rect.height / 2.0f);

                // draw rectangle and center
                cv::rectangle(frame, head_rect, cv::Scalar(0, 255, 0), 2);
                cv::circle(frame, head_center, 5, cv::Scalar(255, 0, 0), -1);
            }
        }

        // convert to RGB for texture
        cv::cvtColor(frame, frame, cv::COLOR_BGR2RGB);
        camFrameTexture->update_texture(frame.data);

        // draw ImGui mirrored image
        ImTextureID textureID = (ImTextureID)(intptr_t)((unsigned int)(*camFrameTexture));
        // ImGuiWindowFlags window_flags = ImGuiWindowFlags_None;
        
        ImGui::Begin("Head Tracker");
        ImGui::Checkbox("Enable Head Tracking", &ht_tracking_enabled);
        if (ht_tracking_enabled) {
            ImVec2 size = ImGui::GetContentRegionAvail();
            ImGui::Image(textureID, size, ImVec2(1, 0), ImVec2(0, 1)); // mirror horizontally
        }
        ImGui::End();
    }

cv::Point3f get_head_center() {
    static float previous_z = 0.f;
    if (head_rect.width == 0 || head_rect.height == 0)
        return smoothed_head;  // return last known position
    
    float focal_length = camWidth * 0.65f;
    float cx = camWidth / 2.0f;
    float cy = camHeight / 2.0f;
    
    const float REAL_HEAD_WIDTH = 100.0f;  // mm
    
    float depth_mm = (REAL_HEAD_WIDTH * focal_length) / head_rect.width;
    float z = depth_mm / 1000.0f;  // meters
    
    float x = (cx - (float)head_center.x) * z / focal_length;
    float y = (cy - (float)head_center.y) * z / focal_length;
    
    cv::Point3f npos(x, y, z);
    
    // different smoothing for Z
    float z_smoothing = 0.88f;  // much heavier smoothing for depth
    smoothed_head.x = smoothed_head.x * smoothing + npos.x * (1.0f - smoothing);
    smoothed_head.y = smoothed_head.y * smoothing + npos.y * (1.0f - smoothing);
    smoothed_head.z = smoothed_head.z * z_smoothing + npos.z * (1.0f - z_smoothing);

    if (fabs(smoothed_head.z - previous_z) < 5e-4) {
        smoothed_head.z = previous_z;
    }
    previous_z = smoothed_head.z;
    
    return smoothed_head;
}

    bool is_head_tracking_enabled() {
        return ht_tracking_enabled;
    }

    void destroy() {
        if (detector) {
            detector->destroy();
            delete detector;
            detector = nullptr;
        }
        delete camFrameTexture;
        delete capture;
    }
}

#endif // HEAD_TRACKER_HPP