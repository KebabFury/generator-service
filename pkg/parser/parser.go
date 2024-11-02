package parser

import (
	"regexp"
	"strings"
)

type Document struct {
	Name     string    `json:"name,omitempty"`
	Types    string    `json:"types,omitempty"`
	Models   string    `json:"models,omitempty"`
	ApiCalls []ApiCall `json:"api_calls,omitempty"`
}

type ApiCall struct {
	Name           string      `json:"name,omitempty"`
	Description    string      `json:"description,omitempty"`
	Parameters     []Parameter `json:"parameters,omitempty"`
	Returns        []Parameter `json:"returns,omitempty"`
	FormatedReturn string
}

type Parameter struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

var argumentRe = regexp.MustCompile(`(?m)^-\s(.+)\s\((.+)\):\s.*`)
var returnRe = regexp.MustCompile(`(?m)^-\s(.+)\s\((.+)\)`)
var typeHintsRe = regexp.MustCompile(`(?mU)# TypeHints Definition\n((?:.|\n)*)#`)
var modelsRe = regexp.MustCompile(`(?mU)# Models definition\n((?:.|\n)*)#`)

func ParseDocument(documentStr string) *Document {
	document := &Document{}
	lines := strings.Split(documentStr, "\n")

	typingHints := typeHintsRe.FindStringSubmatch(documentStr)
	document.Types = typingHints[1]

	models := modelsRe.FindStringSubmatch(documentStr)

	document.Models = models[1]

	var calls []ApiCall
	apiCall := ApiCall{}
	key := "title"
	for _, line := range append(lines, "##") {
		if strings.HasPrefix(line, "##") {
			if apiCall.Name != "" {
				apiCall.FormatedReturn = apiCall.Returns[0].Type
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
