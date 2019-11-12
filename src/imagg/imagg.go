package imagg

import (
	"github.com/golang/glog"
	"github.com/samoand/gogen/src/gogentypes"
	"github.com/samoand/gostructutil"
)

/**
returns gogentypes.ASTNode
*/
func Run(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	var infomodels []gogentypes.ASTNode
	for _, v := range _in {
		_v := v.(gogentypes.ASTNode)
		infomodels = append(infomodels, _v)
	}
	result, err := structutil.MergeAll(infomodels, false)
	if err != nil {
		glog.Fatal(err)
	}
	result["__tag"] = "top"
	return result
}
