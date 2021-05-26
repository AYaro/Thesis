package model

import (
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func EnrichEntites(model *Model) error {
	// Cоздаем вспомогательные поля createdAt, updatedAt, id
	createdAt := defineColumn("createdAt", "Time", true)
	updatedAt := defineColumn("updatedAt", "Time", false)
	id := defineColumn("id", "ID", true)

	// Получаем объекты c директивой entity
	entities := model.Entities()

	// Получаем объекты c директивой entity
	for _, object := range entities {
		object.Definition.Fields = append(
			[]*ast.FieldDefinition{id},
			object.Definition.Fields...,
		)

		// Создаем вспомогательные поля для связи между объектами
		for _, relation := range object.Relationships() {
			if relation.IsToOne() {
				object.Definition.Fields = append(
					object.Definition.Fields,
					defineColumn(relation.Name()+"Id", "ID", false),
				)
			}
		}
		object.Definition.Fields = append(
			object.Definition.Fields,
			updatedAt,
			createdAt)
	}
	return nil
}

// EnrichModel ...
func EnrichModel(m *Model) error {
	definitions := []ast.Node{}
	for _, o := range m.Entities() {
		for _, rel := range o.Relationships() {
			if rel.IsToMany() {
				o.Definition.Fields = append(o.Definition.Fields, defineColumnWithType(rel.Name()+"Ids", nonNull(listType(nonNull(namedType("ID"))))))
				o.Definition.Fields = append(o.Definition.Fields, listFieldResultTypeDefinition(*rel.Target(), rel.Name()+"Connection"))
			}
		}
		definitions = append(definitions, createObjectDefinition(o), updateObjectDefinition(o), createObjectSortType(o), createObjectFilterType(o))
		definitions = append(definitions, objectResultTypeDefinition(&o))
		if o.HasAggregableColumn() {
			definitions = append(definitions, objectResultTypeAggregationsDefinition(&o))
		}
	}

	for _, o := range m.EmbeddedObjects() {
		def := embeddedObjectDefinition(o)
		definitions = append(definitions, def)
	}

	schemaHeaderNodes := []ast.Node{
		schemaDefinition(m),
		queryDefinition(m),
		mutationDefinition(m),
		createObjectSortEnum(),
	}
	m.Doc.Definitions = append(schemaHeaderNodes, m.Doc.Definitions...)
	m.Doc.Definitions = append(m.Doc.Definitions, definitions...)
	m.Doc.Definitions = append(m.Doc.Definitions, createFederationServiceObject())

	return nil
}

func BuildFederatedModel(m *Model) error {
	if m.HasFederatedTypes() {
		m.Doc.Definitions = append(m.Doc.Definitions, createFederationEntityUnion(m))
	}

	for _, e := range m.ObjectExtensions() {
		if e.IsFederatedType() {
			m.Doc.Definitions = append(m.Doc.Definitions, getObjectDefinitionFromFederationExtension(e.Object.Definition))
			m.RemoveObjectExtension(&e)
		}
	}

	for _, obj := range m.Objects() {
		if obj.HasDirective("key") {
			obj.Definition.Directives = filterDirective(obj.Definition.Directives, "key")
		}
	}

	return nil
}

func defineColumn(columnName, columnType string, isNonNull bool) *ast.FieldDefinition {
	t := namedType(columnType)
	if isNonNull {
		t = nonNull(t)
	}
	return defineColumnWithType(columnName, t)
}
func defineColumnWithType(fieldName string, t ast.Type) *ast.FieldDefinition {
	return &ast.FieldDefinition{
		Name: nameNode(fieldName),
		Kind: kinds.FieldDefinition,
		Type: t,
		Directives: []*ast.Directive{
			{
				Kind: kinds.Directive,
				Name: nameNode("column"),
			},
		},
	}
}

func schemaDefinition(m *Model) *ast.SchemaDefinition {
	return &ast.SchemaDefinition{
		Kind: kinds.SchemaDefinition,
		OperationTypes: []*ast.OperationTypeDefinition{
			{
				Operation: "query",
				Kind:      kinds.OperationTypeDefinition,
				Type: &ast.Named{
					Kind: kinds.Named,
					Name: &ast.Name{
						Kind:  kinds.Name,
						Value: "Query",
					},
				},
			},
			{
				Operation: "mutation",
				Kind:      kinds.OperationTypeDefinition,
				Type: &ast.Named{
					Kind: kinds.Named,
					Name: &ast.Name{
						Kind:  kinds.Name,
						Value: "Mutation",
					},
				},
			},
		},
	}
}
