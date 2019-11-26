package astutil

import (
	"github.com/samoand/gogen/src/gogentypes"
	"log"
	"strconv"
)

func GetBoolAtKey(node gogentypes.ASTNode, key string, defaultVal bool) bool {
	if value, ok := node[key]; ok {
		parsed, err := strconv.ParseBool(value.(string))
		if err != nil {
			log.Fatal("Invalid boolean \"" + key + "\" in struct " + node["__name"].(string))
		}
		if parsed { // don't generate abstract classes
			return true
		}
	}
	return defaultVal
}

