#include "ssbo.hpp"
#include <glad/glad.h>
#include <iostream>
#include <cstring>

SSBO::SSBO(GLuint binding) 
    : bufferID(0), bindingPoint(binding), bufferSize(0), isMapped(false), mappedPointer(nullptr) {
    glGenBuffers(1, &bufferID);
}

SSBO::~SSBO() {
    if (isMapped) {
        unmap();
    }
    if (bufferID != 0) {
        glDeleteBuffers(1, &bufferID);
    }
}

void SSBO::initialize(size_t size, GLenum usage) {
    bufferSize = size;
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferID);
    glBufferData(GL_SHADER_STORAGE_BUFFER, size, nullptr, usage);
    glBindBufferBase(GL_SHADER_STORAGE_BUFFER, bindingPoint, bufferID);
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, 0);
}

void* SSBO::map(GLenum access) {
    if (!isMapped) {
        glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferID);
        mappedPointer = glMapBuffer(GL_SHADER_STORAGE_BUFFER, access);
        isMapped = (mappedPointer != nullptr);
        glBindBuffer(GL_SHADER_STORAGE_BUFFER, 0);
        
        if (!isMapped) {
            std::cerr << "Failed to map SSBO for access" << std::endl;
        }
    }
    return mappedPointer;
}

void SSBO::unmap() {
    if (isMapped) {
        glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferID);
        GLboolean success = glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);
        glBindBuffer(GL_SHADER_STORAGE_BUFFER, 0);
        
        if (!success) {
            std::cerr << "Warning: SSBO unmap indicated data corruption" << std::endl;
        }
        
        isMapped = false;
        mappedPointer = nullptr;
    }
}

void SSBO::bind(GLuint binding) {
    bindingPoint = binding;
    glBindBufferBase(GL_SHADER_STORAGE_BUFFER, bindingPoint, bufferID);
}

GLuint SSBO::getID() const {
    return bufferID;
}

size_t SSBO::getSize() const {
    return bufferSize;
}

GLuint SSBO::getBindingPoint() const {
    return bindingPoint;
}

void SSBO::waitForComputeShader() {
    glMemoryBarrier(GL_SHADER_STORAGE_BARRIER_BIT);
}

template<typename T>
void SSBO::uploadData(const std::vector<T>& data) {
    bufferSize = data.size() * sizeof(T);
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferID);
    glBufferData(GL_SHADER_STORAGE_BUFFER, bufferSize, data.data(), GL_DYNAMIC_DRAW);
    glBindBufferBase(GL_SHADER_STORAGE_BUFFER, bindingPoint, bufferID);
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, 0);
}

template<typename T>
std::vector<T> SSBO::downloadData() {
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferID);
    
    // ensure GPU writes are complete
    glMemoryBarrier(GL_SHADER_STORAGE_BARRIER_BIT);
    
    size_t elementCount = bufferSize / sizeof(T);
    std::vector<T> data(elementCount);
    
    void* ptr = glMapBuffer(GL_SHADER_STORAGE_BUFFER, GL_READ_ONLY);
    if (ptr) {
        memcpy(data.data(), ptr, bufferSize);
        glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);
    }
    
    glBindBuffer(GL_SHADER_STORAGE_BUFFER, 0);
    return data;
}

template void SSBO::uploadData<float>(const std::vector<float>& data);
template void SSBO::uploadData<int>(const std::vector<int>& data);
template void SSBO::uploadData<double>(const std::vector<double>& data);
template void SSBO::uploadData<unsigned int>(const std::vector<unsigned int>& data);

template std::vector<float> SSBO::downloadData<float>();
template std::vector<int> SSBO::downloadData<int>();
template std::vector<double> SSBO::downloadData<double>();
template std::vector<unsigned int> SSBO::downloadData<unsigned int>();
