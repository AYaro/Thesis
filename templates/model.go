package templates

var Model = `package gen

import (
	"fmt"
	"reflect"
	"time"
	
	"github.com/99designs/gqlgen/graphql"
	"github.com/mitchellh/mapstructure"
)

{{range $object := .Model.Entities}}

	type {{.Name}}ResponseType struct {
		EntityResultType
	}

	type {{.Name}} struct {
	{{range $col := $object.Columns}}
		{{$col.MethodName}} {{$col.GoType}} ` + "`" + `{{$col.ModelTags}}` + "`" + `{{end}}

	{{range $rel := $object.Relationships}}
	{{$rel.MethodName}} {{$rel.GoType}} ` + "`" + `{{$rel.ModelTags}}` + "`" + `
	{{if $rel.Preload}}{{$rel.MethodName}}Preloaded bool ` + "`gorm:\"-\"`" + `{{end}}
	{{end}}
	}


	{{range $interface := $object.Interfaces}}

	func (m *{{$object.Name}}) Is{{$interface}}() {}
	{{end}}

	type {{.Name}}Changes struct {
		{{range $col := $object.Columns}}
		{{$col.MethodName}} {{$col.InputTypeName}}{{end}}
		{{range $rel := $object.Relationships}}{{if $rel.IsToMany}}
		{{$rel.ChangesName}} {{$rel.ChangesType}}{{end}}{{end}}
	}

	{{range $rel := $object.Relationships}}
		{{if and $rel.IsManyToMany $rel.IsMainRelationshipForManyToMany}}
		type {{$rel.ManyToManyObjectName}} struct {
			{{$rel.ForeignKeyDestinationColumn}} string
			{{$rel.InverseRelationship.ForeignKeyDestinationColumn}} string
		}

		func ({{$rel.ManyToManyObjectName}}) TableName() string {
			return TableName("{{$rel.ManyToManyJoinTable}}")
		}
		{{end}}
	{{end}}
{{end}}

func ApplyChanges(changes map[string]interface{}, to interface{}) error {
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:      to,
		ZeroFields:  true,
		ErrorUnused: true,
		TagName:     "json",
		DecodeHook: func(a reflect.Type, b reflect.Type, v interface{}) (interface{}, error) {

			if b == reflect.TypeOf(time.Time{}) {
				switch a.Kind() {
				case reflect.String:
					return time.Parse(time.RFC3339, v.(string))
				case reflect.Int64:
					return time.Unix(0, v.(int64)*int64(time.Millisecond)), nil
				default:
					return v, fmt.Errorf("Unable to parse date from %v", v)
				}
			}

			if reflect.PtrTo(b).Implements(reflect.TypeOf((*graphql.Unmarshaler)(nil)).Elem()) {
				resultType := reflect.New(b)
				result := resultType.MethodByName("UnmarshalGQL").Call([]reflect.Value{reflect.ValueOf(v)})

				err, _ := result[0].Interface().(error)
				return resultType.Elem().Interface(), err
			}

			return v, nil
		},
	})

	if err != nil {
		return err
	}

	return dec.Decode(changes)
}
`
