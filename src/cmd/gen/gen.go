package main

import (
	"encoding/json"
	"fmt"
	"github.com/samoand/gogen/src/astmutate/targetgo"
	"io/ioutil"
	"path/filepath"

	"github.com/samoand/gogen/src/astmutate"
	"github.com/samoand/gogen/src/astmutate/inheritance"
	"github.com/samoand/gogen/src/config"
	"github.com/samoand/gogen/src/gogentypes"
	"github.com/samoand/gogen/src/imagg"
	"github.com/samoand/gogen/src/imreader"
	pmapreduce "github.com/samoand/gopmapreduce"
	structutil "github.com/samoand/gostructutil"
)

func initPipelineFuncRepos() {
	pmapreduce.PMapperRepo = make(map[string]func(interface{}) interface{})
	pmapreduce.ReducerRepo = make(map[string]func([]interface{}) interface{})
	pmapreduce.PMapperRepo["imreader"] = imreader.Run
	pmapreduce.PMapperRepo["imagg"] = imagg.Run
	pmapreduce.PMapperRepo["scope-merger"] = inheritance.MergeInheritedScopes
	pmapreduce.PMapperRepo["node-parents"] = astmutate.SetParentNodes
	pmapreduce.PMapperRepo["init-props"] = astmutate.InitProps
	pmapreduce.PMapperRepo["spread-package"] = astmutate.SpreadPackage
	pmapreduce.PMapperRepo["spread-type"] = astmutate.SpreadType
	pmapreduce.PMapperRepo["go-prop-types"] = targetgo.SetGolangPropTypeData("scope3")
	pmapreduce.PMapperRepo["go-bubble-imports"] = targetgo.BubblePropImports("scope3")
	pmapreduce.PMapperRepo["inherited-props"] = inheritance.BuildInheritedStructData

	pmapreduce.ReducerRepo[""] = pmapreduce.ReduceToFirst
	pmapreduce.ReducerRepo["first-mapped"] = pmapreduce.ReduceToFirst
	pmapreduce.ReducerRepo["none"] = pmapreduce.ReduceToNone
}

/**
in: ref to map which is configData. Required info is at (*configData)["ddn_def_path"]
return: content of yaml which describes pipeline: the sequence of mappers and reducers
*/
func GetPipelineConfig(configData map[string]interface{}) []byte {
	ddn_def_path := configData["ddn_def_path"].(string)
	content, err := ioutil.ReadFile(ddn_def_path)
	if err != nil {
		panic("ddn_def_path misconfigured, set to " + ddn_def_path)
	}

	return content
}

func getConfigPath(configType string) (string, error) {
	configDir := config.GetGlobalVars()["DNDNR_GOGEN_CONFIG_ROOT"]
	configFilename := "config." + configType + ".yaml"
	return filepath.Abs(filepath.Join(configDir, configFilename))
}

func main() {
	configPath, err := getConfigPath("dev")
	if err != nil {
		panic(err)
	}
	config.InitConfig(configPath)
	configData := config.GetConfig().(map[string]interface{})
	config.SetupLog(configData)
	imRoots := configData["im_roots"].([]interface{})
	initPipelineFuncRepos()
	pipelineConfig := GetPipelineConfig(configData)
	im := pmapreduce.RunPipeline(pipelineConfig, imRoots).(gogentypes.ASTNode)
	if im != nil {
		fmt.Println("Done")
	}
	special := make(map[interface{}]func(interface{}) string)
	special["__parent"] = func(parent interface{}) string {
		if parent != nil {
			return (*(parent.(*gogentypes.ASTNode)))["__tag"].(string)
		} else {
			return ""
		}
	}
	keysStringified := structutil.Stringify(im, special)
	jsonified, _ := json.MarshalIndent(keysStringified, "", "  ")
	fmt.Println(string(jsonified))
}
