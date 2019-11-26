package targetgo

import (
	"github.com/samoand/gogen/src/gogentypes"
	"log"
	"strings"
)

func typeDefinedInIm(
	packageName string,
	typeName string,
	types []gogentypes.ASTNode,
	localImportPrefix string) bool {
	for _, tnode := range types {
		if tnode["__name"] == typeName && tnode["__package"] == packageName {
			return true
		}
	}
	return false
}

func ProcessTypedNode(
	typedNode gogentypes.ASTNode,
	imToGoTypeMappings gogentypes.ASTNode,
	externalPackages gogentypes.ASTNode,
	types []gogentypes.ASTNode,
	localImportPrefix string) {
	propTypeDecl := typedNode["type"].(string)
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
		typedNode["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
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
			typedNode["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
			return
		}
		if typeDefinedInIm(declaredPackage, declaredProp, types, localImportPrefix) {
			importKey := localImportPrefix + declaredPackage
			importVal := declaredPackage
			importStmt := gogentypes.ASTNode{
				importKey: importVal,
			}
			typerepr := goPrefix + importVal + "." + declaredProp
			typedNode["godep"] = gogentypes.ASTNode{"import": importStmt, "typerepr": typerepr}
			return
		}
		log.Fatal("Unable to identify package dependency for type " +
			typedNode["__struct"].(string) + " typedNode " + typedNode["__name"].(string))
	} else {
		typerepr := goPrefix + propType
		typedNode["godep"] = gogentypes.ASTNode{"typerepr": typerepr}
	}
}
