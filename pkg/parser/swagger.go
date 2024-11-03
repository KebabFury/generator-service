package parser

import (
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var pythonModelsFixRe = regexp.MustCompile(`(?mU):\s([a-zA-Z0-9-_, \]\[]*)\s=\sField\(\s*\.\.\.`)
var getFirstClassRe = regexp.MustCompile(`(?m)class`)
var semaphore = make(chan struct{}, 1)

type SwaggerParser struct {
}

func NewSwaggerParser() *SwaggerParser {
	return &SwaggerParser{}
}

var swaggerRe = regexp.MustCompile(`(?m)"swagger":\s?"2\.0"`)

func (s *SwaggerParser) Parse(swaggerFile []byte) *Document {
	semaphore <- struct{}{}
	defer func() {
		<-semaphore
	}()

	schemaJSON := swaggerFile

	var openapi3Spec openapi3.T
	if swaggerRe.Match(swaggerFile) {
		var doc openapi2.T
		err := json.Unmarshal(swaggerFile, &doc)
		if err != nil {
			panic(err)
		}
		apiV3, err := openapi2conv.ToV3(&doc)
		if err != nil {
			panic(err)
		}
		openapi3Spec = *apiV3

		schemaJSON, err = json.Marshal(openapi3Spec)
		if err != nil {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(swaggerFile, &openapi3Spec); err != nil {
			panic(err)
		}
	}

	err := os.WriteFile("./tmp/swagger_v3.json", schemaJSON, 0644)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("datamodel-codegen", "--input", "./tmp/swagger_v3.json", "--output", "./tmp/models.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Ошибка при выполнении команды:", err)
		panic(err)
	} else {
		fmt.Println("Команда выполнена успешно.")
	}

	pythonFile, err := os.ReadFile("./tmp/models.py")
	if err != nil {
		panic(err)
	}

	// NEEDS TO DO NOT OPTIONAL PROPS
	// PIZDETS HUETA
	pythonFileContent := pythonModelsFixRe.ReplaceAllStringFunc(string(pythonFile), func(s string) string {
		t := strings.TrimSpace(strings.Split(s, "=")[0])
		t = t[1:]
		if strings.HasPrefix(t, " Optional[") {
			t = strings.TrimPrefix(t, " Optional[")
			t = strings.TrimSuffix(t, "]")
		}
		return strings.Replace(s, "...", t, 1)
	})

	schemas := s.ParseSchemaMap(pythonFileContent)
	parseDocument := Document{
		Models: template.HTML(pythonFileContent[getFirstClassRe.FindStringIndex(pythonFileContent)[0]:]),
	}
	for path, pathItem := range openapi3Spec.Paths.Map() {

		operations := map[string]*openapi3.Operation{}
		switch {
		case pathItem.Get != nil:
			operations[http.MethodGet] = pathItem.Get
		case pathItem.Post != nil:
			operations[http.MethodPost] = pathItem.Post
		case pathItem.Put != nil:
			operations[http.MethodPut] = pathItem.Put
		case pathItem.Delete != nil:
			operations[http.MethodDelete] = pathItem.Delete
		case pathItem.Patch != nil:
			operations[http.MethodPatch] = pathItem.Patch
		}

		for method, operation := range operations {
			apiCall := ApiCall{
				Type:   ApiCallTypeCustom,
				Url:    openapi3Spec.Servers[0].URL + path,
				Method: method,
			}
			var annotations []Annotated
			apiCall.Name = operation.OperationID
			apiCall.Description = operation.Summary + " " + operation.Description
			for _, parameter := range operation.Parameters {
				if parameter.Value == nil {
					continue
				}

				param := s.ParseParameter(operation.OperationID, parameter.Value)
				apiCall.Parameters = append(apiCall.Parameters, param.Parameter)
				annotations = append(annotations, param.Annotated)
			}

			if operation.RequestBody != nil {
				bodyParams := s.ParseRequestBody(schemas, operation.RequestBody.Value)
				for _, param := range bodyParams {
					apiCall.Parameters = append(apiCall.Parameters, param)
				}
			}

			ressponse := s.ParseResponse(schemas, operation.Responses)
			for _, parameter := range ressponse {
				apiCall.Returns = append(apiCall.Returns, parameter)
			}
			parseDocument.ApiCalls = append(parseDocument.ApiCalls, apiCall)
		}

	}

	//kk, err := json.Marshal(parseDocument)
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = os.WriteFile("doc.json", kk, 0644)
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = os.WriteFile("do.md", []byte(DocumentToMarkdown(&parseDocument)), 0644)
	//if err != nil {
	//	panic(err)
	//}
	return &parseDocument
}

var classRe = regexp.MustCompile(`(?m)^class (.*)\(BaseModel\):$`)
var propNameRe = regexp.MustCompile(`(?m)([a-zA-Z_-]*):`)
var paramRe = regexp.MustCompile(`(?mU)([a-z]*)='(.*)'`)

func (s *SwaggerParser) ParseSchemaMap(fileContents string) map[string]Schema {
	m := map[string]Schema{}

	schema := Schema{}
	class := ""
	for _, line := range strings.Split(fileContents, "\n") {
		classMatch := classRe.FindStringSubmatch(line)
		if len(classMatch) > 0 {
			if class == "" {
				class = classMatch[1]
				continue
			}

			m[class] = schema
			schema = Schema{}
			class = classMatch[1]
		}

		if class == "" {
			continue
		}

		nameMatch := propNameRe.FindStringSubmatch(line)
		if len(nameMatch) == 0 {
			continue
		}
		name := nameMatch[1]

		typeMatch := pythonModelsFixRe.FindStringSubmatch(line)
		if len(typeMatch) == 0 {
			continue
		}
		t := typeMatch[1]
		description := ""
		for _, match := range paramRe.FindAllStringSubmatch(line, -1) {
			if match[1] == "description" {
				description = match[2]
				break
			}
		}

		schema.Props = append(schema.Props, Parameter{
			Name:        name,
			Type:        t,
			In:          "body",
			Description: description,
		})
	}

	if class != "" {
		m[class] = schema
	}

	return m
}

type ParamResponse struct {
	Parameter Parameter
	Annotated Annotated
}

func (s *SwaggerParser) ParseParameter(operationId string, parameter *openapi3.Parameter) ParamResponse {
	response := ParamResponse{}

	if parameter.Schema.Ref != "" {
		return response
	}

	response.Parameter.Type = fmt.Sprintf("Optional[%s]", parseType(parameter.Schema.Value))
	if parameter.Required {
		response.Annotated = parseAnnotated(operationId, parameter)
		response.Parameter.Type = response.Annotated.Type
	}
	response.Parameter.Name = parameter.Name
	response.Parameter.In = parameter.In
	response.Parameter.Description = parameter.Description
	return response
}

func parseAnnotated(operationId string, parameter *openapi3.Parameter) Annotated {
	return Annotated{
		Name:          parameter.Name,
		OperationId:   operationId,
		PrimitiveType: parseType(parameter.Schema.Value),
		Type:          fmt.Sprintf(`Annotated[%s, Field(description="%s")]`, parseType(parameter.Schema.Value), parameter.Description),
	}
}

func parseType(schema *openapi3.Schema) string {
	schemaType := *schema.Type

	switch schemaType[0] {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "array":
		return fmt.Sprintf("List[%s]", parseType(schema.Items.Value))
	}

	return ""
}

type Schema struct {
	Props []Parameter
}

func (s *SwaggerParser) ParseRequestBody(parsedSchemas map[string]Schema, body *openapi3.RequestBody) []Parameter {
	bodyReq := body.Content.Get("application/json")
	if bodyReq == nil {
		return []Parameter{}
	}
	if bodyReq.Schema.Ref != "" {
		schemaName := strings.Split(bodyReq.Schema.Ref, "/")[3]
		schema, ok := parsedSchemas[schemaName]
		if !ok {
			return []Parameter{}
		}
		return schema.Props
	}
	types := *bodyReq.Schema.Value.Type
	switch types[0] {
	case "array":
		schemaName := strings.Split(bodyReq.Schema.Value.Items.Ref, "/")[3]
		schema, ok := parsedSchemas[schemaName]
		if !ok {
			return []Parameter{}
		}
		return schema.Props
	}

	return []Parameter{}
}

func (s *SwaggerParser) ParseResponse(parsedSchemas map[string]Schema, response *openapi3.Responses) []Parameter {
	resp := response.Status(200)
	if resp == nil {
		return []Parameter{}
	}

	content := resp.Value.Content.Get("application/json")
	if content == nil {
		return []Parameter{}
	}

	if content.Schema.Ref != "" {
		schemaName := strings.Split(content.Schema.Ref, "/")[3]
		return []Parameter{
			{Name: schemaName, Type: schemaName},
		}
	}
	if content.Schema.Value.Type == nil {
		return []Parameter{}
	}
	types := *content.Schema.Value.Type
	switch types[0] {
	case "array":
		schemaName := strings.Split(content.Schema.Value.Items.Ref, "/")[3]
		return []Parameter{
			{Name: schemaName, Type: "List[" + schemaName + "]"},
		}
	}

	return []Parameter{}
}
