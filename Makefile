all: runtime

runtime:
	g++ -std=c++17 -fsanitize=address -g main.cpp src/*.cpp src/*.cc extern/imgui/src/*.cpp extern/glad/src/glad.c -o sdf -I"include" -I"extern/glad/include" -I"extern/glm/include" -I"extern/OBJ_Loader/include" -I"extern/imgui/include" -I"extern/tinygltfloader/include" -I"extern/stb/include" -I"extern/httplib" `pkg-config --cflags --libs opencv4` -lglfw -lm -lpthread