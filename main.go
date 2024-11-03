package main

import (
	"github.com/KebabFury/generator-service/pkg/parser"
	"os"
)

func main() {
	input, err := os.ReadFile("swagger.json")
	if err != nil {
		panic(err)
	}

	swagParser := parser.NewSwaggerParser()
	swagParser.Parse(input)
	//fmt.Println(err)
	//fmt.Printf("Schema for %s:\n%s\n\n", "asd", string(schemaJSON))
}
