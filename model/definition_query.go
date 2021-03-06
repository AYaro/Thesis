package model

import (
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

func queryDefinition(m *Model) *ast.ObjectDefinition {
	fields := []*ast.FieldDefinition{}

	if m.HasFederatedTypes() {
		fields = append(fields, createFederationEntitiesQueryField())
	}

	for _, obj := range m.Entities() {
		fields = append(fields, fetchFieldDefinition(obj), listFieldDefinition(obj))
	}
	return &ast.ObjectDefinition{
		Kind: kinds.ObjectDefinition,
		Name: &ast.Name{
			Kind:  kinds.Name,
			Value: "Query",
		},
		Fields: fields,
	}
}

func fetchFieldDefinition(obj Object) *ast.FieldDefinition {
	return &ast.FieldDefinition{
		Kind: kinds.FieldDefinition,
		Name: nameNode(inflection.Singular(strcase.ToLowerCamel(obj.Name()))),
		Type: namedType(obj.Name()),
		Arguments: []*ast.InputValueDefinition{
			{
				Kind:        kinds.InputValueDefinition,
				Name:        nameNode("id"),
				Description: &ast.StringValue{Kind: kinds.StringValue, Value: "Search by ID"},
				Type:        namedType("ID"),
			},
			{
				Kind: kinds.InputValueDefinition,
				Name: nameNode("filter"),
				Type: namedType(obj.Name() + "FilterType"),
			},
		},
	}
}

func listFieldDefinition(obj Object) *ast.FieldDefinition {
	return listFieldResultTypeDefinition(obj, inflection.Plural(strcase.ToLowerCamel(obj.Name())))
}
func listFieldResultTypeDefinition(obj Object, name string) *ast.FieldDefinition {
	return &ast.FieldDefinition{
		Kind: kinds.FieldDefinition,
		Name: nameNode(name),
		Type: nonNull(namedType(obj.Name() + "ResultType")),
		Arguments: []*ast.InputValueDefinition{
			{
				Kind: kinds.InputValueDefinition,
				Name: nameNode("offset"),
				Type: namedType("Int"),
			},
			{
				Kind:         kinds.InputValueDefinition,
				Name:         nameNode("limit"),
				DefaultValue: &ast.IntValue{Kind: kinds.IntValue, Value: "30"},
				Type:         namedType("Int"),
			},
			{
				Kind: kinds.InputValueDefinition,
				Name: nameNode("sort"),
				Type: listType(nonNull(namedType(obj.Name() + "SortType"))),
			},
			{
				Kind: kinds.InputValueDefinition,
				Name: nameNode("filter"),
				Type: namedType(obj.Name() + "FilterType"),
			},
		},
	}
}
