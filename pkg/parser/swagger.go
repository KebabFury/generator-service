package parser

import (
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
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

	err = cmd.Run()
	if err != nil {
		fmt.Println("Ошибка при выполнении команды:", err)
		//panic(err)
	} else {
		fmt.Println("Команда выполнена успешно.")
	}

	var pythonFile []byte
	if err == nil {
		pythonFile, err = os.ReadFile("./tmp/models.py")
		if err != nil {
			panic(err)
		}
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
	initialSchemas := map[string]Schema{}
	for k, v := range schemas {
		initialSchemas[k] = v
	}
	parseDocument := Document{}
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
				bodyType, bodyParams := s.ParseRequestBody(operation.OperationID, schemas, operation.RequestBody.Value)
				apiCall.BodyType = bodyType
				for _, param := range bodyParams {
					apiCall.Parameters = append(apiCall.Parameters, param)
				}
			}

			ressponse := s.ParseResponse(operation.OperationID, schemas, operation.Responses)
			for _, parameter := range ressponse {
				apiCall.Returns = append(apiCall.Returns, parameter)
			}
			parseDocument.ApiCalls = append(parseDocument.ApiCalls, apiCall)
		}

	}

	modelBlock := ""

	if len(pythonFileContent) > 0 {
		modelBlock = pythonFileContent[getFirstClassRe.FindStringIndex(pythonFileContent)[0]:]
	}
	keys := make([]string, 0, len(schemas))

	for k := range schemas {
		if _, ok := initialSchemas[k]; ok || k == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := len(keys) - 1; i >= 0; i-- {
		modelBlock += schemas[keys[i]].GeneratePythonModel(keys[i])

	}
	parseDocument.Models = template.HTML(modelBlock)
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
	err = os.WriteFile("do.md", []byte(DocumentToMarkdown(&parseDocument)), 0644)
	if err != nil {
		panic(err)
	}
	return &parseDocument
}

var breaksRe = regexp.MustCompile("(?m)(?:\n)|(?:<br>)")
var classRe = regexp.MustCompile(`(?m)^class (.*)\(BaseModel\):$`)
var propNameRe = regexp.MustCompile(`(?m)([a-zA-Z_-]*):`)
var typeRe = regexp.MustCompile(`(?m)(?::\s*([a-zA-Z\-_\]\[0-9]*)\s?)`)
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

		typeMatch := typeRe.FindStringSubmatch(line)

		if len(typeMatch) == 0 {
			continue
		}
		t := typeMatch[1]
		if strings.HasPrefix(t, "Optional[") {
			t = strings.TrimPrefix(t, "Optional[")
			t = t[:len(t)-1]
		}
		description := ""
		for _, match := range paramRe.FindAllStringSubmatch(line, -1) {
			fmt.Println(match)
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
	response.Parameter.Name = fixReservedWords(parameter.Name)
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
	case "boolean":
		return "bool"
	case "array":
		return fmt.Sprintf("List[%s]", parseType(schema.Items.Value))
	}

	return ""
}

type Schema struct {
	Props []Parameter
}

func (s Schema) GeneratePythonModel(name string) string {
	if len(s.Props) == 0 {
		return ""
	}

	m := "class " + name + "(BaseModel):\n"
	for _, prop := range s.Props {
		if prop.Name == "" || prop.Type == "" {
			continue
		}
		typeForField := prop.Type
		if strings.HasPrefix(prop.Type, "Optional[") {
			typeForField = "None"
		}
		m += "\t" + prop.Name + ": " + prop.Type + " = Field(" + typeForField + ", description=" + strconv.Quote(breaksRe.ReplaceAllString(prop.Description, "")) + ")\n"
	}

	return m
}

