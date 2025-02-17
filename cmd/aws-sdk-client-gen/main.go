package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

//go:generate go run ../aws-sdk-client-gen-gen/main.go
//go:generate go get ./...

const templateStr = `// Code generated by cmd/aws-sdk-client-gen/main.go; DO NOT EDIT.
package sdkclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .PkgName }}"
)

{{ range .Methods }}
func {{ $.PkgName }}_{{ .Name }}(ctx context.Context, awsCfg aws.Config, b json.RawMessage) (any, error) {
	svc := {{ $.PkgName }}.NewFromConfig(awsCfg)
	var in {{ .Input }}
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}
	return svc.{{ .Name }}(ctx, &in)
}

{{ end }}

func init() {
{{- range .Methods }}
	clientMethods["{{ $.PkgName }}#Client.{{ .Name }}"] = {{ $.PkgName }}_{{ .Name }}
{{- end }}
}
`

func main() {
	generateAll()
}

func gen(pkgName string, clientType reflect.Type, genNames []string) error {
	log.Printf("generating %s_gen.go", pkgName)

	methods := make([]map[string]string, 0)
	for i := 0; i < clientType.NumMethod(); i++ {
		method := clientType.Method(i)
		if len(genNames) > 0 && !contains(genNames, method.Name) {
			continue
		}
		params := make([]string, 0)
		for j := 0; j < method.Type.NumIn(); j++ {
			params = append(params, method.Type.In(j).String())
		}
		if len(params) <= 1 {
			log.Printf("no params func %s", method.Name)
			continue
		}
		methods = append(methods, map[string]string{
			"Name":  method.Name,
			"Input": strings.TrimPrefix(params[2], "*"),
		})
	}

	tmpl, err := template.New("clientGen").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := map[string]interface{}{
		"PkgName": pkgName,
		"Methods": methods,
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	if err := os.WriteFile(pkgName+"_gen.go", buf.Bytes(), 0644); err != nil {
		return err
	}
	log.Printf("generated %s_gen.go", pkgName)
	return nil
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
