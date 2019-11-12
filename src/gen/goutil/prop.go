package goutil

import (
	"strconv"
	"strings"

	"github.com/samoand/gogen/src/gogentypes"
)

func BuildProp(propNode gogentypes.ASTNode) []byte {
	result := "\t"
	anon := false
	if v, ok := propNode["anonymous"]; ok {
		anon, _ = strconv.ParseBool(v.(string))
	}
	if !anon {
		result += strings.Title(propNode["__name"].(string)) + "\t"
	}
	result += propNode["godep"].(gogentypes.ASTNode)["typerepr"].(string)
	protobuf_decl, protobuf_decl_exists := propNode["protobuf"]
	json_decl, json_decl_exists := propNode["json"]
	if json_decl_exists || protobuf_decl_exists {
		result += "\t`"
	}

	if json_decl_exists {
		result += "json:\"" + json_decl.(string) + "\""
		if protobuf_decl_exists {
			result += " "
		}
	}
	if protobuf_decl_exists {
		result += "protobuf:\"" + protobuf_decl.(string) + "\""
	}
	if json_decl_exists || protobuf_decl_exists {
		result += "`"
	}
	result += "\n"
	return []byte(result)
}
