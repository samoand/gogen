// In: ddn_roots, as defined in gogen config Out: map of FilePath/DDRs

package imreader

import (
	"bytes"
	"github.com/samoand/gogen/src/gogentypes"
	"github.com/witesand/gogen/ddn-processors/src/config"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"text/template"
)

var m sync.RWMutex
var fileReaderWg sync.WaitGroup

func normalizeDocument(node *yaml.Node, imSource *string) interface{} {
	return normalizeYamlData(node.Content[0], imSource)
}

func normalizeSequenceMaps(node *yaml.Node, imSource *string) interface{} {
	var result interface{}
	result = make(gogentypes.ASTNode)
	for _, el := range node.Content {
		elType := el.Content[0].Value
		elKey := el.Content[1].Value
		value := normalizeMapping(el, imSource)
		value.(gogentypes.ASTNode)["__tag"] = elType
		value.(gogentypes.ASTNode)["__name"] = elKey
		value.(gogentypes.ASTNode)["__source"] = *imSource
		delete(value.(gogentypes.ASTNode), elType)
		result.(gogentypes.ASTNode)[elKey] = value
	}
	return result
}

func normalizeSequenceScalars(node *yaml.Node) interface{} {
	return node.Value
}

func normalizeSequence(node *yaml.Node, imSource *string) interface{} {
	if len(node.Content) == 0 {
		return nil
	} else if node.Content[0].Kind == yaml.MappingNode {
		return normalizeSequenceMaps(node, imSource)
	} else if node.Content[0].Kind == yaml.ScalarNode {
		return normalizeSequenceScalars(node)
	} else {
		panic("Uknnown type of the first element of node.Content")
	}
}

func normalizeMapping(node *yaml.Node, imSource *string) interface{} {
	var result interface{}
	result = make(gogentypes.ASTNode)
	for index, el := range node.Content {
		if index%2 == 0 {
			key := el.Value
			value := normalizeYamlData(node.Content[index+1], imSource)
			if value != nil && reflect.ValueOf(value).Kind() == reflect.Map {
				value.(gogentypes.ASTNode)["__tag"] = el.Value
				value.(gogentypes.ASTNode)["__name"] = el.Value
				value.(gogentypes.ASTNode)["__source"] = *imSource
			}
			result.(gogentypes.ASTNode)[key] = value
		}
	}

	return result
}

func normalizeScalar(node *yaml.Node) interface{} {
	return node.Value
}

func normalizeAlias(node *yaml.Node) interface{} {
	return nil
}

func normalizeYamlData(node *yaml.Node, imSource *string) interface{} {
	var result interface{}
	result = make(gogentypes.ASTNode)

	switch node.Kind {
	case yaml.DocumentNode:
		result = normalizeDocument(node, imSource)
	case yaml.SequenceNode:
		result = normalizeSequence(node, imSource)
	case yaml.MappingNode:
		result = normalizeMapping(node, imSource)
	case yaml.ScalarNode:
		result = normalizeScalar(node)
	case yaml.AliasNode:
		result = normalizeAlias(node)
	}

	return result
}

func processIMFile(
	aggIM gogentypes.ASTNode,
	globalVars map[string]string,
	imRoot string) func(string, os.FileInfo, error) error {
	contentProcessor := func(path string, info os.FileInfo) {
		base := filepath.Base(path)
		splitBase := strings.Split(base, ".")
		suffix := ""
		prefix := ""
		if len(splitBase) > 1 {
			suffix = splitBase[len(splitBase)-1]
			prefix = splitBase[0]
		}
		if info.IsDir() || suffix != "yaml" || prefix == "" {
			return
		}

		tmpl, err := template.ParseFiles(path)
		if err != nil {
			panic(err)
		}
		var cb bytes.Buffer
		if err = tmpl.Execute(&cb, globalVars); err != nil {
			panic(err)
		}
		var data yaml.Node
		err = yaml.Unmarshal(cb.Bytes(), &data)

		if err != nil {
			panic(err)
		}
		imSource := strings.TrimPrefix(path, imRoot)
		if os.IsPathSeparator(imSource[0]) {
			imSource = imSource[1:]
		}
		imSource = filepath.Clean(imSource)
		normalizedData := normalizeYamlData(&data, &imSource)
		m.Lock()
		aggIM[path] = normalizedData
		// todo: inject reference to parent
		defer m.Unlock()
	}
	walker := func(path string, info os.FileInfo, err error) error {
		fileReaderWg.Add(1)
		go func() {
			defer fileReaderWg.Done()
			contentProcessor(path, info)
		}()
		return nil
	}

	return walker
}

// out: map, keys: elements of imRoot (paths), values: processed yaml
func ReadIM(imRoots []interface{}) gogentypes.ASTNode {
	result := make(gogentypes.ASTNode)
	for _, imRoot := range imRoots {
		filepath.Walk(imRoot.(string), processIMFile(
			result, config.GetGlobalVars(), imRoot.(string)))
	}
	fileReaderWg.Wait()

	return result
}

// in: configData["im_roots"], casted to []interface{}
func Run(anIn interface{}) interface{} {
	in := anIn.([]interface{})
	result := ReadIM(in)

	return result
}
