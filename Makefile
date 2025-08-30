all: runtime

runtime:
	g++ -std=c++17 main.cpp src/*.cpp src/*.cc extern/glad/src/glad.c -o sdf -I"include" -I"extern/glad/include" -I"extern/glm/include" -I"extern/OBJ_Loader/include" -I"extern/tinygltfloader/include" -I"extern/stb/include" -lglfw -lm