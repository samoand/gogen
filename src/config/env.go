package config

import (
	"flag"
	typeutil "github.com/samoand/gogen/src/typeutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type globalConstants struct {
	//ProjectRoot string
	GenRoot               string
	ImRoot                string
	ConfigRoot            string
	GoControllerModelRoot string
}

var globalVars map[string]string
var envInitialized = false

func buildEnvVars() map[string]string {
	result := make(map[string]string)
	for _, vardef := range os.Environ() {
		pair := strings.Split(vardef, "=")
		result[pair[0]] = pair[1]
	}
	return result
}

func buildDefaultVars() map[string]string {
	result := make(map[string]string)
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		dir, _ := filepath.Split(filename)
		result["DNDNR_GOGEN_ROOT"], _ = filepath.Abs(
			filepath.Join(dir, "..", ".."))
	} else {
		panic("ERROR: failed to determine pathname for env.goutil, exiting")
	}
	result["DNDNR_GOGEN_CONFIG_ROOT"] = filepath.Join(
		result["DNDNR_GOGEN_ROOT"], "config")
	result["DNDNR_IM_ROOT"] = filepath.Join(
		result["DNDNR_GOGEN_ROOT"], "im")
	return result
}

func GetGlobalVars() map[string]string {
	if !envInitialized {
		globalVars = make(map[string]string)
		for k, v := range buildDefaultVars() {
			globalVars[k] = v
		}
		for k, v := range buildEnvVars() {
			globalVars[k] = v
		}
		envInitialized = true

		//jsonified, _ := json.MarshalIndent(globalVars, "", "  ")
		//fmt.Println("*** GOGEN globalVars initialized as follows ***\n" + string(jsonified))
	}
	return globalVars
}

func SetupLog(configData map[string]interface{}) {
	flag.Parse()
	flag.Set("stderrthreshold", configData["log_level"].(string))
	flag.Set("log_dir", configData["log_dir"].(string))
	flag.Set("logtostderr", typeutil.BoolToString(configData["logtostderr"]))
	flag.Set("alsologtostderr", typeutil.BoolToString(configData["alsologtostderr"]))
}
