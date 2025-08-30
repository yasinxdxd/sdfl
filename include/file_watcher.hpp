#ifndef FILE_WATCHER_HPP
#define FILE_WATCHER_HPP

#include <iostream>
#include <string>
#include <stdexcept>

#ifdef _WIN32
#include <windows.h>
#else
#include <sys/stat.h>
#include <unistd.h>
#endif

class FileWatcher {
public:
    explicit FileWatcher(const std::string& path) : filepath(path) {
        lastWriteTime = getLastWriteTime(filepath);
    }

    bool hasChanged() {
        auto current = getLastWriteTime(filepath);
        if (current != lastWriteTime) {
            lastWriteTime = current;
            return true;
        }
        return false;
    }

private:
    std::string filepath;

#ifdef _WIN32
    using FileTime = FILETIME;
#else
    using FileTime = time_t;
#endif

    FileTime lastWriteTime;

    static FileTime getLastWriteTime(const std::string& path) {
#ifdef _WIN32
        WIN32_FILE_ATTRIBUTE_DATA fad;
        if (!GetFileAttributesExA(path.c_str(), GetFileExInfoStandard, &fad)) {
            throw std::runtime_error("Failed to get file attributes: " + path);
        }
        return fad.ftLastWriteTime;
#else
        struct stat attr;
        if (stat(path.c_str(), &attr) != 0) {
            throw std::runtime_error("Failed to stat file: " + path);
        }
        return attr.st_mtime;
#endif
    }
};

#endif // FILE_WATCHER_HPP