#ifndef SDF_CLIENT_HPP
#define SDF_CLIENT_HPP

#include <httplib.h>
#include <iostream>
#include <vector>
#include <string>
#include <chrono>

struct ProgramMetaData {
    uint64_t program_id;
    std::string name;
    std::string created_at;
    std::vector<uint8_t> preview_image;
    std::vector<std::string> tags;
};

void publish_program(const char* name,
                    const char* description,
                    const std::string& code,
                    const std::string& sequence_representation,
                    const std::vector<uint8_t>& preview_image,
                    const std::vector<std::string>& tags)
{
    httplib::Client cli("http://localhost:9999");

    // unique per request
    auto timestamp = std::chrono::system_clock::now().time_since_epoch().count();
    std::string boundary = "----SDFClientBoundary_" + std::to_string(timestamp);
    std::string body;

    auto add_field = [&](const std::string& field_name, const std::string& value) {
        body += "--" + boundary + "\r\n";
        body += "Content-Disposition: form-data; name=\"" + field_name + "\"\r\n\r\n";
        body += value + "\r\n";
    };

    // text fields
    add_field("name", name);
    add_field("description", description);

    add_field("code", code);
    add_field("sequence_representation", sequence_representation);

    // add tags (multiple fields with same name "tags")
    for (const auto& tag : tags) {
        add_field("tags", tag);
    }

    // preview_image field
    body += "--" + boundary + "\r\n";
    body += "Content-Disposition: form-data; name=\"preview_image\"; filename=\"preview.jpg\"\r\n";
    body += "Content-Type: application/octet-stream\r\n\r\n";

    // append binary data
    std::string body_with_image = body;
    body_with_image.append(reinterpret_cast<const char*>(preview_image.data()), preview_image.size());
    body_with_image += "\r\n";

    // end boundary
    body_with_image += "--" + boundary + "--\r\n";

    std::string content_type = "multipart/form-data; boundary=" + boundary;

    // send request
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

void update_cache() {
    httplib::Client cli("http://localhost:9999");
    auto res = cli.Get("/programs");

    if (res) {
        std::cout << "Status: " << res->status << "\n";
        std::cout << "Body: " << res->body << "\n";
    } else {
        std::cerr << "Request failed\n";
    }
}

std::vector<ProgramMetaData> get_programs_from_cache(const std::string& cache_path = "programs_cache.bin") {
    std::vector<ProgramMetaData> programs;

    std::ifstream file(cache_path, std::ios::binary);
    if (!file.is_open()) {
        std::cerr << "Failed to open cache file, updating cache..." << cache_path << "\n";
    }
    while (!file.is_open()) {
        update_cache();
        file.open(cache_path, std::ios::binary);
    }

    auto read_uint64 = [&](uint64_t& val) {
        file.read(reinterpret_cast<char*>(&val), sizeof(uint64_t));
        return !file.fail();
    };

    auto read_string = [&](std::string& s) {
        uint64_t len;
        if (!read_uint64(len)) return false;
        if (len > (1ull << 30)) return false; // sanity check: avoid absurd sizes
        s.resize(len);
        if (len > 0) {
            file.read(&s[0], len);
        }
        return !file.fail();
    };

    auto read_bytes = [&](std::vector<uint8_t>& v) {
        uint64_t len;
        if (!read_uint64(len)) return false;
        if (len > (1ull << 30)) return false; // sanity check
        v.resize(len);
        if (len > 0) {
            file.read(reinterpret_cast<char*>(v.data()), len);
        }
        return !file.fail();
    };

    uint64_t program_count;
    if (!read_uint64(program_count)) return programs;

    for (uint64_t i = 0; i < program_count; ++i) {
        ProgramMetaData p;

        if (!read_uint64(p.program_id)) break;
        if (!read_string(p.name)) break;
        if (!read_string(p.created_at)) break;
        if (!read_bytes(p.preview_image)) break;

        uint64_t tag_count;
        if (!read_uint64(tag_count)) break;
        p.tags.resize(tag_count);
        for (uint64_t t = 0; t < tag_count; ++t) {
            if (!read_string(p.tags[t])) break;
        }

        programs.push_back(std::move(p));
    }

    return programs;
}

void generate_random_program() {
    httplib::Client cli("http://localhost:9999");
    auto res = cli.Get("/generate_random");

    if (res && res->status == 200) {
        std::cout << "Status: " << res->status << "\n";
    } else {
        std::cerr << "Request failed: " << res->status <<"\n";
    }
}


void shutdown_server() {
    httplib::Client cli("http://localhost:9999");
    auto res = cli.Get("/shutdown");
    
    if (res && res->status == 200) {
        std::cout << "Server shutdown successfully\n";
    } else {
        std::cerr << "Failed to shutdown server\n";
    }
}

#endif // SDF_CLIENT_HPP
