package goutil

import (
	"bytes"
	"go/format"
	"log"
	"strings"
	"text/template"

	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

// decorates content of previously built content represented as a []byte
// use to inject ad-hoc content to buff up generated directives where needed.
// node: AST Node currently being processed
// decorateKey: key to look up appropriate decorator in the map of decorators.
//	Note: decoratorKey may not be same as node tag
// contentDecorators: map of decorators passed around accessed by decoratorKey
func decorateContent(node gogentypes.ASTNode, content []byte, decorateKey string,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	if decorator, ok := contentDecorators[decorateKey]; ok {
		return decorator(node, content)
	} else {
		return content
	}
}

func getImports(node gogentypes.ASTNode) []string {
	result := []string{}
	importDecl := node["import"].(gogentypes.ASTNode)
	for uri, alias := range importDecl {
		result = append(result, alias.(string)+" \""+uri.(string)+"\"")
	}
	return result
}

func buildImports(imports []string, contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	// if no imports return newline
	if len(imports) == 0 {
		return []byte("\n")
	}
	tmpl := `
import (
{{- range $i, $a := . }}
	{{$a}}
{{- end}}
)
`
	var cb bytes.Buffer
	t := template.Must(template.New("test").Parse(tmpl))
	if err := t.Execute(&cb, imports); err != nil {
		log.Fatal(err)
	}
	return cb.Bytes()
}

func buildPackage(propNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	var result []byte

	if propNode["__tag"] == "package" {
		result = []byte("package " + propNode["__name"].(string) + "\n")
	} else {
		result = []byte("package " + propNode["__package"].(string) + "\n")
	}
	return decorateContent(propNode, result, "package", contentDecorators)
}

func buildEnum(enumNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {

	return decorateContent(
		enumNode,
		BuildRawEnum(enumNode),
		"enum",
		contentDecorators)
}

func buildStructHeader(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	return decorateContent(
		structNode,
		[]byte("type "+structNode["__name"].(string)+" struct {\n"),
		"structHeader",
		contentDecorators)
}

func buildStructProps(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	var propComponents [][]byte
	var result []byte
	props, ok := structNode["props"]
	if !ok {
		return result
	}
	for k, v := range props.(gogentypes.ASTNode) {
		if strings.HasPrefix(k.(string), "__") {
			continue
		} else {
			propComponents = append(propComponents, BuildProp(v.(gogentypes.ASTNode)))
		}
	}
	for _, c := range propComponents {
		result = append(result, c...)
	}
	return decorateContent(structNode, result, "structProps", contentDecorators)
}

func buildStructMetaMethods(structNode gogentypes.ASTNode) []byte {
	tmpl := `
func (c *{{.}}) Kind() string {
	return "{{.}}"
}

`
	var cb bytes.Buffer
	t := template.Must(template.New("test").Parse(tmpl))
	if err := t.Execute(
		&cb, structNode["__name"].(string)); err != nil {
		log.Fatal(err)
	}
	return cb.Bytes()
}

func buildStructMethods(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {

	methodBuilder := func(methodAggNode gogentypes.ASTNode, tmpl string) []byte {
		result := make([]byte, 0)
		for _, method := range astutil.FindTags(methodAggNode, "method", nil,0, true) {
			methodName := strings.Title(method["__name"].(string))
			field := method["field"].(string)
			fieldType := method["godep"].(gogentypes.ASTNode)["typerepr"].(string)
			titleComps := make([]string, 0)
			for _, comp := range strings.Split(field, ".") {
				titleComps = append(titleComps, strings.Title(comp))
			}
			field = strings.Join(titleComps, ".")

			type params struct {
				StructName	string
				MethodName 	string
				TypeRepr   	string
				Field 		string
			}

			var cb bytes.Buffer
			t := template.Must(template.New("test").Parse(tmpl))
			if err := t.Execute(
				&cb,
				params{structNode["__name"].(string),
					methodName,
					fieldType,
					field,}); err != nil {
				log.Fatal(err)
			}
			result = append(result, decorateContent(method, cb.Bytes(), methodName, contentDecorators)...)
		}

		return result
	}

	result := make([]byte, 0)
	accessorTmpl := `
func (entity *{{ $.StructName }}) {{ $.MethodName }} () {{ $.TypeRepr }} {
	return entity.{{ $.Field }}
}

`
	accessorsRoot := astutil.FindTags(structNode, "accessors", nil, 3, true)
	if len(accessorsRoot) == 1 {
		result = append(result, methodBuilder(accessorsRoot[0], accessorTmpl)...)
	}
	mutatorTmpl := `
func (entity *{{ $.StructName }}) {{ $.MethodName }} (v {{ $.TypeRepr }}) {
	entity.{{ $.Field }} = v
}

`
	mutatorsRoot := astutil.FindTags(structNode, "mutators", nil, 3, true)
	if len(accessorsRoot) == 1 {
		result = append(result, methodBuilder(mutatorsRoot[0], mutatorTmpl)...)
	}
	return result
}

func buildStructFooter() []byte {
	return []byte("}\n")
}

func buildStructContent(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) ([]byte, bool) {
	if astutil.GetBoolAtKey(structNode, "abstract", false) {
		return nil, false
	}
	components := [][]byte{
		buildStructHeader(structNode, contentDecorators),
		buildStructProps(structNode, contentDecorators),
		buildStructFooter(),
	}
	var structContent []byte
	for _, c := range components {
		structContent = append(structContent, c...)
	}

	return decorateContent(structNode, structContent, "struct", contentDecorators), true
}

func StructToGoFile(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) []byte {
	if astutil.GetBoolAtKey(structNode, "abstract", false) {
		return []byte("")
	}
	var structContent []byte
	structContent = append(structContent, buildPackage(structNode, contentDecorators)...)
	imports := getImports(structNode)
	structContent = append(structContent, buildImports(imports, contentDecorators)...)
	structContent = append(structContent, []byte("\n")...)
	structContent, _ = buildStructContent(structNode, contentDecorators)
	structContent = append(structContent, structContent...)
	structContent = decorateContent(structNode, structContent, "typeGoFile", contentDecorators)
	structContent = append(structContent, []byte(buildStructMethods(structNode, contentDecorators))...)
	if astutil.GetBoolAtKey(structNode, "gen-meta", false) {
		structContent = append(structContent, []byte(buildStructMetaMethods(structNode))...)
	}

	return structContent
}

func PackageToGoFile(packageNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) ([]byte, bool) {
	var packageContent []byte
	structNodes := astutil.FindTags(packageNode, "struct", nil, 0, true)
	enumNodes := astutil.FindTags(packageNode, "enum", nil, 0, true)
	packageContent = append(packageContent, buildPackage(packageNode, contentDecorators)...)
	imports := getImports(packageNode)
	packageContent = append(packageContent, buildImports(imports, contentDecorators)...)
	packageContent = append(packageContent, []byte("\n")...)

	needToGen := false
	for _, enumNode := range enumNodes {
		enumContent := buildEnum(enumNode, contentDecorators)
		packageContent = append(packageContent, enumContent...)
		needToGen = true
	}

	for _, structNode := range structNodes {
		structContent, concrete := buildStructContent(structNode, contentDecorators)
		if concrete {
			packageContent = append(packageContent, structContent...)
			packageContent = append(packageContent, []byte(buildStructMethods(structNode, contentDecorators))...)
			if astutil.GetBoolAtKey(structNode, "gen-meta", false) {
				packageContent = append(packageContent, []byte(buildStructMetaMethods(structNode))...)
			}
			needToGen = true
		}
	}
	if needToGen {
		packageContent = decorateContent(packageNode, packageContent, "typeGoFile", contentDecorators)
		packageContent, _ := format.Source(packageContent)
		return packageContent, true
	} else {
		return nil, false
	}
}
