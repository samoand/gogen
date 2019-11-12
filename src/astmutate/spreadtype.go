// for each prop, mark in what type it's defined
package astmutate

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

func SpreadType(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	embelish := func(node gogentypes.ASTNode, typename string) {
		node["__struct"] = typename
	}
	structNodes := astutil.FindTags(_in, "struct", nil, 0, true)
	for _, structNode := range structNodes {
		propNodes := astutil.FindTags(structNode, "prop", nil, 0, true)
		for _, propNode := range propNodes {
			embelish(propNode, structNode["__name"].(string))
		}
	}

	return _in
}
