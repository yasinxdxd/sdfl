package sdfl

import "fmt"

var functionSymbols = map[string]FunDef{
	"scene":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_SCENE, Id: "scene", FunDefArgNames: []string{"background", "camera", "children"}},
	"local":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_LOCAL, Id: "local", FunDefArgNames: []string{"children"}},
	"camera":             {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_CAMERA, Id: "camera", FunDefArgNames: []string{"position"}},
	"plane":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "plane", FunDefArgNames: []string{"height"}},
	"sphere":             {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "sphere", FunDefArgNames: []string{"position", "radius"}},
	"ellipsoid":          {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "ellipsoid", FunDefArgNames: []string{"position", "radius"}},
	"box":                {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "box", FunDefArgNames: []string{"position", "size"}},
	"torus":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "torus", FunDefArgNames: []string{"position", "radius", "thickness"}},
	"rotateAround":       {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_ROTATE_AROUND, Id: "rotateAround", FunDefArgNames: []string{"position", "rotation", "child"}},
	"smoothUnion":        {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothUnion", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
	"smoothSubtraction":  {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothSubtraction", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
	"smoothIntersection": {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothIntersection", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
	"union":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "union", FunDefArgNames: []string{"child1", "child2"}},
	"subtraction":        {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "subtraction", FunDefArgNames: []string{"child1", "child2"}},
	"intersection":       {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "intersection", FunDefArgNames: []string{"child1", "child2"}},
}

var generatedCodeFragmentShader = ""
var generatedCodeComputeShader = ""

func Reset() {
	generatedCodeFragmentShader = ""
	generatedCodeComputeShader = ""
}

func GetFragmentCode() string {
	return generatedCodeFragmentShader
}

func GetComputeCode() string {
	return generatedCodeComputeShader
}

type generator interface {
	generate(args ...any)
}

func generateFragmentCode(code string, args ...any) {
	generatedCodeFragmentShader += fmt.Sprintf(code, args...)
}

func generateComputeCode(code string, args ...any) {
	generatedCodeComputeShader += fmt.Sprintf(code, args...)
}

func generateCodeBoth(code string, args ...any) {
	generateFragmentCode(code, args...)
	generateComputeCode(code, args...)
}

func Generate(prog *Program) {
	prog.generate()
}

func (prog *Program) generate(args ...any) {
	generateGlslFragmentHeader()
	generateGlslFragmentGetMaterial()
	generateGlslComputeHeader()
	generateGlslBuiltinSDFFunctions()

	sceneCall := prog.Expr.FunCall
	if sceneCall == nil || sceneCall.Id != "scene" {
		fmt.Println("ERROR: scene function must be called")
		return
	}
	if _, ok := sceneCall.FunNamedArgs["camera"]; !ok {
		fmt.Println("ERROR: scene function had argument camera")
		return
	}
	cameraCall := sceneCall.FunNamedArgs["camera"].Expr.FunCall
	if cameraCall == nil {
		fmt.Println("ERROR: scene, camera argument is empty")
		return
	}
	if _, ok := cameraCall.FunNamedArgs["position"]; !ok {
		fmt.Println("ERROR: camera function had argument position")
		return
	}
	cameraPos := cameraCall.FunNamedArgs["position"].Expr.Tuple
	if cameraPos == nil {
		fmt.Println("ERROR: camera, position argument is empty")
		return
	}
	if _, ok := sceneCall.FunNamedArgs["children"]; !ok {
		fmt.Println("ERROR: scene function had argument children")
		return
	}
	childrenArr := sceneCall.FunNamedArgs["children"].Expr.ArrExpr
	if childrenArr == nil {
		fmt.Println("ERROR: scene, children argument is empty")
		return
	}

	backgroundStr := "vec3(0, 0, 0)"
	if _, ok := sceneCall.FunNamedArgs["background"]; ok {
		if sceneCall.FunNamedArgs["background"].Expr.Tuple != nil {
			r := sceneCall.FunNamedArgs["background"].Expr.Tuple.Values[0]
			g := sceneCall.FunNamedArgs["background"].Expr.Tuple.Values[1]
			b := sceneCall.FunNamedArgs["background"].Expr.Tuple.Values[2]
			backgroundStr = fmt.Sprintf("vec3(%s, %s, %s)", r, g, b)
		} else {
			fmt.Println("ERROR: scene function had argument background as tuple")
			return
		}
	}

	generateGlslPushScene()
	for _, stmt := range prog.Stmts {
		stmt.generate()
	}

	generateGlslDistSceneBegin()
	for _, expr := range childrenArr.Exprs {
		expr.generate()
	}
	generateGlslDistSceneEnd()

	generateGlslRaymarchEngine()

	generateGlslFragmentMain(cameraCall, backgroundStr)
	generateGlslComputeMain()
}

func (stmt *Stmt) generate(args ...any) {
	switch stmt.Type {
	case AST_FUN_DEF:
		stmt.FunDef.generate()
	default:
		fmt.Printf("gen error: unknown stmt type: %v\n", stmt.Type)
	}
}

func (funDef *FunDef) generate(args ...any) {
	// TODO: get arguments
	generateCodeBoth("\nfloat %s(%s) {\n", funDef.Id, "vec3 p")
	generateCodeBoth("    SceneResult d = SceneResult(SDFL_MAX_DISTANCE, 0);\n")
	localCall := funDef.Expr.FunCall
	if localCall == nil || localCall.Id != "local" {
		fmt.Println("ERROR: local function must be called in a function definition")
		return
	}
	if _, ok := localCall.FunNamedArgs["children"]; !ok {
		fmt.Println("ERROR: local function had argument children")
		return
	}
	childrenArr := localCall.FunNamedArgs["children"].Expr.ArrExpr
	if childrenArr == nil {
		fmt.Println("ERROR: local, children argument is empty")
		return
	}

	for _, expr := range childrenArr.Exprs {
		expr.generate()
	}

	generateCodeBoth("    return d.distance;\n")
	generateCodeBoth("}\n")
}

func (expr *Expr) generate(args ...any) {
	switch expr.Type {
	case AST_FUN_CALL:
		expr.FunCall.generate()
	case AST_TUPLE1:
		fallthrough
	case AST_TUPLE2:
		fallthrough
	case AST_TUPLE3:
		expr.Tuple.generate()
	case AST_ARR_EXPR:
		expr.ArrExpr.generate()
	case AST_NUMBER:
		expr.Number.generate()
	default:
		fmt.Printf("gen error: unknown expr type: %v\n", expr.Type)
	}
}

var varCounters = make(map[string]int)

func freshVar(base string) string {
	count := varCounters[base]
	name := fmt.Sprintf("%s%d", base, count)
	varCounters[base] = count + 1
	return name
}

func (funCall *FunCall) generate(args ...any) string {
	rayPosition := "p"
	parentIsOp := false
	if len(args) > 0 {
		rayPosition = args[0].(string)
		if len(args) > 1 {
			parentIsOp = args[1].(bool)
		}
	}

	funDef, ok := functionSymbols[funCall.Id]
	println(funCall.Id, funDef.SymbolType, rayPosition)

	if !ok {
		return ""
	}

	orderedArgs := func() []*Expr {
		exprs := []*Expr{}
		// figure out named parameter order
		for j := 0; j < len(funDef.FunDefArgNames); j++ {
			funNamedArg, ok := funCall.FunNamedArgs[funDef.FunDefArgNames[j]]
			if !ok {
				// TODO: Better error messages
				fmt.Println("ERROR!")
			}
			if funDef.FunDefArgNames[j] == funNamedArg.ArgName {
				exprs = append(exprs, &funNamedArg.Expr)
			}
		}
		return exprs
	}

	genFunCall := func(funId string) string {
		return "sdfl_builtin_" + funId
	}

	switch funDef.SymbolType {
	case FUN_BUILTIN_ROTATE_AROUND:
		posExpr, okPos := funCall.FunNamedArgs["position"]
		rotExpr, okRot := funCall.FunNamedArgs["rotation"]
		childExpr, okChild := funCall.FunNamedArgs["child"]

		if !okPos || !okRot || !okChild {
			fmt.Println("ERROR: rotateAround missing args (needs position, rotation, child)")
			return ""
		}

		qVar := freshVar("q")

		// Generate the rotation transformation code
		generateCodeBoth(fmt.Sprintf("    vec3 %s = %s - ", qVar, rayPosition))
		posExpr.Expr.generate()
		generateCodeBoth(";\n")

		generateCodeBoth(fmt.Sprintf("    %s = sdfl_RotationMatrix(radians(", qVar))
		rotExpr.Expr.generate()
		generateCodeBoth(fmt.Sprintf(")) * %s;\n", qVar))

		generateCodeBoth(fmt.Sprintf("    %s += ", qVar))
		posExpr.Expr.generate()
		generateCodeBoth(";\n")

		// CRITICAL: Pass the new coordinate system (qVar) to the child
		// This ensures all nested shapes use the rotated coordinates
		return childExpr.Expr.FunCall.generate(qVar, parentIsOp)

	case FUN_BUILTIN_OP:
		exprs := orderedArgs()

		// Both children should use the SAME rayPosition (which might be "p" or "q16" etc)
		// and both should be marked as parentIsOp=true so they don't call sdfl_PushScene
		child1Var := exprs[0].FunCall.generate(rayPosition, true)
		child2Var := exprs[1].FunCall.generate(rayPosition, true)

		sd := freshVar("sd")
		// Use child1, child2 order to match the expected output
		// smoothUnion(child1: sphere, child2: rotateAround) -> smoothUnion(child1_var, child2_var)
		generateCodeBoth(fmt.Sprintf("    SceneResult %s = %s(%s, %s", sd, genFunCall(funDef.Id), child1Var, child2Var))

		// smooth_transition parameter
		if len(exprs) > 2 {
			generateCodeBoth(", ")
			exprs[2].generate()
		}
		generateCodeBoth(");\n")

		// Only push to scene if this isn't part of a larger operation
		if !parentIsOp {
			generateCodeBoth(fmt.Sprintf("    d = sdfl_PushScene(%s);\n", sd))
		}
		return sd

	case FUN_BUILTIN:
		exprs := orderedArgs()

		sd := freshVar("sd")
		generateCodeBoth(fmt.Sprintf("    SceneResult %s = SceneResult(%s(%s, ", sd, genFunCall(funDef.Id), rayPosition))
		for i, e := range exprs {
			e.generate()
			if i < len(exprs)-1 {
				generateCodeBoth(", ")
			}
		}
		generateCodeBoth("), 0);\n")

		if !parentIsOp {
			generateCodeBoth(fmt.Sprintf("    d = sdfl_PushScene(%s);\n", sd))
		}
		return sd

	case FUN_USER_DEFINED:
		exprs := orderedArgs()

		sd := freshVar("sd")
		generateCodeBoth(fmt.Sprintf("    SceneResult %s = SceneResult(%s(%s", sd, funDef.Id, rayPosition))
		for i, e := range exprs {
			if i < len(exprs)-1 {
				generateCodeBoth(", ")
			}
			e.generate()
		}
		generateCodeBoth("), 0);\n")

		if !parentIsOp {
			generateCodeBoth(fmt.Sprintf("    d = sdfl_PushScene(%s);\n", sd))
		}
		return sd

	default:
		fmt.Println("NOT IMPLEMENTED YET!")
		return ""
	}
}

func (tuple *Tuple) generate(args ...any) {
	var isCamera = false
	if len(args) > 0 {
		isCamera = args[0].(bool)
	}
	if isCamera {
		generateFragmentCode("vec3(%s, %s, %s)", tuple.Values[0], tuple.Values[1], tuple.Values[2])
	} else {
		generateCodeBoth("vec3(%s, %s, %s)", tuple.Values[0], tuple.Values[1], tuple.Values[2])
	}
}

func (number *Number) generate(args ...any) {
	generateCodeBoth(number.Value)
}

func (arrExpr *ArrExpr) generate(args ...any) {
	for _, expr := range arrExpr.Exprs {
		expr.generate()
	}
}

func generateGlslCamera(cameraFunCall *FunCall) {
	generateFragmentCode("    // generated camera position\n")
	generateFragmentCode("    vec3 ray_origin = ")
	cameraFunCall.FunNamedArgs["position"].Expr.Tuple.generate(true)
	generateFragmentCode(";")
	generateFragmentCode(`
    vec3 ray_dir = normalize(vec3(uv, -1)); // ray direction for the each pixel
`)
}

func generateGlslFragmentMain(cameraFunCall *FunCall, bg string) {
	generateFragmentCode(`
void main() {
    vec2 uv = o_vertex_uv * 2. - 1.;
    uv.y *= float(window_size.y) / float(window_size.x);
`)

	generateGlslCamera(cameraFunCall)

	generateFragmentCode(`
    SceneResult result = sdfl_RayMarch(ray_origin, ray_dir);
    
    vec3 color = vec3(0.0);
    
    if (result.distance < SDFL_MAX_DISTANCE) {
        vec3 p = ray_origin + ray_dir * result.distance;
        vec3 view_dir = -ray_dir;
        
        Material mat = sdfl_GetMaterial(result.materialId);
        color = sdfl_CalculateLighting(p, view_dir, mat);
    } else {
        // Background/sky
        color = mix(vec3(0.5, 0.7, 1.0), %s, uv.y * 0.5 + 0.5);
    }
    
    frag_color = vec4(color, 1.0);
}
`, bg)
}

func generateGlslComputeMain() {
	generateComputeCode(`
void main() {
    ivec3 gid = ivec3(gl_GlobalInvocationID);
    if (any(greaterThanEqual(gid, ivec3(resolution)))) return;

    vec3 uv = vec3(gid) / float(resolution - 1);
    vec3 p = mix(minBound, maxBound, uv);

    // float d = sdSphere(p, 0.3);
    float d = sdfl_GetDistScene(p).distance;

    // convert 3D index to 1D
    int index = gid.z * resolution * resolution + gid.y * resolution + gid.x;
    sdfData[index] = d;
}
`)
}

func generateGlslFragmentHeader() {
	generateFragmentCode(`
// sdfl generated code

#version 430 core

in vec3 o_vertex_color;
in vec2 o_vertex_uv;

out vec4 frag_color;


// uniforms
uniform ivec2 window_size;
uniform float elapsed_time;

#define SDFL_MAX_STEPS 100
#define SDFL_MAX_DISTANCE 100.
#define SDFL_HIT_DISTANCE .01
#define SDFL_SHADOW_CAST_DISTANCE .05

// material system
struct Material {
    vec3 albedo;
    float roughness;
    float metallic;
    vec3 emission;
};

struct SceneResult {
    float distance;
    int materialId;
};
`)
}

func generateGlslFragmentGetMaterial() {
	generateFragmentCode(`
Material sdfl_GetMaterial(int id) {
    Material mat;
    
    if (id == 0) { // Default/Ground
        mat.albedo = vec3(0.8, 0.8, 0.8);
        mat.roughness = 0.9;
        mat.metallic = 0.0;
        mat.emission = vec3(0.0);
    }
    else if (id == 1) { // Main object (torus-subtracted sphere)
        mat.albedo = vec3(0.2, 0.6, 1.0);
        mat.roughness = 0.3;
        mat.metallic = 0.1;
        mat.emission = vec3(0.0);
    }
    else if (id == 2) { // Eyes/spheres
        mat.albedo = vec3(1.0, 0.3, 0.2);
        mat.roughness = 0.1;
        mat.metallic = 0.0;
        mat.emission = vec3(0.1, 0.0, 0.0); // slight red glow
    }
    else { // Fallback
        mat.albedo = vec3(0.5, 0.5, 0.5);
        mat.roughness = 0.5;
        mat.metallic = 0.0;
        mat.emission = vec3(0.0);
    }
    
    return mat;
}
`)
}

func generateGlslComputeHeader() {
	generateComputeCode(`
// sdfl generated code

#version 430

layout(local_size_x = 8, local_size_y = 8, local_size_z = 8) in;

// SSBO
layout(std430, binding = 0) buffer SDFBuffer {
    float sdfData[];
};

uniform vec3 minBound;
uniform vec3 maxBound;
uniform int resolution;

#define SDFL_MAX_DISTANCE 100.

struct SceneResult {
    float distance;
    int materialId;
};
`)
}

func generateGlslBuiltinSDFFunctions() {
	code := `
float sdfl_builtin_plane(vec3 p, float height) {    
    return p.y - height;
}

float sdfl_builtin_sphere(vec3 p, vec3 pos, float r) {    
    return distance(pos, p) - r;
}

float sdfl_builtin_ellipsoid(vec3 p, vec3 pos, vec3 r) {
    vec3 q = (p - pos) / r;
    return (length(q) - 1.0) * min(min(r.x, r.y), r.z);
}

float sdfl_builtin_box(vec3 p, vec3 bpos, vec3 bsize) {
    // Shift point into the box's local coordinate system
    vec3 q = abs(p - bpos) - bsize;
    // Outside distance + inside distance
    return length(max(q, 0.0)) + min(max(q.x, max(q.y, q.z)), 0.0);
}

float sdfl_builtin_torus(vec3 p, vec3 pos, float radius, float thickness) {
	vec3 wp = p - pos;
	vec2 t = vec2(radius, thickness);
    vec2 q = vec2(length(wp.xz)-t.x,wp.y);
    return length(q)-t.y;
}

mat3 sdfl_RotationMatrix(vec3 angles) {
    // angles = (rx, ry, rz) in radians
    float cx = cos(angles.x), sx = sin(angles.x);
    float cy = cos(angles.y), sy = sin(angles.y);
    float cz = cos(angles.z), sz = sin(angles.z);

    // compose rotation: Rz * Ry * Rx
    return mat3(
        cy*cz, cz*sx*sy - cx*sz, sx*sz + cx*cz*sy,
        cy*sz, cx*cz + sx*sy*sz, cx*sy*sz - cz*sx,
        -sy,   cy*sx,            cx*cy
    );
}

SceneResult sdfl_builtin_union(SceneResult d1, SceneResult d2) {
    if (d1.distance < d2.distance) {
        return SceneResult(d1.distance, d1.materialId);
    } else {
        return SceneResult(d2.distance, d2.materialId);
    }
}

SceneResult sdfl_builtin_subtraction(SceneResult d1, SceneResult d2) {
    float dist = max(d2.distance, -d1.distance);
    return SceneResult(dist, d2.materialId);
}

SceneResult sdfl_builtin_intersection(SceneResult d1, SceneResult d2) {
    if (d1.distance > d2.distance) {
        return SceneResult(d1.distance, d1.materialId);
    } else {
        return SceneResult(d2.distance, d2.materialId);
    }
}

// https://iquilezles.org/articles/distfunctions/

SceneResult sdfl_builtin_smoothUnion(SceneResult d1, SceneResult d2, float k) {
    float h = clamp(0.5 + 0.5*(d2.distance-d1.distance)/k, 0.0, 1.0);
    float dist = mix(d2.distance, d1.distance, h) - k*h*(1.0-h);
    int matId = (h > 0.5) ? d1.materialId : d2.materialId;
    return SceneResult(dist, matId);
}

SceneResult sdfl_builtin_smoothSubtraction(SceneResult d1, SceneResult d2, float k) {
    float h = clamp(0.5 - 0.5*(d2.distance+d1.distance)/k, 0.0, 1.0);
    float dist = mix(d2.distance, -d1.distance, h) + k*h*(1.0-h);
    // For subtraction, keep the material of the object being subtracted from
    return SceneResult(dist, d2.materialId);
}

SceneResult sdfl_builtin_smoothIntersection(SceneResult d1, SceneResult d2, float k) {
    float h = clamp(0.5 - 0.5*(d2.distance-d1.distance)/k, 0.0, 1.0);
    float dist = mix(d2.distance, d1.distance, h) + k*h*(1.0-h);
    int matId = (h > 0.5) ? d1.materialId : d2.materialId;
    return SceneResult(dist, matId);
}
`
	generateFragmentCode(code)
	generateComputeCode(code)
}

func generateGlslRaymarchEngine() {
	generateFragmentCode(`
SceneResult sdfl_RayMarch(vec3 ray_origin, vec3 ray_dir) {
    float dfo = 0.;
    SceneResult result = SceneResult(SDFL_MAX_DISTANCE, 0);

    for (int i = 0; i < SDFL_MAX_STEPS; i++) {
        vec3 p = ray_origin + ray_dir * dfo;
        
        SceneResult scene = sdfl_GetDistScene(p);
        dfo += scene.distance;
        result.materialId = scene.materialId;

        if (dfo > SDFL_MAX_DISTANCE || scene.distance < SDFL_HIT_DISTANCE) {
            break;
        }
    }
    
    result.distance = dfo;
    return result;
}

vec3 sdfl_GetNormal(vec3 p) {
    float d = sdfl_GetDistScene(p).distance;
    vec2 off = vec2(.01, 0.);

    vec3 normal = vec3(
        d - sdfl_GetDistScene(p - off.xyy).distance,
        d - sdfl_GetDistScene(p - off.yxy).distance,
        d - sdfl_GetDistScene(p - off.yyx).distance
    );

    return normalize(normal);
}

float sdfl_GetShadow(vec3 p, vec3 light_dir, float light_distance) {
    float shadow = 1.0;
    float penumbra_factor = 10.0; // higher values = sharper shadows
    
    vec3 start_pos = p + sdfl_GetNormal(p) * SDFL_SHADOW_CAST_DISTANCE;
    float t = 0.0;
    
    for(int i = 0; i < 32; ++i) {
        vec3 ray_pos = start_pos + light_dir * t;
        SceneResult result = sdfl_GetDistScene(ray_pos);
        
        if(result.distance + SDFL_SHADOW_CAST_DISTANCE < SDFL_SHADOW_CAST_DISTANCE) {
            return 0.1; // hard shadow
        }
        
        // calculate soft shadow contribution
        shadow = min(shadow, penumbra_factor * result.distance / t);
        
        t += result.distance;
        
        if(t >= light_distance) {
            break;
        }
    }
    
    return clamp(shadow, 0.1, 1.0);
}

vec3 sdfl_CalculateLighting(vec3 p, vec3 view_dir, Material mat) {
    vec3 light_pos = vec3(0, 8, 8);
    vec3 light_color = vec3(1.0, 0.95, 0.8);
    float light_intensity = 2.0;
    
    vec3 light_dir = normalize(light_pos - p);
    vec3 normal = sdfl_GetNormal(p);
    vec3 half_dir = normalize(light_dir + view_dir);
    
    float light_distance = distance(light_pos, p);
    float attenuation = 1.0 / (1.0 + 0.1 * light_distance + 0.01 * light_distance * light_distance);
    
    // Diffuse
    float ndotl = max(dot(normal, light_dir), 0.0);
    vec3 diffuse = mat.albedo * light_color * ndotl * light_intensity * attenuation;
    
    // Specular (simplified)
    float ndoth = max(dot(normal, half_dir), 0.0);
    float roughness2 = mat.roughness * mat.roughness;
    float spec_power = 2.0 / (roughness2 * roughness2) - 2.0;
    vec3 specular = mix(vec3(0.04), mat.albedo, mat.metallic) * 
                   light_color * pow(ndoth, spec_power) * light_intensity * attenuation;
    
    // Shadow
    float shadow = sdfl_GetShadow(p, light_dir, light_distance);
    
    // Ambient
    vec3 ambient = mat.albedo * 0.1;
    
    return ambient + (diffuse + specular) * shadow + mat.emission;
}
`)
}

func generateGlslPushScene() {
	code := `
SceneResult _scene_result = SceneResult(SDFL_MAX_DISTANCE, 0);

SceneResult sdfl_PushScene(SceneResult sr) {
    if (sr.distance < _scene_result.distance) {
        _scene_result.distance = sr.distance;
        _scene_result.materialId = sr.materialId;
    }
    return _scene_result;
}	
`
	generateCodeBoth(code)
}

func generateGlslDistSceneBegin() {
	code := `
SceneResult sdfl_GetDistScene(vec3 p) {
    // reset
    _scene_result = SceneResult(SDFL_MAX_DISTANCE, 0);
    
	SceneResult d;

`
	generateCodeBoth(code)
}

func generateGlslDistSceneEnd() {
	code := `
    return d;
}
`
	generateCodeBoth(code)
}
