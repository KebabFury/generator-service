package parser

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

type Document struct {
	Name        string        `json:"name,omitempty"`
	Types       template.HTML `json:"types,omitempty"`
	Models      template.HTML `json:"models,omitempty"`
	ApiCalls    []ApiCall     `json:"api_calls,omitempty"`
	Provider    string
	Description string
}

const ApiCallTypeCustom = "custom"

type ApiCall struct {
	Type           string      `json:"type,omitempty"`
	Method         string      `json:"method,omitempty"`
	Url            string      `json:"url,omitempty"`
	BodyType       string      `json:"bodyType,omitempty"`
	Name           string      `json:"name,omitempty" json:"name,omitempty"`
	Description    string      `json:"description,omitempty" json:"description,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty" json:"parameters,omitempty"`
	Returns        []Parameter `json:"returns,omitempty" json:"returns,omitempty"`
	FormatedReturn string      `json:"formated_return,omitempty"`
}

func (c ApiCall) PathParameters() string {
	var s []string
	for _, param := range c.Parameters {
		if param.In == "path" {
			s = append(s, fmt.Sprintf(`\"%s\": {%s}`, param.Name, param.Name))
		}
	}
	return strings.Join(s, ", ")
}

func (c ApiCall) BodyParameters() string {
	var s []string
	for _, param := range c.Parameters {
		if param.In == "body" {
			s = append(s, fmt.Sprintf(`\"%s\": {%s}`, param.Name, param.Name))
		}
	}
	return strings.Join(s, ", ")
}

func (c ApiCall) QueryParameters() string {
	var s []string
	for _, param := range c.Parameters {
		if param.In == "query" {
			s = append(s, fmt.Sprintf(`\"%s\": {%s}`, param.Name, param.Name))
		}
	}
	return strings.Join(s, ", ")
}

type Parameter struct {
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	In          string
	Description string
}

func (p Parameter) TypeQuote() string {
	return strings.Trim(strconv.Quote(p.Type), "\"")
}

func (p Parameter) TypeHtml() template.HTML {
	return template.HTML(p.Type)
}

var argumentRe = regexp.MustCompile(`(?m)^-\s(.+)\s\((.+)\):\s(?:/([a-z]*)/)?.*`)
var returnRe = regexp.MustCompile(`(?m)^-\s(.+)\s\((.+)\)`)
var typeHintsRe = regexp.MustCompile(`(?mU)# TypeHints Definition\n((?:.|\n)*)#`)
var modelsRe = regexp.MustCompile(`(?mU)# Models definition\n((?:.|\n)*)#`)

func ParseDocument(documentStr string) *Document {
	document := &Document{}
	lines := strings.Split(documentStr, "\n")

	typingHints := typeHintsRe.FindStringSubmatch(documentStr)
	if len(typingHints) > 0 {
		document.Types = template.HTML(typingHints[1])
	}

	models := modelsRe.FindStringSubmatch(documentStr)

	document.Models = template.HTML(models[1])

	var calls []ApiCall
	apiCall := ApiCall{}
	key := "title"
	for _, line := range append(lines, "##") {
		if strings.HasPrefix(line, "##") {
			if apiCall.Name != "" {
				if len(apiCall.Returns) > 0 {
					apiCall.FormatedReturn = apiCall.Returns[0].Type
				}
				if len(apiCall.Returns) > 1 {
					joined := ""
					for i, parameter := range apiCall.Returns {
						if i == 0 {
							joined = parameter.Type
							continue
						}
						joined += "," + parameter.Type
					}
					apiCall.FormatedReturn = "Union[" + joined + "]"
				}

				calls = append(calls, apiCall)
				apiCall = ApiCall{}
			}
			key = "title"
			line = strings.TrimPrefix(line, "## ")
			apiCall.Name = strings.ReplaceAll(line, "`", "")
			continue
		}
		if strings.HasPrefix(line, "// URL: ") {
			apiCall.Url = strings.TrimPrefix(line, "// URL: ")
			continue
		}
		if strings.HasPrefix(line, "// METHOD: ") {
			apiCall.Method = strings.TrimPrefix(line, "// METHOD: ")
			continue
		}

		if strings.HasPrefix(line, "// BODY_TYPE: ") {
			apiCall.BodyType = strings.TrimPrefix(line, "// BODY_TYPE: ")
			continue
		}
		if strings.HasPrefix(line, "// TYPE: ") {
			apiCall.Type = strings.TrimPrefix(line, "// TYPE: ")
			continue
		}
		if line == "**Description**:" {
			key = "description"
			continue
		}

		if line == "**Parameters**:" {
			key = "parameters"
			continue
		}

		if line == "**Returns**:" {
			key = "returns"
			continue
		}

		switch key {
		case "parameters":
			sub := argumentRe.FindStringSubmatch(line)
			if len(sub) == 0 {
				break
			}

			apiCall.Parameters = append(apiCall.Parameters, Parameter{
				Name: sub[1],
				Type: sub[2],
				In:   sub[3],
			})
		case "description":
			apiCall.Description += line
		case "returns":
			sub := returnRe.FindStringSubmatch(line)
			if len(sub) == 0 {
				break
			}

			apiCall.Returns = append(apiCall.Returns, Parameter{
				Name: sub[1],
				Type: sub[2],
			})
		}

	}
	document.ApiCalls = calls

	return document
}
