#ifndef SDF_CLIENT_HPP
#define SDF_CLIENT_HPP

#include <httplib.h>
#include <iostream>
#include <vector>
#include <string>
#include <chrono>

void create_program(const char* name, const char* description, const std::vector<uint8_t>& preview_image) {
    httplib::Client cli("http://localhost:9999");

    // avoid collisions if multiple requests happen simultaneously
    auto timestamp = std::chrono::system_clock::now().time_since_epoch().count();
    std::string boundary = "----SDFClientBoundary_" + std::to_string(timestamp);
    std::string body;

    // name field
    body += "--" + boundary + "\r\n";
    body += "Content-Disposition: form-data; name=\"name\"\r\n\r\n";
    body += name;
    body += "\r\n";

    // description field
    body += "--" + boundary + "\r\n";
    body += "Content-Disposition: form-data; name=\"description\"\r\n\r\n";
    body += description;
    body += "\r\n";

    // preview_image field
    body += "--" + boundary + "\r\n";
    body += "Content-Disposition: form-data; name=\"preview_image\"; filename=\"preview.jpg\"\r\n";
    body += "Content-Type: application/octet-stream\r\n\r\n";

    // reserve space and append raw bytes safely
    std::string body_with_image = body;
    body_with_image.append(reinterpret_cast<const char*>(preview_image.data()), preview_image.size());
    body_with_image += "\r\n";

    // end boundary
    body_with_image += "--" + boundary + "--\r\n";

    std::string content_type = "multipart/form-data; boundary=" + boundary;

    // send using raw pointer + size to avoid truncation
    auto res = cli.Post("/program",
                        body_with_image.data(),
                        body_with_image.size(),
                        content_type.c_str());

    if (res) {
        std::cout << "Status: " << res->status << "\n";
        std::cout << "Body: " << res->body << "\n";
    } else {
        std::cerr << "Request failed\n";
    }
}

#endif // SDF_CLIENT_HPP
