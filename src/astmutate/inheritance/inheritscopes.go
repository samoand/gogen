package inheritance

import (
	"encoding/gob"
	"errors"
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
	structutil "github.com/samoand/gostructutil"
)

func MergeInheritedScopes(in interface{}) interface{} {	
	_in := in.(gogentypes.ASTNode)
	peerScopeFinder := func(
		scopeNode gogentypes.ASTNode,
		peerScopeName string) (gogentypes.ASTNode, error) {
			candidates := astutil.FindTags(
				_in, "scope", func(node gogentypes.ASTNode) bool {
				return node["__name"] == peerScopeName
			},7, true)
			if len(candidates) == 0 {
				return nil, errors.New("Scope with name " + peerScopeName + " is not found")
			} else if len(candidates) == 0 {
				return nil, errors.New("Found multiple scopes with name " + peerScopeName)
			} else {
				return candidates[0], nil
			}
	}
	graph := buildInheritanceGraph(_in, "scope", peerScopeFinder)
	ValidateInheritanceGraph(graph)
	blank := struct{}{}
	processed := make(map[string]struct{})
	var mutateScope func(scopeName string)
	mutateScope = func(scopeName string) {
		scopeNode := graph[scopeName].Node
		if _, ok := processed[scopeName]; ok {
			return
		} else {
			processed[scopeName] = blank
		}
		inheritanceData := graph[scopeName].BaseNodes
		for _, super := range inheritanceData {
			mutateScope(super["__name"].(string))
			gob.Register(map[interface{}]interface{}{})
			scopeName := scopeNode["__name"].(string)
			cloned := structutil.CloneMap(super)
			structutil.Merge(scopeNode, cloned, false)
			scopeNode["__name"] = scopeName
		}
	}
	for scopeName, _ := range graph {
		mutateScope(scopeName)
	}
	return _in
}
