package model

import (
	"github.com/graphql-go/graphql/language/parser"
)

func Parse(input string) (Model, error) {
	var model Model

	astDocument, err := parser.Parse(parser.ParseParams{
		Options: parser.ParseOptions{
			NoLocation: true,
		},
		Source: input,
	})
	if err != nil {
		return model, err
	}

	model = Model{astDocument}
	return model, nil
}
