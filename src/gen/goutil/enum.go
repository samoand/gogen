package goutil

import (
	"bytes"
	"github.com/samoand/gogen/src/gogentypes"
	"html/template"
	"log"
	"strings"
)

type enumTmplParams struct {
	EnumName     string
	EnumListName string
	EnumValues   []string
}

func buildRawEnum(params enumTmplParams) []byte {
	result := ""
	tmpl := `
type {{ .EnumName }} string

const (
{{- range $i, $a := .EnumValues }}
	{{ $.EnumName }}{{$a}} {{ $.EnumName }} = "{{$a}}"
{{- end}}
)

const (
{{- range $i, $a := .EnumValues }}
	{{$a}} = {{$i}}
{{- end}}
)

var {{ $.EnumListName }} = [...]{{ $.EnumName }} { {{- range $i, $a := .EnumValues }} {{ $.EnumName }}{{$a}}, {{- end }} }

`
	var cb bytes.Buffer
	t := template.Must(template.New("test").Parse(tmpl))
	if err := t.Execute(&cb, params); err != nil {
		log.Fatal(err)
	}
	return cb.Bytes()
	return []byte(result)
}

func BuildRawEnum(enumNode gogentypes.ASTNode) []byte {
	enumValues := make([]string, 0)
	for _, val := range strings.Split(enumNode["values"].(string), ",") {
		enumValues = append(enumValues, strings.TrimSpace(val))
	}
	return buildRawEnum(enumTmplParams{
		enumNode["__name"].(string),
		enumNode["__name"].(string) + "Enums",
		enumValues})
}
