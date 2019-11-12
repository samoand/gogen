package inheritance

import (
	"github.com/samoand/gogen/src/astutil"
	"github.com/samoand/gogen/src/gogentypes"
	"log"
	"strings"
)

type TypeInheritance struct {
	Node      gogentypes.ASTNode   // subtype
	BaseNodes []gogentypes.ASTNode // supertypes for the subtype
}

func (in *TypeInheritance) AddSupertypeNode(supertypeNode gogentypes.ASTNode) []gogentypes.ASTNode {
	in.BaseNodes = append(in.BaseNodes, supertypeNode)
	return in.BaseNodes
}

// return slice of declared supertypes for a node
func declaredSupertypes(node gogentypes.ASTNode) []string {
	result := make([]string, 0)
	contains := func(list []string, s string) bool {
		for _, el := range list {
			if el == s {
				return true
			}
		}
		return false
	}

	supertypesDecl, ok := node["extends"]
	if ok {
		supertypes := strings.Split(supertypesDecl.(string), ",")
		for _, supertype := range supertypes {
			supertype = strings.TrimSpace(supertype)
			if !contains(result, supertype) {
				result = append(result, supertype)
			} else {
				log.Fatal("Repeated supertype declaration. Supertype: " + supertype + ", subtype: " + node["__name"].(string))
			}
		}
	}

	return result
}

// this builds graph of inheritance rules specified in IM
func buildInheritanceGraph(
	in gogentypes.ASTNode, tag string,
	peerFinder func(
		node gogentypes.ASTNode,
		peerId string) (gogentypes.ASTNode, error)) map[string]*TypeInheritance {
	result := make(map[string]*TypeInheritance)

	inner := func(node gogentypes.ASTNode) {
		if _, ok := result[node["__name"].(string)]; !ok {
			result[node["__name"].(string)] = &TypeInheritance{
				Node:      node,
				BaseNodes: make([](gogentypes.ASTNode), 0),
			}
			supertypes := declaredSupertypes(node)
			for _, supertype := range supertypes {
				supertypeNode, err := peerFinder(node, supertype)
				if err != nil {
					log.Fatal(err)
				} else {
					inheritanceData := result[node["__name"].(string)]
					inheritanceData.AddSupertypeNode(supertypeNode)
				}
			}
		}
	}

	nodes := astutil.FindTags(in, tag, nil, 0, true)
	for _, node := range nodes {
		inner(node)
	}
	return result
}

// valid graph would be a directed acyclic graph (DAL)
// this function checks that there are no cycles
// it returns pointers to those nodes that are involved in cycles
func findCycles(inheritanceGraph map[string]*TypeInheritance) map[string]struct{} {
	initFalses := func() map[string]bool {
		result := make(map[string]bool)
		for key, _ := range inheritanceGraph {
			result[key] = false
		}
		return result
	}
	visited := initFalses()
	recStack := initFalses()
	blank := struct{}{}
	inCycle := make(map[string]struct{})
	var checkCycle func(nodeName string) bool
	checkCycle = func(nodeName string) bool {
		visited[nodeName] = true
		recStack[nodeName] = true
		inheritanceData := inheritanceGraph[nodeName]
		for _, superType := range inheritanceData.BaseNodes {
			superTypeName := superType["__name"].(string)
			if !visited[superTypeName] {
				if checkCycle(superTypeName) {
					inCycle[nodeName] = blank
					return true
				}
			} else if recStack[superTypeName] {
				inCycle[nodeName] = blank
				return true
			}
		}
		recStack[nodeName] = false
		return false
	}

	for nodeName, _ := range inheritanceGraph {
		if !visited[nodeName] {
			checkCycle(nodeName)
		}
	}
	return inCycle
}

func ValidateInheritanceGraph(graph map[string]*TypeInheritance) {
	cycles := findCycles(graph)
	if len(cycles) > 0 {
		msg := "Circular inheritance detected for types: "
		index := 0
		for name, _ := range cycles {
			msg += name
			if index < len(cycles)-1 {
				msg += ", "
			}
			index++
		}
		log.Fatal(msg)
	}
}

