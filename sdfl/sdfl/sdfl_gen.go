package sdfl

import "fmt"

var functionSymbols = map[string]FunDef{
	"scene":  {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_SCENE, Id: "scene", FunDefArgNames: []string{"background", "camera", "children"}},
	"camera": {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN_CAMERA, Id: "camera", FunDefArgNames: []string{"position"}},
	"plane":  {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "plane", FunDefArgNames: []string{"height"}},
	"sphere": {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "sphere", FunDefArgNames: []string{"position", "radius"}},
	"box":    {Type: AST_FUN_DEF, SymbolType: FUN_BUILTIN, Id: "box", FunDefArgNames: []string{"position", "size"}},
}

var generatedCode = ""

func Reset() {
	generatedCode = ""
}

func GetCode() string {
	return generatedCode
}

type generator interface {
	generate()
}

func generateCode(code string, args ...any) {
	generatedCode += fmt.Sprintf(code, args...)
}

func Generate(prog *Program) {
	prog.generate()
}

func (prog *Program) generate() {
	generateGlslHeader()
	generateGlslBuiltinFunctions()
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

	generateGlslMain(cameraCall)
}

func (expr *Expr) generate() {
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

func (funCall *FunCall) generate() {
	funDef, ok := functionSymbols[funCall.Id]

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
		case FUN_BUILTIN:
			exprs := orderedArgs()
			// println(len(exprs))
			generateCode("    d = %s(%s(p, ", "sdfl_PushScene", genFunCall(funDef.Id))
			for i, e := range exprs {
				e.generate()
				if i < len(exprs)-1 {
					generateCode(",")
				}
			}
			generateCode("));\n")
		default:
			fmt.Println("NOT IMPLEMENTED YET!")
		}
	}
}

func (tuple *Tuple) generate() {
	generateCode("vec3(%s, %s, %s)", tuple.Values[0], tuple.Values[1], tuple.Values[2])
}

func (number *Number) generate() {
	generateCode(number.Value)
}

func (arrExpr *ArrExpr) generate() {
	for _, expr := range arrExpr.Exprs {
		expr.generate()
	}
}

func generateGlslCamera(cameraFunCall *FunCall) {
	generateCode("    // generated camera position\n")
	generateCode("    vec3 ray_origin = ")
	cameraFunCall.FunNamedArgs["position"].Expr.Tuple.generate()
	generateCode(";")
	generateCode(`
    vec3 ray_dir = normalize(vec3(uv, -1)); // ray direction for the each pixel
    
    float d = sdfl_RayMarch(ray_origin, ray_dir);
`)
}

func generateGlslMain(cameraFunCall *FunCall) {
	generateCode(`
void main() {
    vec2 uv = o_vertex_uv * 2. - 1.;
    uv.y *= float(window_size.y) / float(window_size.x);
`)

	generateGlslCamera(cameraFunCall)

	generateCode(`
    // lightning
    vec3 p = ray_origin + ray_dir * d;
    float diff = sdfl_GetLight(p); // diffuse lightning
    vec3 color = vec3(diff);
    frag_color = vec4(color, 1.0);
}
`)

}

func generateGlslHeader() {
	generateCode(`
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

func generateGlslBuiltinFunctions() {
	generateCode(`
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
`)

}

func generateGlslRaymarchEngine() {
	generateCode(`
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
	generateCode(`
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
`)
}

func generateGlslDistSceneEnd() {
	generateCode(`
    return d;
}
`)
}
