// for type and for each prop,
// add info about package in which it's defined
package astmutate

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

func SpreadPackage(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	embelish := func(node gogentypes.ASTNode, packname string) {
		node["__package"] = packname
	}
	packageNodes := astutil.FindTags(_in, "package", nil, 0, true)
	for _, packageNode := range packageNodes {
		structNodes := astutil.FindTags(packageNode, "struct", nil, 0, true)
		for _, structNode := range structNodes {
			embelish(structNode, packageNode["__name"].(string))
		}
		propNodes := astutil.FindTags(packageNode, "prop", nil, 0, true)
		for _, propNode := range propNodes {
			embelish(propNode, packageNode["__name"].(string))
		}
	}

	return _in
}
