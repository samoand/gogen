package targetgo

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
	"log"
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
func setTypedElementData(in interface{}) interface{} {
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
	structs := astutil.FindTags(_in, "struct", nil, 7, true)
	var imToGoTypeMappings gogentypes.ASTNode
	if val, ok := _in["type-mappings"]; ok {
		imToGoTypeMappings = val.(gogentypes.ASTNode)
	}

	var externalPackages gogentypes.ASTNode
	if val, ok := _in["external-packages"]; ok {
		externalPackages = val.(gogentypes.ASTNode)
	}
	for _, typedef := range structs {
		for _, prop := range astutil.FindTags(
			typedef,
			"prop",
			nil,
			0,
			true) {
			ProcessTypedNode(prop, imToGoTypeMappings, externalPackages, structs, localImportPrefix)
		}
	}
	for _, typedef := range structs {
		for _, prop := range astutil.FindTags(
			typedef,
			"method",
			nil,
			0,
			true) {
			ProcessTypedNode(prop, imToGoTypeMappings, externalPackages, structs, localImportPrefix)
		}
	}
	return _in
}

func SetTypedElementData(scope string) func(interface{}) interface{} {
	inner := func(in interface{}) interface{} {
		_in := in.(gogentypes.ASTNode)
		scopes := astutil.FindTags(_in, "scope", func(node gogentypes.ASTNode) bool {
			return node["__name"] == scope
		}, 7, true)
		if len(scopes) == 0 {
			log.Fatal("Failed to SetTypedElementData because scope " + scope + " is not found")
		}
		setTypedElementData(scopes[0])
		return _in
	}
	return inner
}

func bubbleImports(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode) // this needs to correspond to scope
	for _, packageNode := range astutil.FindTags(_in, "package", nil, 7, true) {
		packageImports := gogentypes.ASTNode{}
		for _, structNode := range astutil.FindTags(packageNode, "struct", nil, 7, true) {
			typeImports := gogentypes.ASTNode{}
			bubble := func(tagName string) {
				for _, propNode := range astutil.FindTags(structNode, tagName, nil, 7, true) {
					godep := propNode["godep"]
					if importDef, ok := godep.(gogentypes.ASTNode)["import"]; ok {
						for url, alias := range importDef.(gogentypes.ASTNode) {
							typeImports[url] = alias
							packageImports[url] = alias
						}
					}
				}
			}
			bubble("prop")
			bubble("method")
			structNode["import"] = typeImports
		}
		packageNode["import"] = packageImports
	}
	return _in
}

func BubbleImports(scope string) func(interface{}) interface{} {
	inner := func(in interface{}) interface{} {
		_in := in.(gogentypes.ASTNode)
		scopes := astutil.FindTags(_in, "scope", func(node gogentypes.ASTNode) bool {
			return node["__name"] == scope
		}, 7, true)
		if len(scopes) == 0 {
			log.Fatal("Failed to BubbleImports because scope " + scope + " is not found")
		}
		bubbleImports(scopes[0])
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
			log.Fatal("Failed to BubbleImports because scope " + scope + " is not found")
		}
		anonymousType(scopes[0])
		return _in
	}
	return inner
}
