#ifndef HEAD_TRACKER_HPP
#define HEAD_TRACKER_HPP

#include <algorithm>

#include <opencv2/objdetect/objdetect.hpp>
#include <opencv2/highgui/highgui.hpp>
#include <opencv2/imgproc/imgproc.hpp>
#include <opencv2/calib3d/calib3d.hpp>

#include "imgui.h"

namespace ht {

const cv::String face_cascade_name = "./assets/haarcascade_frontalface_alt.xml";
cv::CascadeClassifier face_cascade;
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
        if(!face_cascade.load(face_cascade_name)) {
            fprintf(stderr, "ERROR:HEAD_TRACKER: loading 'haarcascade_frontalface_alt.xml'\n");
            return false;
        };

        // start video capture from camera
        capture = new cv::VideoCapture(0);

        // lower resolution
        capture->set(cv::CAP_PROP_FRAME_WIDTH, 320);
        capture->set(cv::CAP_PROP_FRAME_HEIGHT, 240);

        // check that video is working
        if (capture == NULL || !capture->isOpened()) {
            fprintf(stderr, "ERROR:HEAD_TRACKER: Could not start video capture\n");
            return false;
        }

        camWidth = 320; //(int) capture->get(cv::CAP_PROP_FRAME_WIDTH);
        camHeight = 240; // (int) capture->get(cv::CAP_PROP_FRAME_HEIGHT);
        printf("camWidth: %d, camHeight: %d\n", camWidth, camHeight);

        camFrameTexture = new Texture2D(camWidth, camHeight, nullptr);
        camFrameTexture->generate_texture();
        return true;
    }

    void draw_cam_frame() {
        (*capture) >> frame;
        if (frame.empty()) return;

        cv::Mat gray;
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
        cv::equalizeHist(gray, gray);

        // detect faces
        std::vector<cv::Rect> faces;
        face_cascade.detectMultiScale(gray, faces, 1.1, 3, 0, cv::Size(50, 50));

        if (!faces.empty()) {
            // pick largest face
            head_rect = faces[0];
            for (size_t i = 1; i < faces.size(); i++) {
                if (faces[i].area() > head_rect.area()) {
                    head_rect = faces[i];
                }
            }

            // compute head center
            head_center = cv::Point(head_rect.x + head_rect.width / 2,  head_rect.y + head_rect.height / 2);

            // draw rectangle and center
            cv::rectangle(frame, head_rect, cv::Scalar(0, 255, 0), 2);
            cv::circle(frame, head_center, 5, cv::Scalar(255, 0, 0), -1);
        }

        // convert to RGB for texture
        cv::cvtColor(frame, frame, cv::COLOR_BGR2RGB);
        camFrameTexture->update_texture(frame.data);

        // draw ImGui mirrored image
        ImTextureID textureID = (ImTextureID)(intptr_t)((unsigned int)(*camFrameTexture));
        ImGui::Begin("Head Tracker");
        ImGui::Checkbox("Enable Head Tracking", &ht_tracking_enabled);
        if (ht_tracking_enabled) {
            ImVec2 size = ImGui::GetContentRegionAvail();
            ImGui::Image(textureID, size, ImVec2(1, 0), ImVec2(0, 1)); // mirror horizontally
        }
        ImGui::End();
    }

    cv::Point3f get_head_center() {
        if (head_rect.width == 0 || head_rect.height == 0)
            return cv::Point3f(0, 0, 0);
        
        // camera parameters
        // TODO: make it calibrated.
        float focal_length = camWidth * 0.65f;
        float cx = camWidth / 2.0f;
        float cy = camHeight / 2.0f;
        
        const float REAL_HEAD_WIDTH = 100.0f;  // mm
        
        float depth_mm = (REAL_HEAD_WIDTH * focal_length) / head_rect.width;
        
        // Convert to meters or your preferred unit
        float z = depth_mm / 1000.0f;  // meters
        
        // calculate real-world X and Y
        // X = (pixel_x - cx) * Z / focal_length
        // Y = (pixel_y - cy) * Z / focal_length
        float x = (cx - (float)head_center.x) * z / focal_length;  // mirrored
        float y = (cy - (float)head_center.y) * z / focal_length;
        
        // normalized point
        cv::Point3f npos(x, y, z);
        
        // smooth movement
        smoothed_head.x = smoothed_head.x * smoothing + npos.x * (1.0f - smoothing);
        smoothed_head.y = smoothed_head.y * smoothing + npos.y * (1.0f - smoothing);
        smoothed_head.z = smoothed_head.z * smoothing + npos.z * (1.0f - smoothing);
        
        return smoothed_head;
    }

    bool is_head_tracking_enabled() {
        return ht_tracking_enabled;
    }

    void destroy() {
        delete camFrameTexture;
        delete capture;
    }
}


#endif // HEAD_TRACKER_HPP