// mark explicitly defined properties with
// ...
// owner = ""
// ...
package astmutate

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

func InitProps(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	// do nothing for now
	// propNodes := astutil.FindTags(_in, "prop", nil, 0, true)
	structNodes := astutil.FindTags(_in, "struct", nil, 0, true)
	for _, structNode := range structNodes {
		if structNode["props"] == nil {
			structNode["props"] = gogentypes.ASTNode{}
		}
	}
	return _in
}
