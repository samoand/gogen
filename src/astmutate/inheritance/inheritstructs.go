package inheritance

import (
	"errors"
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
	"log"
	"strings"
)

func BuildInheritedStructData(in interface{}) interface{} {
	_in := in.(gogentypes.ASTNode)
	for _, scope := range astutil.FindTags(_in, "scope", nil, 3, true) {
		peerFinder := func(
			node gogentypes.ASTNode,
			peerId string) (gogentypes.ASTNode, error) {
			peerIdComps := strings.Split(peerId, ".")
			peerPackage := astutil.FindFirstParent(node, "package")
			peerName := peerId
			if len(peerIdComps) > 1 {
				candidates := astutil.FindTags(scope, "package",
					func(packageNode gogentypes.ASTNode) bool {
						return packageNode["__name"] == peerIdComps[0]
					},0,true)
				if len(candidates) == 0 {
					log.Fatal("Invalid package " + peerIdComps[0] + " declared for superclass of " + node["__name"].(string))
				} else if len(candidates) > 1 {
					log.Fatal("Wat? multiple packages with same name " + peerIdComps[0])
				}
				peerPackage = candidates[0]
				peerName = peerIdComps[1]
			}
			peers := astutil.FindTags(
				peerPackage,
				"struct",
				func(structNode gogentypes.ASTNode) bool {
					return structNode["__name"] == peerName
				}, 0, true)
			if len(peers) == 0 {
				return nil, errors.New(
					"Supertype " + peerId + " declared for type " +
						node["__name"].(string) + " is not found")
			} else if len(peers) > 1 {
				return nil, errors.New("Found multiple entries for type " + peerId)
			}
			return peers[0], nil
		}
		graph := buildInheritanceGraph(scope, "struct", peerFinder)
		ValidateInheritanceGraph(graph)
		blank := struct{}{}
		processed := make(map[string]struct{})
		ownPropExists := func(structNode gogentypes.ASTNode, prop string) bool {
			props := structNode["props"].(gogentypes.ASTNode)
			_, ok := props[prop]
			return ok
		}

		var mutateStruct func(typeName string)
		mutateStruct = func(typeName string) {
			structNode := graph[typeName].Node
			if _, ok := processed[typeName]; ok {
				return
			} else {
				processed[typeName] = blank
			}
			inheritanceData := graph[typeName].BaseNodes
			for _, super := range inheritanceData {
				mutateStruct(super["__name"].(string))
				superProps := super["props"].(gogentypes.ASTNode)
				for k, v := range superProps {
					superPropName := k.(string)
					foo, ok := v.(gogentypes.ASTNode)
					if !ok {
						continue
					}
					superPropData := foo
					if !ownPropExists(structNode, superPropName) {
						newTypeProp := make(gogentypes.ASTNode)
						for k, v := range superPropData {
							newTypeProp[k] = v
						}
						if superPropData["__inheritedFrom"] != nil {
							newTypeProp["__inheritedFrom"] = (superPropData["__struct"].(string) +
								"-+-" +
								superPropData["__inheritedFrom"].(string))
						} else {
							newTypeProp["__inheritedFrom"] = superPropData["__struct"]
						}
						newTypeProp["__struct"] = structNode["__name"]
						structNode["props"].(gogentypes.ASTNode)[superPropName] = newTypeProp
					}
				}
				// copy other declarations under struct if they aren't defined
				// use this function to decide whether a struct subnode should
				// be inherited
				shouldInherit := func(subnode string) bool{
					if subnode == "format-header" ||
						subnode == "format-header-post-linebreaks" ||
						subnode == "accessors" ||
						subnode == "mutators" ||
						subnode == "gen-meta" {
						return true
					} else {
						return false
					}
				}

				for k, v := range super {
					if shouldInherit(k.(string)) {
						if _, setAtStruct := structNode[k]; !setAtStruct {
							structNode[k] = v
						}
					}
				}
			}
		}
		for typeName, _ := range graph {
			mutateStruct(typeName)
		}
	}
	return _in
}
