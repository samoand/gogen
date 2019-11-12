package goutil

import (
	"bytes"
	"log"
	"strconv"
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

func buildStructFooter() []byte {
	return []byte("}\n")
}

func isAbstract(structNode gogentypes.ASTNode) bool {
	if abstract, ok := structNode["abstract"]; ok {
		parsed, err := strconv.ParseBool(abstract.(string))
		if err != nil {
			log.Fatal("Invalid boolean \"abstract\" in struct " + structNode["__name"].(string))
		}
		if parsed { // don't generate abstract classes
			return true
		}
	}
	return false
}
func buildStructContent(structNode gogentypes.ASTNode,
	contentDecorators map[string]func(gogentypes.ASTNode, []byte) []byte) ([]byte, bool) {
	if isAbstract(structNode) {
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
	if isAbstract(structNode) {
		return []byte("")
	}
	var structContent []byte
	structContent = append(structContent, buildPackage(structNode, contentDecorators)...)
	imports := getImports(structNode)
	structContent = append(structContent, buildImports(imports, contentDecorators)...)
	structContent = append(structContent, []byte("\n")...)
	structContent, concrete := buildStructContent(structNode, contentDecorators)
	if concrete {
		structContent = append(structContent, structContent...)
	}
	return decorateContent(structNode, structContent, "typeGoFile", contentDecorators)
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
			needToGen = true
		}
	}
	if needToGen {
		return decorateContent(packageNode, packageContent, "typeGoFile", contentDecorators), true
	} else {
		return nil, false
	}
}
