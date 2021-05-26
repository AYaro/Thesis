package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/inflection"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/iancoleman/strcase"
)

// Object ...
type Object struct {
	Definition *ast.ObjectDefinition
	Model      *Model
	Extension  *ObjectExtension
}

// Name ...
func (o *Object) Name() string {
	return o.Definition.Name.Value
}

// PluralName ...
func (o *Object) PluralName() string {
	return inflection.Plural(o.Name())
}

// LowerName ...
func (o *Object) LowerName() string {
	return strcase.ToLowerCamel(o.Definition.Name.Value)
}

// TableName ...
func (o *Object) TableName() string {
	return strcase.ToSnake(inflection.Plural(o.LowerName()))
}

// HasColumn ...
func (o *Object) HasColumn(name string) bool {
	return o.Column(name) != nil
}

// HasField ...
func (o *Object) HasField(name string) bool {
	return o.Field(name) != nil
}

// Column ...
func (o *Object) Column(name string) *ObjectField {
	for _, f := range o.Definition.Fields {
		if f.Name.Value == name {
			field := &ObjectField{f, o}
			if field.IsColumn() {
				return field
			} else {
				return nil
			}
		}
	}
	return nil
}

// Columns ...
func (o *Object) Columns() []ObjectField {
	columns := []ObjectField{}
	for _, f := range o.Fields() {
		if f.IsColumn() {
			columns = append(columns, f)
		}
	}
	return columns
}

// Field ...
func (o *Object) Field(name string) *ObjectField {
	for _, f := range o.Definition.Fields {
		if f.Name.Value == name {
			return &ObjectField{f, o}
		}
	}
	return nil
}

// Fields ...
func (o *Object) Fields() []ObjectField {
	fields := []ObjectField{}
	for _, f := range o.Definition.Fields {
		fields = append(fields, ObjectField{f, o})
	}
	return fields
}

// HasEmbeddedField ...
func (o *Object) HasEmbeddedField() bool {
	for _, f := range o.Fields() {
		if f.IsEmbedded() {
			return true
		}
	}
	return false
}

// HasReadonlyColumns ...
func (o *Object) HasReadonlyColumns() bool {
	for _, c := range o.Columns() {
		if c.IsReadonlyType() {
			return true
		}
	}
	return false
}

// IsToManyColumn ...
func (o *Object) IsToManyColumn(c ObjectField) bool {
	if c.Obj.Name() != o.Name() {
		return false
	}
	return o.HasRelationship(strings.TrimSuffix(c.Name(), "Ids"))
}

// Relationships ...
func (o *Object) Relationships() []*ObjectRelationship {
	relationships := []*ObjectRelationship{}
	for _, f := range o.Definition.Fields {
		if o.isRelationship(f) {
			relationships = append(relationships, &ObjectRelationship{f, o})
		}
	}
	return relationships
}

// Relationship ...
func (o *Object) Relationship(name string) *ObjectRelationship {
	for _, rel := range o.Relationships() {
		if rel.Name() == name {
			return rel
		}
	}
	panic(fmt.Sprintf("relationship %s->%s not found", o.Name(), name))
}

// HasAnyRelationships ...
func (o *Object) HasAnyRelationships() bool {
	return len(o.Relationships()) > 0
}

// HasRelationship ....
func (o *Object) HasRelationship(name string) bool {
	for _, rel := range o.Relationships() {
		if rel.Name() == name {
			return true
		}
	}
	return false
}

// NeedsQueryResolver ....
func (o *Object) NeedsQueryResolver() bool {
	return o.HasAnyRelationships() || o.HasEmbeddedField() || o.Model.HasObjectExtension(o.Name())
}

// PreloadableRelationships ...
func (o *Object) PreloadableRelationships() []*ObjectRelationship {
	result := []*ObjectRelationship{}
	for _, r := range o.Relationships() {
		if r.Preload() {
			result = append(result, r)
		}
	}
	return result
}

// HasPreloadableRelationships ...
func (o *Object) HasPreloadableRelationships() bool {
	return len(o.PreloadableRelationships()) > 0
}

// Directive ...
func (o *Object) Directive(name string) *ast.Directive {
	for _, d := range o.Definition.Directives {
		if d.Name.Value == name {
			return d
		}
	}
	return nil
}

// HasDirective ...
func (o *Object) HasDirective(name string) bool {
	return o.Directive(name) != nil
}

func (o *Object) isRelationship(f *ast.FieldDefinition) bool {
	for _, d := range f.Directives {
		if d != nil && d.Name.Value == "relationship" {
			return true
		}
	}
	return false
}

// IsExtended ....
func (o *Object) IsExtended() bool {
	return o.Extension != nil
}

// Interfaces ...
func (o *Object) Interfaces() []string {
	interfaces := []string{}
	for _, item := range o.Definition.Interfaces {
		interfaces = append(interfaces, item.Name.Value)
	}
	return interfaces
}

func (o *Object) HasAggregableColumn() bool {
	for _, column := range o.Columns() {
		if column.IsAggregable() {
			return true
		}
	}
	return false
}

// AggregationsByField ...
func (o *Object) AggregationsByField() (res map[string]*ObjectFieldAggregation) {
	res = map[string]*ObjectFieldAggregation{}
	for _, column := range o.Columns() {
		if column.IsAggregable() {
			for _, agg := range column.Aggregations() {
				val := agg
				res[agg.FieldName()] = &val
			}
		}
	}
	return
}

type ObjectFieldAggregation struct {
	Field string
	Name  string
	Type  ast.Type
}

// FieldName ...
func (a *ObjectFieldAggregation) FieldName() string {
	return a.Field + a.Name
}

// SQLColumn ...
func (a *ObjectFieldAggregation) SQLColumn() string {
	return fmt.Sprintf("%s(%s) as %s_%s", strings.ToUpper(a.Name), a.Field, a.Field, strings.ToLower(a.Name))
}

// IsAggregable ...
func (o *ObjectField) IsAggregable() bool {
	return o.IsString() || o.IsNumeric()
}

// Aggregations ...
func (o *ObjectField) Aggregations() []ObjectFieldAggregation {
	res := []ObjectFieldAggregation{
		{Field: o.Name(), Name: "Min", Type: o.Def.Type},
		{Field: o.Name(), Name: "Max", Type: o.Def.Type},
	}
	if o.IsNumeric() {
		res = append(res,
			ObjectFieldAggregation{Field: o.Name(), Name: "Avg", Type: namedType("Float")},
			ObjectFieldAggregation{Field: o.Name(), Name: "Sum", Type: o.Def.Type},
		)
	}
	return res
}

// ObjectExtension ...
type ObjectExtension struct {
	Def    *ast.TypeExtensionDefinition
	Model  *Model
	Object *Object
}

// IsFederatedType ...
func (oe *ObjectExtension) IsFederatedType() bool {
	return oe.Object.IsFederatedType()
}

// ExtendsLocalObject ...
func (oe *ObjectExtension) ExtendsLocalObject() bool {
	return oe.Model.HasObject(oe.Object.Name())
}

// IsExternal ...
func (oe *ObjectExtension) HasAnyNonExternalField() bool {
	for _, f := range oe.Object.Fields() {
		if !f.IsExternal() {
			return true
		}
	}
	return false
}
