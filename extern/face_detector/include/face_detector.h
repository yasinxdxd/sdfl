/*
 * Simple Face Detection Library
 * Wrapper around MediaPipe TensorFlow Lite face detector
 */

#ifndef FACE_DETECTOR_H
#define FACE_DETECTOR_H

#include <opencv2/opencv.hpp>
#include <vector>

namespace FaceDetector {

// Structure to hold detection results
struct FaceRect {
    float x;      // Normalized x position (0-1)
    float y;      // Normalized y position (0-1)
    float width;  // Normalized width (0-1)
    float height; // Normalized height (0-1)
    float score;  // Confidence score (0-1)
    
    // Convert to pixel coordinates
    cv::Rect toPixelRect(int img_width, int img_height) const {
        return cv::Rect(
            static_cast<int>(x * img_width),
            static_cast<int>(y * img_height),
            static_cast<int>(width * img_width),
            static_cast<int>(height * img_height)
        );
    }
};

// Opaque handle to detector internals
class DetectorImpl;

class Detector {
public:
    /**
     * Initialize the face detector
     * @param model_path Path to the .tflite model file
     * @param confidence_threshold Minimum confidence score for detections (default: 0.75)
     * @param nms_threshold Non-maximum suppression threshold (default: 0.3)
     * @return true if initialization successful, false otherwise
     */
    bool init(const char* model_path, 
              float confidence_threshold = 0.75f,
              float nms_threshold = 0.3f);
    
    /**
     * Detect faces in an image
     * @param frame Input image (BGR format from OpenCV)
     * @param detections Output vector of detected faces
     * @param mirror Apply horizontal flip for mirror effect (default: true)
     * @return true if detection successful, false otherwise
     */
    bool detect(const cv::Mat& frame, 
                std::vector<FaceRect>& detections,
                bool mirror = true);
    
    /**
     * Draw detection rectangles on an image
     * @param frame Image to draw on (will be modified)
     * @param detections Face detections to draw
     * @param color Rectangle color (default: blue)
     * @param thickness Line thickness (default: 2)
     */
    void drawDetections(cv::Mat& frame,
                       const std::vector<FaceRect>& detections,
                       const cv::Scalar& color = cv::Scalar(255, 0, 0),
                       int thickness = 2);
    
    /**
     * Clean up and release resources
     */
    void destroy();
    
    /**
     * Check if detector is initialized
     */
    bool isInitialized() const;
    
    // Constructor/Destructor
    Detector();
    ~Detector();
    
    // Disable copy
    Detector(const Detector&) = delete;
    Detector& operator=(const Detector&) = delete;

private:
    DetectorImpl* impl_;
};

} // namespace FaceDetector

#endif // FACE_DETECTOR_H