package main

//goutil:generate echo Hello, two!

import (
	"../../internal/mock"
	"fmt"
	"github.com/samoand/gogen/src/config"
	"github.com/samoand/gogen/src/gogentypes"
	"gopkg.in/yaml.v3"
	"log"
)

var dataA = `
a: Easy!
b:
 c: 2
 d: [3, 4]
`
var dataT = `
a: Hard!
b:
 c: 5
 d: [6, 7]
`

func doA() {
	a := mock.A{}
	err := yaml.Unmarshal([]byte(dataA), &a)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- a:\n%v\n\n", a)

	d, err := yaml.Marshal(&a)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- a dump:\n%s\n\n", string(d))

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(dataA), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- am:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- am dump:\n%s\n\n", string(d))
}

func doT() {
	t := mock.T{}
	err := yaml.Unmarshal([]byte(dataT), &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", t)

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t dump:\n%s\n\n", string(d))

	m := make(gogentypes.ASTNode)

	err = yaml.Unmarshal([]byte(dataT), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- tm:\n%v\n\n", m)

	d, err = yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- tm dump:\n%s\n\n", string(d))
}

func main() {
	doA()
	doT()
	config := config.Configuration{
		Val: "",
		Proxy: struct {
			Address string
			Port    string
		}{"", ""},
	}
	fmt.Print(config)
}
