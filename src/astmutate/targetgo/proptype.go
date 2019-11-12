package targetgo

import (
	"log"
	"strings"

	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

/*
this method adds following data to each prop:
godep: {
	"import: {
		<url-or-string-to-import>: <package-alias-to-be-used-in-reference-to-type>
	}
	"typerepr": <type-as-appears-in-source-code>
}

 <url-or-string-to-import> is local if dot-separated-prefix is among known packages within scope
local package is prefixed by URL specific to the current distribution, as defined at "scope"
<type-as-appears-in-source-code> - type string representation
*/
func setGolangPropTypeData(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode) // this needs to correspond to scope
	if _in["__tag"] != "scope" {
		log.Fatal("SetGolangProTypeData must be called on a scope. Called on " + _in["__tag"].(string) + " instead.")
	}
	localImportPrefix := ""
	if l, ok := _in["local-import-prefix"]; ok {
		declaredImport := l.(string)
		if declaredImport[len(declaredImport)-1] == '/' {
			localImportPrefix = declaredImport
		} else {
			localImportPrefix = declaredImport + "/"
		}
	}
	types := astutil.FindTags(_in, "struct", nil, 7, true)
	var imToGoTypeMappings gogentypes.ASTNode
	if val, ok := _in["type-mappings"]; ok {
		imToGoTypeMappings = val.(gogentypes.ASTNode)
	}

	var externalPackages gogentypes.ASTNode
	if val, ok := _in["external-packages"]; ok {
		externalPackages = val.(gogentypes.ASTNode)
	}
	// take last space-separated token from property type. This is the base type

	typeDefinedInIm := func(packageName string, typeName string) bool {
		for _, tnode := range types {
			if tnode["__name"] == typeName && tnode["__package"] == packageName {
				return true
			}
		}
		return false
	}

	setGolangPropData := func(prop gogentypes.ASTNode) {
		propTypeDecl := prop["type"].(string)
		propTypeTokens := strings.Split(propTypeDecl, " ")
		propType := propTypeTokens[len(propTypeTokens)-1]
		propPrefixComponents := propTypeTokens[:len(propTypeTokens)-1]
		goPrefixComponents := make([]string, 0)
		for _, comp := range propPrefixComponents {
			if comp == "list" || comp == "[]"{
				goPrefixComponents = append(goPrefixComponents, "[]")
			} else {
				goPrefixComponents = append(goPrefixComponents, comp)
			}
		}
		goPrefix := ""
		if len(goPrefixComponents) > 0 {
			goPrefix = strings.Join(goPrefixComponents, " ") + " "
		}

		// first, check if declared type is in imToGoTypeMappings
		if propInTypeMappings, ok := imToGoTypeMappings[propType]; ok {
			importKey := propInTypeMappings.(gogentypes.ASTNode)["goimport"].(string)
			importVal := propInTypeMappings.(gogentypes.ASTNode)["alias"].(string)
			importStmt := gogentypes.ASTNode{
				importKey: importVal,
			}
			typerepr := goPrefix + propInTypeMappings.(gogentypes.ASTNode)["gotype"].(string)
			prop["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
			return
		}
		// next, check if there is package declaration in the declared type
		propParts := strings.Split(propType, ".")
		if len(propParts) > 1 {
			declaredPackage := propParts[0]
			declaredProp := propParts[1]
			// check if declared package is in external dependencies
			if externalPackage, ok := externalPackages[declaredPackage]; ok {
				importKey := externalPackage.(gogentypes.ASTNode)["goimport"].(string)
				importVal := externalPackage.(gogentypes.ASTNode)["__name"].(string)
				importStmt := gogentypes.ASTNode{
					importKey: importVal,
				}
				typerepr := goPrefix + propType
				prop["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
				return
			}
			if typeDefinedInIm(declaredPackage, declaredProp) {
				importKey := localImportPrefix + declaredPackage
				importVal := declaredPackage
				importStmt := gogentypes.ASTNode{
					importKey: importVal,
				}
				typerepr := goPrefix + importVal + "." + declaredProp
				prop["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
				return
			}
			log.Fatal("Unable to identify package dependency for type " +
				prop["__struct"].(string) + " prop " + prop["__name"].(string))
		} else {
			typerepr := goPrefix + propType
			prop["godep"] = gogentypes.ASTNode{"typerepr": typerepr}
		}
	}
	for _, typedef := range types {
		for _, prop := range astutil.FindTags(
			typedef,
			"prop",
			nil,
			0,
			true) {
			setGolangPropData(prop)
		}
	}
	return _in
}

func SetGolangPropTypeData(scope string) func(interface{}) interface{} {
	inner := func(in interface{}) interface{} {
		_in := in.(gogentypes.ASTNode)
		scopes := astutil.FindTags(_in, "scope", func(node gogentypes.ASTNode) bool {
			return node["__name"] == scope
		}, 7, true)
		if len(scopes) == 0 {
			log.Fatal("Failed to SetGolangPropTypeData because scope " + scope + " is not found")
		}
		setGolangPropTypeData(scopes[0])
		return _in
	}
	return inner
}

func bubblePropImports(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode) // this needs to correspond to scope
	for _, packageNode := range astutil.FindTags(_in, "package", nil, 7, true) {
		packageImports := gogentypes.ASTNode{}
		for _, structNode := range astutil.FindTags(packageNode, "struct", nil, 7, true) {
			typeImports := gogentypes.ASTNode{}
			for _, propNode := range astutil.FindTags(structNode, "prop", nil, 7, true) {
				godep := propNode["godep"]
				if importDef, ok := godep.(gogentypes.ASTNode)["import"]; ok {
					for url, alias := range importDef.(gogentypes.ASTNode) {
						typeImports[url] = alias
						packageImports[url] = alias
					}
				}
			}
			structNode["import"] = typeImports
		}
		packageNode["import"] = packageImports
	}
	return _in
}

func BubblePropImports(scope string) func(interface{}) interface{} {
	inner := func(in interface{}) interface{} {
		_in := in.(gogentypes.ASTNode)
		scopes := astutil.FindTags(_in, "scope", func(node gogentypes.ASTNode) bool {
			return node["__name"] == scope
		}, 7, true)
		if len(scopes) == 0 {
			log.Fatal("Failed to BubblePropImports because scope " + scope + " is not found")
		}
		bubblePropImports(scopes[0])
		return _in
	}
	return inner
}

func anonymousType(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode) // this needs to correspond to scope
	for _, propNode := range astutil.FindTags(_in, "prop", nil, 7, true) {
		if _, ok := propNode["type"]; !ok {
			propNode["type"] = propNode["__name"]
			propNode["anonymous"] = "true"
		}
	}
	return _in
}

func AnonymousType(scope string) func(interface{}) interface{} {
	inner := func(in interface{}) interface{} {
		_in := in.(gogentypes.ASTNode)
		scopes := astutil.FindTags(_in, "scope", func(node gogentypes.ASTNode) bool {
			return node["__name"] == scope
		}, 7, true)
		if len(scopes) == 0 {
			log.Fatal("Failed to BubblePropImports because scope " + scope + " is not found")
		}
		anonymousType(scopes[0])
		return _in
	}
	return inner
}
