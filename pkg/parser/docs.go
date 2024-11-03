package parser

import "fmt"

func DocumentToMarkdown(doc *Document) string {
	markdown := ""
	markdown += "# TypeHints Definition\n"
	markdown += string(doc.Types)
	markdown += "# Models definition\n"
	markdown += string(doc.Models)
	markdown += "\n"
	markdown += "# API Calls Documentation\n"

	for _, call := range doc.ApiCalls {
		markdown += fmt.Sprintf("## `%s`\n", call.Name)
		markdown += "// TYPE: " + call.Type + "\n"
		markdown += "// URL: " + call.Url + "\n"
		markdown += "// METHOD: " + call.Method + "\n"
		markdown += "// BODY_TYPE: " + call.BodyType + "\n"
		markdown += "**Description**:\n"
		markdown += call.Description + "\n\n"
		if len(call.Parameters) > 0 {
			markdown += "**Parameters**:\n"
		}
		for _, parameter := range call.Parameters {
			markdown += fmt.Sprintf("- %s (%s): /%s/ %s \n", parameter.Name, parameter.Type, parameter.In, parameter.Description)
		}

		markdown += "\n"

		if len(call.Returns) > 0 {
			markdown += "**Returns**:\n"
		}

		for _, parameter := range call.Returns {
			markdown += fmt.Sprintf("- %s (%s)\n", parameter.Name, parameter.Type)
		}

		markdown += "\n"
		markdown += "\n"
	}

	return markdown
}
