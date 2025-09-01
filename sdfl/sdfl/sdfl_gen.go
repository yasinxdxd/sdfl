package sdfl

import "fmt"

var functionSymbols = map[string]FunDef{
	"scene":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_SCENE, Id: "scene", FunDefArgNames: []string{"background", "camera", "children"}},
	"camera":             {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_CAMERA, Id: "camera", FunDefArgNames: []string{"position"}},
	"plane":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "plane", FunDefArgNames: []string{"height"}},
	"sphere":             {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "sphere", FunDefArgNames: []string{"position", "radius"}},
	"box":                {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "box", FunDefArgNames: []string{"position", "size"}},
	"torus":              {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "torus", FunDefArgNames: []string{"position", "radius", "thickness"}},
	"rotateAround":       {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_ROTATE_AROUND, Id: "rotateAround", FunDefArgNames: []string{"position", "rotation", "child"}},
	"smoothUnion":        {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothUnion", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
	"smoothSubtraction":  {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothSubtraction", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
	"smoothIntersection": {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_OP, Id: "smoothIntersection", FunDefArgNames: []string{"child1", "child2", "smooth_transition"}},
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
	generateGlslComputeHeader()
	generateGlslBuiltinSDFFunctions()
	sceneCall := prog.Exprs[0].FunCall
	if sceneCall.Id != "scene" {
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

	generateGlslDistSceneBegin()
	for _, expr := range childrenArr.Exprs {
		expr.generate()
	}
	generateGlslDistSceneEnd()

	generateGlslRaymarchEngine()

	generateGlslFragmentMain(cameraCall)
	generateGlslComputeMain()
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

func undoFreshVar(base string) {
	if varCounters[base] > 0 {
		varCounters[base]--
	}
}

func (funCall *FunCall) generate(args ...any) {
	rayPosition := "p"
	if len(args) > 0 {
		rayPosition = args[0].(string)
	}

	funDef, ok := functionSymbols[funCall.Id]
	println(funCall.Id, funDef.SymbolType, rayPosition)

	if ok {

		orderedArgs := func() []*Expr {
			exprs := []*Expr{}
			// figure out named parameter order
			for j := 0; j < len(funDef.FunDefArgNames); j++ {
				funNamedArg, ok := funCall.FunNamedArgs[funDef.FunDefArgNames[j]]
				if !ok {
					// TODO: better error messages
					fmt.Println("ERROR!")
				}
				if funDef.FunDefArgNames[j] == funNamedArg.ArgName {
					exprs = append(exprs, &funNamedArg.Expr)
				}
			}
			return exprs
		}

		genFunCall := func(funId string) string {
			symbol := "sdfl_builtin_" + funId
			return symbol
		}

		switch funDef.SymbolType {
		// case FUN_BUILTIN_SCENE:
		// case FUN_BUILTIN_CAMERA:
		case FUN_BUILTIN_ROTATE_AROUND:
			posExpr, okPos := funCall.FunNamedArgs["position"]
			rotExpr, okRot := funCall.FunNamedArgs["rotation"]
			childExpr, okChild := funCall.FunNamedArgs["child"]

			if !okPos || !okRot || !okChild {
				fmt.Println("ERROR: rotateAround missing args (needs position, rotation, child)")
				return
			}

			qVar := freshVar("q")

			// subtract pivot
			generateCodeBoth(fmt.Sprintf("    vec3 %s = p - ", qVar))
			posExpr.Expr.generate()
			generateCodeBoth(";\n")

			// rotate
			generateCodeBoth(fmt.Sprintf("    %s = sdfl_RotationMatrix(radians(", qVar))
			rotExpr.Expr.generate()
			generateCodeBoth(fmt.Sprintf(")) * %s;\n", qVar))

			// add pivot back
			generateCodeBoth(fmt.Sprintf("    %s += ", qVar))
			posExpr.Expr.generate()
			generateCodeBoth(";\n")

			// recurse with qVar instead of p
			childExpr.Expr.FunCall.generate(qVar)
		case FUN_BUILTIN_OP:
			exprs := orderedArgs()
			// child1
			println(rayPosition)
			exprs[0].FunCall.generate(rayPosition)
			// child2
			println(rayPosition)
			exprs[1].FunCall.generate(rayPosition)
			sd := freshVar("sd")
			generateCodeBoth("    float %s = %s(", sd, genFunCall(funDef.Id))
			undoFreshVar("sd")
			undoFreshVar("sd")
			sd = freshVar("sd")
			generateCodeBoth(sd)
			generateCodeBoth(", ")
			undoFreshVar("sd")
			undoFreshVar("sd")
			sd = freshVar("sd")
			generateCodeBoth(sd)
			generateCodeBoth(", ")
			// smooth_transition
			exprs[2].generate()
			generateCodeBoth(");\n")

			generateCodeBoth("    d = %s(", "sdfl_PushScene")
			sd = freshVar("sd")
			sd = freshVar("sd")
			generateCodeBoth(sd)
			generateCodeBoth(");\n")

		case FUN_BUILTIN:
			exprs := orderedArgs()

			sd := freshVar("sd")
			generateCodeBoth("    float %s = %s(%s, ", sd, genFunCall(funDef.Id), rayPosition)
			for i, e := range exprs {
				e.generate()
				if i < len(exprs)-1 {
					generateCodeBoth(",")
				}
			}
			generateCodeBoth(");\n")

			generateCodeBoth("    d = %s(", "sdfl_PushScene")
			undoFreshVar("sd")
			sd = freshVar("sd")
			generateCodeBoth(sd)
			generateCodeBoth(");\n")
		default:
			fmt.Println("NOT IMPLEMENTED YET!")
		}
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
    
    float d = sdfl_RayMarch(ray_origin, ray_dir);
`)
}

func generateGlslFragmentMain(cameraFunCall *FunCall) {
	generateFragmentCode(`
void main() {
    vec2 uv = o_vertex_uv * 2. - 1.;
    uv.y *= float(window_size.y) / float(window_size.x);
`)

	generateGlslCamera(cameraFunCall)

	generateFragmentCode(`
    // lightning
    vec3 p = ray_origin + ray_dir * d;
    float diff = sdfl_GetLight(p); // diffuse lightning
    vec3 color = vec3(diff);
    frag_color = vec4(color, 1.0);
}
`)
}

func generateGlslComputeMain() {
	generateComputeCode(`
void main() {
    ivec3 gid = ivec3(gl_GlobalInvocationID);
    if (any(greaterThanEqual(gid, ivec3(resolution)))) return;

    vec3 uv = vec3(gid) / float(resolution - 1);
    vec3 p = mix(minBound, maxBound, uv);

    // float d = sdSphere(p, 0.3);
    float d = sdfl_GetDistScene(p);

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

// https://iquilezles.org/articles/distfunctions/

float sdfl_builtin_smoothUnion(float d1, float d2, float k) {
    float h = clamp( 0.5 + 0.5*(d2-d1)/k, 0.0, 1.0 );
    return mix( d2, d1, h ) - k*h*(1.0-h);
}

float sdfl_builtin_smoothSubtraction(float d1, float d2, float k) {
    float h = clamp( 0.5 - 0.5*(d2+d1)/k, 0.0, 1.0 );
    return mix( d2, -d1, h ) + k*h*(1.0-h);
}

float sdfl_builtin_smoothIntersection(float d1, float d2, float k) {
    float h = clamp( 0.5 - 0.5*(d2-d1)/k, 0.0, 1.0 );
    return mix( d2, d1, h ) + k*h*(1.0-h);
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
`
	generateFragmentCode(code)
	generateComputeCode(code)
}

func generateGlslRaymarchEngine() {
	generateFragmentCode(`
float sdfl_RayMarch(vec3 ray_origin, vec3 ray_dir) {
    float dfo = 0.; // distance from ray origin

    for (int i = 0; i < SDFL_MAX_STEPS; i++) {
        vec3 p = ray_origin + ray_dir * dfo;
        
        float ds = sdfl_GetDistScene(p); // distance to the scene
        dfo += ds;

        if (dfo > SDFL_MAX_DISTANCE || ds < SDFL_HIT_DISTANCE) break;
    }

    return dfo;
}

vec3 sdfl_GetNormal(vec3 p) {
    float d = sdfl_GetDistScene(p);
    vec2 off = vec2(.01, 0.);

    vec3 normal = vec3(
        d - sdfl_GetDistScene(p - off.xyy),
        d - sdfl_GetDistScene(p - off.yxy),
        d - sdfl_GetDistScene(p - off.yyx)
    );

    return normalize(normal);
}

float sdfl_GetLight(vec3 p) {
    vec3 light_pos = vec3(0, 8, 3);
    // light_pos.xz += vec2(sin(elapsed_time * 0.1), cos(elapsed_time * 0.1)) * 2;
    vec3 light_dir = normalize(light_pos - p);
    vec3 normal_p = sdfl_GetNormal(p);

    float diff = clamp(dot(normal_p, light_dir), 0., 1.); // diffuse lightning, clamp it because dot gives between -1 and 1

    // shadows
    float d = sdfl_RayMarch(p + normal_p*SDFL_SHADOW_CAST_DISTANCE, light_dir);
    if (d < distance(light_pos, p)) {
        diff *= 0.1;
    }

    return diff;
}
`)
}

func generateGlslDistSceneBegin() {
	code := `
float _scene_dist = SDFL_MAX_DISTANCE;

float sdfl_PushScene(float d) {
    _scene_dist = min(_scene_dist, d);
    return _scene_dist;
}

float sdfl_GetDistScene(vec3 p) {
    // reset
    _scene_dist = SDFL_MAX_DISTANCE;

    // generated sdf shapes
    float d;
`
	generateFragmentCode(code)
	generateComputeCode(code)
}

func generateGlslDistSceneEnd() {
	code := `
    return d;
}
`
	generateFragmentCode(code)
	generateComputeCode(code)
}
