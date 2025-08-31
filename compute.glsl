#version 430

layout(local_size_x = 8, local_size_y = 8, local_size_z = 8) in;

// SSBO
layout(std430, binding = 0) buffer SDFBuffer {
    float sdfData[];
};

uniform vec3 minBound;
uniform vec3 maxBound;
uniform int resolution;

float sdSphere(vec3 p, float r) {
    return length(p) - r;
}

void main() {
    ivec3 gid = ivec3(gl_GlobalInvocationID);
    if (any(greaterThanEqual(gid, ivec3(resolution)))) return;

    vec3 uv = vec3(gid) / float(resolution - 1);
    vec3 p = mix(minBound, maxBound, uv);

    float d = sdSphere(p, 0.3);

    // convert 3D index to 1D
    int index = gid.z * resolution * resolution + gid.y * resolution + gid.x;
    sdfData[index] = d;
}