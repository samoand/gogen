// mark explicitly defined properties with
// ...
// owner = ""
// ...
package astmutate

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
)

func InjectJsonDecl(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	// do nothing for now
	// propNodes := astutil.FindTags(_in, "prop", nil, 0, true)
	propNodes := astutil.FindTags(_in, "prop", nil, 0, true)
	for _, propNode := range propNodes {
		if _, ok := propNode["json"]; !ok {
			propNode["json"] = propNode["__name"]
		}
	}
	return _in
}