func (s *SwaggerParser) ParseRequestBody(operationId string, parsedSchemas map[string]Schema, body *openapi3.RequestBody) (string, []Parameter) {
	mimeType := ""
	bodyReq := body.Content.Get("application/json")
	for t, mediaType := range body.Content {
		bodyReq = mediaType
		mimeType = t
		// TODO: add support for multiple content
		break
	}
	if bodyReq == nil {
		return mimeType, []Parameter{}
	}

	s.ParseSchema(operationId, false, "", "Body", bodyReq.Schema, parsedSchemas)

	props := parsedSchemas[operationId+"Body"].Props
	for i := 0; i < len(props); i++ {
		props[i].In = "body"
	}
	return mimeType, props
}

func (s *SwaggerParser) ParseResponse(operationId string, parsedSchemas map[string]Schema, response *openapi3.Responses) []Parameter {
	resp := response.Status(200)
	if resp == nil {
		return []Parameter{}
	}

	content := resp.Value.Content.Get("application/json")
	if content == nil {
		return []Parameter{}
	}

	s.ParseSchema(operationId, false, "", "Response", content.Schema, parsedSchemas)

	return parsedSchemas[operationId+"Response"].Props
}

var caser = cases.Title(language.Und)

func (s *SwaggerParser) ParseSchema(operationId string, required bool, parentPn string, pn string, schema *openapi3.SchemaRef, schemas map[string]Schema) {
	if schema.Ref != "" {
		tmp, ok := schemas[operationId+parentPn]
		if !ok {
			tmp = Schema{}
		}
		description := ""
		if schema.Value != nil {
			description = strings.Trim(strconv.Quote(breaksRe.ReplaceAllString(schema.Value.Description, "")), "\"")
		}
		tmp.Props = append(tmp.Props, Parameter{Name: fixReservedWords(pn), Type: operationId + parentPn + caser.String(pn), Description: description})
		schemas[operationId+parentPn] = tmp
		schemas[operationId+parentPn+caser.String(pn)] = schemas[strings.Split(schema.Ref, "/")[3]]
		return
	}

	schemaType := *schema.Value.Type
	switch schemaType[0] {
	case "object":
		requiredMap := map[string]struct{}{}
		for _, requiredProp := range schema.Value.Required {
			requiredMap[requiredProp] = struct{}{}
		}
		schemas[operationId+parentPn+caser.String(pn)] = Schema{}
		for propName, prop := range schema.Value.Properties {
			if parentPn != "" {
				tmp := schemas[operationId+parentPn]

				typ := "Optional[" + operationId + parentPn + caser.String(pn) + "]"
				// TODO: FIX THIS BUG
				if !required {
					typ = operationId + parentPn + caser.String(pn)
				}
				tmp.Props = append(tmp.Props, Parameter{Name: fixReservedWords(pn), Type: typ, Description: strings.Trim(strconv.Quote(breaksRe.ReplaceAllString(schema.Value.Description, "")), "\"")})
				schemas[operationId+parentPn] = tmp
			}

			_, ok := requiredMap[propName]
			s.ParseSchema(operationId, ok, parentPn+caser.String(pn), propName, prop, schemas)
		}
	case "integer":
		fallthrough
	case "string":
		mainTyp := fmt.Sprintf(`Annotated[%s, Field(%s, description="%s")]`, parseType(schema.Value), parseType(schema.Value), strings.Trim(strconv.Quote(breaksRe.ReplaceAllString(schema.Value.Description, "")), "\""))
		typ := "Optional[" + mainTyp + "]"
		if required {
			typ = mainTyp
		}
		tmp := schemas[operationId+parentPn]
		tmp.Props = append(tmp.Props, Parameter{
			Name: fixReservedWords(pn),
			Type: typ,
		})

		schemas[operationId+parentPn] = tmp
	}
}

var reservedWords = map[string]struct{}{
	"from":   {},
	"import": {},
	"as":     {},
}

func fixReservedWords(s string) string {
	if _, ok := reservedWords[s]; ok {
		return s + "_"
	}
	return s
}
