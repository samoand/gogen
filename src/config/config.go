package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"html/template"
)

var config interface{}
var configInitialized = false

func InitConfig(configPath string) {
	if !configInitialized {
		tmpl, err := template.ParseFiles(configPath)
		if err != nil {
			panic(err)
		}
		var cb bytes.Buffer
		if err = tmpl.Execute(&cb, GetGlobalVars()); err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(cb.Bytes(), &config)
		if err != nil {
			panic(err)
		}
		configInitialized = true

		jsonified, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println("*** GOGEN initialized as follows ***\n" + string(jsonified))

	}
}

func GetConfig() interface{} {
	if !configInitialized {
		panic("Config is not initialized")
	}

	return config
}
