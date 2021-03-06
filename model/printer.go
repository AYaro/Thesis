package model

import (
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/printer"
)

func filterDirective(ds []*ast.Directive, name string) []*ast.Directive {
	res := []*ast.Directive{}
	for _, d := range ds {
		if d.Name.Value != name {
			res = append(res, d)
		}
	}
	return res
}

func PrintSchema(model Model) (string, error) {
	for _, o := range model.Objects() {
		fields := []*ast.FieldDefinition{}
		for _, f := range o.Definition.Fields {
			f.Directives = filterDirective(f.Directives, "relationship")
			f.Directives = filterDirective(f.Directives, "column")
			fields = append(fields, f)
		}
		o.Definition.Fields = fields
		o.Definition.Directives = filterDirective(o.Definition.Directives, "entity")
	}

	printed := printer.Print(model.Doc)
	printedString, _ := printed.(string)

	return printedString, nil
}
