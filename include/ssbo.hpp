#ifndef SSBO_HPP
#define SSBO_HPP

#include <vector>
#include <iostream>

#define SSBO_DYNAMIC_DRAW 0x88E8
#define SSBO_READ_ONLY 0x88B8
#define SSBO_WRITE_ONLY 0x88B9
#define SSBO_READ_WRITE 0x88BA
#define SSBO_BUFFER_ACCESS 0x88BB
#define SSBO_SHADER_STORAGE_BARRIER_BIT 0x00002000

class SSBO {
private:
    unsigned int bufferID;
    unsigned int bindingPoint;
    size_t bufferSize;
    bool isMapped;
    void* mappedPointer;

public:
    SSBO(unsigned int binding = 0);
    ~SSBO();

    // Initialize buffer with size
    void initialize(size_t size, unsigned int usage = SSBO_DYNAMIC_DRAW);

    // Upload data to buffer
    template<typename T>
    void uploadData(const std::vector<T>& data);

    // Download data from buffer
    template<typename T>
    std::vector<T> downloadData();

    // map buffer for persistent access
    void* map(unsigned int access = SSBO_READ_WRITE);

    // unmap buffer
    void unmap();

    // Bind to specific binding point
    void bind(unsigned int binding);

    // Synchronization
    void waitForComputeShader();

    // Get buffer info
    unsigned int getID() const;
    size_t getSize() const;
    unsigned int getBindingPoint() const;
};

// Template method declarations (implementations in .cpp)
template<typename T>
void uploadData(const std::vector<T>& data);

template<typename T>
std::vector<T> downloadData();

#endif // SSBO_HPP