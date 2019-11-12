package astmutate

import "github.com/samoand/gogen/src/gogentypes"

/**
returns gogentypes.ASTNode
*/

func setParentNodes(parent gogentypes.ASTNode, current interface{}) {
	if currentAsMap, ok := current.(gogentypes.ASTNode); ok {
		for _, v := range currentAsMap {
			setParentNodes(currentAsMap, v)
		}
		currentAsMap["__parent"] = &parent
	}
}

func SetParentNodes(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	_in["__parent"] = nil
	for _, v := range _in {
		setParentNodes(_in, v)
	}
	return _in
}
