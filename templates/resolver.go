package templates

var ResolverCore = `package gen

import (
	"context"
	"time"
	
	"github.com/graph-gophers/dataloader"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gofrs/uuid"
	"github.com/vektah/gqlparser/ast"
)

type ResolverHandlers struct {
	{{range $obj := .Model.ObjectEntities}}
		Delete{{$obj.Name}} func(ctx context.Context, r *GeneratedResolver, id string) (item *{{$obj.Name}}, err error)
		DeleteAll{{$obj.PluralName}} func (ctx context.Context, r *GeneratedResolver) (bool, error) 
		Create{{$obj.Name}} func (ctx context.Context, r *GeneratedResolver, input map[string]interface{}) (item *{{$obj.Name}}, err error)
		Update{{$obj.Name}} func(ctx context.Context, r *GeneratedResolver, id string, input map[string]interface{}) (item *{{$obj.Name}}, err error)
		Query{{$obj.Name}} func (ctx context.Context, r *GeneratedResolver, opts Query{{$obj.Name}}HandlerOptions) (*{{$obj.Name}}, error)
		Query{{$obj.PluralName}} func (ctx context.Context, r *GeneratedResolver, opts Query{{$obj.PluralName}}HandlerOptions) (*{{$obj.Name}}ResultType, error)
		{{range $col := $obj.Fields}}{{if $col.NeedsQueryResolver}}
			{{$obj.Name}}{{$col.MethodName}} func (ctx context.Context,r *GeneratedResolver, obj *{{$obj.Name}}) (res {{$col.GoResultType}}, err error)
		{{end}}{{end}}
		{{range $rel := $obj.Relationships}}
			{{$obj.Name}}{{$rel.MethodName}} func (ctx context.Context,r *GeneratedResolver, obj *{{$obj.Name}}) (res {{$rel.ReturnType}}, err error)
		{{end}}
	{{end}}
}

func DefaultResolverHandlers() ResolverHandlers {
	handlers := ResolverHandlers{
		{{range $obj := .Model.ObjectEntities}}
			Delete{{$obj.Name}}: Delete{{$obj.Name}}Handler,
			DeleteAll{{$obj.PluralName}}: DeleteAll{{$obj.PluralName}}Handler,
			Create{{$obj.Name}}: Create{{$obj.Name}}Handler,
			Update{{$obj.Name}}: Update{{$obj.Name}}Handler,
			Query{{$obj.Name}}: Query{{$obj.Name}}Handler,
			Query{{$obj.PluralName}}: Query{{$obj.PluralName}}Handler,
			{{range $col := $obj.Fields}}{{if $col.NeedsQueryResolver}}
				{{$obj.Name}}{{$col.MethodName}}: {{$obj.Name}}{{$col.MethodName}}Handler,
			{{end}}{{end}}
			{{range $rel := $obj.Relationships}}
				{{$obj.Name}}{{$rel.MethodName}}: {{$obj.Name}}{{$rel.MethodName}}Handler,
			{{end}}
		{{end}}
	}
	return handlers
}

type GeneratedResolver struct {
	Handlers ResolverHandlers
	db *DB
}

func NewGeneratedResolver(handlers ResolverHandlers, db *DB) *GeneratedResolver {
	return &GeneratedResolver{Handlers: handlers, db: db}
}

func (r *GeneratedResolver) GetDB(ctx context.Context) *gorm.DB {
	db, _ := ctx.Value(KeyMutationTransaction).(*gorm.DB)
	if db == nil {
		db = r.db.Query()
	}
	return db
}
`

var ResolverGen = `package src

func NewResolver(db *gen.DB) *Resolver {
	handlers := gen.DefaultResolutionHandlers()
	return &Resolver{gen.NewGeneratedResolver(handlers, db)}
}

type Resolver struct {
	*gen.GeneratedResolver
}

type MutationResolver struct {
	*gen.GeneratedMutationResolver
}

func (r * MutationResolver)BeginTransaction(ctx context.Context,fn func(context.Context) error) error {
	ctx = gen.EnrichContextWithMutations(ctx, r.GeneratedResolver)
	err := fn(ctx)
	if err!=nil{
		tx := r.GeneratedResolver.GetDB(ctx)
		tx.Rollback()
		return err
	}
	return gen.FinishMutationContext(ctx, r.GeneratedResolver)
}

type QueryResolver struct {
	*gen.GeneratedQueryResolver
}

func (r *Resolver) Mutation() gen.MutationResolver {
	return &MutationResolver{&gen.GeneratedMutationResolver{r.GeneratedResolver}}
}

func (r *Resolver) Query() gen.QueryResolver {
	return &QueryResolver{&gen.GeneratedQueryResolver{r.GeneratedResolver}}
}


{{range $obj := .Model.ObjectEntities}}
	// {{$obj.Name}}ResultTypeResolver ...
	type {{$obj.Name}}ResultTypeResolver struct {
		*gen.Generated{{$obj.Name}}ResultTypeResolver
	}

	// {{$obj.Name}}ResultType ...
	func (r *Resolver) {{$obj.Name}}ResultType() gen.{{$obj.Name}}ResultTypeResolver {
		return &{{$obj.Name}}ResultTypeResolver{&gen.Generated{{$obj.Name}}ResultTypeResolver{r.GeneratedResolver}}
	}
	{{if $obj.NeedsQueryResolver}}
		// {{$obj.Name}}Resolver ...
		type {{$obj.Name}}Resolver struct {
			*gen.Generated{{$obj.Name}}Resolver
		}

		// {{$obj.Name}} ...
		func (r *Resolver) {{$obj.Name}}() gen.{{$obj.Name}}Resolver {
			return &{{$obj.Name}}Resolver{&gen.Generated{{$obj.Name}}Resolver{r.GeneratedResolver}}
		}
	{{end}}
{{end}}
{{range $ext := .Model.ObjectExtensions}}
	{{$obj := $ext.Object}}
	{{if not $ext.ExtendsLocalObject}}
		// {{$obj.Name}}Resolver ...
		type {{$obj.Name}}Resolver struct {
			*gen.Generated{{$obj.Name}}Resolver
		}
		{{if or $obj.HasAnyRelationships $obj.HasReadonlyColumns $ext.HasAnyNonExternalField}}
			// {{$obj.Name}} ...
			func (r *Resolver) {{$obj.Name}}() gen.{{$obj.Name}}Resolver {
				return &{{$obj.Name}}Resolver{&gen.Generated{{$obj.Name}}Resolver{r.GeneratedResolver}}
			}
		{{end}}
	{{end}}
{{end}}
`

var ResolverMutations = `package gen

import (
	"context"
	"os"
	"time"
	
	"github.com/graph-gophers/dataloader"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gofrs/uuid"
	"github.com/vektah/gqlparser/ast"
)

type GeneratedMutationResolver struct{ *GeneratedResolver }

func FinishMutationContext(ctx context.Context, r *GeneratedResolver) (err error) {
	tx := r.GetDB(ctx)
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

func RollbackMutationContext(ctx context.Context, r *GeneratedResolver) error {
	tx := r.GetDB(ctx)
	return tx.Rollback().Error
}


{{range $obj := .Model.ObjectEntities}}
	func (r *GeneratedMutationResolver) Create{{$obj.Name}}(ctx context.Context, input map[string]interface{}) (item *{{$obj.Name}}, err error) {
		ctx = EnrichContextWithMutations(ctx, r.GeneratedResolver)
		item, err = r.Handlers.Create{{$obj.Name}}(ctx, r.GeneratedResolver, input)
		if err!=nil{
			return
		}
		err = FinishMutationContext(ctx, r.GeneratedResolver)
		return
	}

	func Create{{$obj.Name}}Handler(ctx context.Context, r *GeneratedResolver, input map[string]interface{}) (item *{{$obj.Name}}, err error) {
		principalID := GetPrincipalIDFromContext(ctx)
		now := time.Now()
		item = &{{$obj.Name}}{ID: uuid.Must(uuid.NewV4()).String(), CreatedAt: now, CreatedBy: principalID}
		tx := r.GetDB(ctx)
		
		var changes {{$obj.Name}}Changes
		err = ApplyChanges(input, &changes)
		if err != nil {
			tx.Rollback()
			return 
		}

		{{range $col := .Columns}}{{if $col.IsCreatable}}
			{{if $col.IsEmbeddedColumn}}
				if _, ok := input["{{$col.Name}}"]; ok {
					_value,_err := json.Marshal(changes.{{$col.MethodName}})
					if _err != nil {
						err = _err
						return
					}
					strval := string(_value)
					value := {{if $col.IsOptional}}&{{end}}strval
					if item.{{$col.MethodName}} != value {{if $col.IsOptional}}&& (item.{{$col.MethodName}} == nil || value == nil || *item.{{$col.MethodName}} != *value){{end}} { 
						item.{{$col.MethodName}} = value
					}
				}
			{{else}}
				if _, ok := input["{{$col.Name}}"]; ok && (item.{{$col.MethodName}} != changes.{{$col.MethodName}}){{if $col.IsOptional}} && (item.{{$col.MethodName}} == nil || changes.{{$col.MethodName}} == nil || *item.{{$col.MethodName}} != *changes.{{$col.MethodName}}){{end}} {
					item.{{$col.MethodName}} = changes.{{$col.MethodName}}
				}
			{{end}}
		{{end}}{{end}}
		
		err = tx.Create(item).Error
		if err != nil {
			tx.Rollback()
			return
		}
		
		{{range $rel := $obj.Relationships}}
			{{if not $rel.Target.IsExtended}}
				{{if $rel.IsManyToMany}}
					if ids,exists:=input["{{$rel.Name}}Ids"]; exists {
						items := []{{$rel.TargetType}}{}
						err = tx.Find(&items, "id IN (?)", ids).Error
						if err != nil {
							return
						}
						association := tx.Model(&item).Association("{{$rel.MethodName}}")
						err = association.Replace(items).Error
						if err != nil {
							return
						}
					}
				{{else if $rel.IsToMany}}
					if ids,exists:=input["{{$rel.Name}}Ids"]; exists {
						err = tx.Model({{$rel.TargetType}}{}).Where("id IN (?)", ids).Update("{{$rel.ForeignKeyDestinationColumn}}", item.ID).Error
						if err != nil {
							return
						}
					}
				{{end}}
			{{end}}
		{{end}}

		return 
	}

	func (r *GeneratedMutationResolver) Update{{$obj.Name}}(ctx context.Context, id string, input map[string]interface{}) (item *{{$obj.Name}}, err error) {
		ctx = EnrichContextWithMutations(ctx, r.GeneratedResolver)
		item,err = r.Handlers.Update{{$obj.Name}}(ctx, r.GeneratedResolver, id, input)
		if err!=nil{
			RollbackMutationContext(ctx, r.GeneratedResolver)
			return
		}
		err = FinishMutationContext(ctx, r.GeneratedResolver)
		return
	}

	func Update{{$obj.Name}}Handler(ctx context.Context, r *GeneratedResolver, id string, input map[string]interface{}) (item *{{$obj.Name}}, err error) {
		principalID := GetPrincipalIDFromContext(ctx)
		item = &{{$obj.Name}}{}
		now := time.Now()
		tx := r.GetDB(ctx)

		var changes {{$obj.Name}}Changes
		err = ApplyChanges(input, &changes)
		if err != nil {
			tx.Rollback()
			return 
		}

		err = GetItem(ctx, tx, item, &id)
		if err != nil {
			tx.Rollback()
			return 
		}

		item.UpdatedBy = principalID

		{{range $col := .Columns}}{{if $col.IsUpdatable}}
			{{if $col.IsEmbeddedColumn}}
				if _, ok := input["{{$col.Name}}"]; ok {
					_value,_err := json.Marshal(changes.{{$col.MethodName}})
					if _err != nil {
						err = _err
						return
					}
					if _err!=nil {
						err = _err
						return
					}
					strval := string(_value)
					value := {{if $col.IsOptional}}&{{end}}strval
					if item.{{$col.MethodName}} != value {{if $col.IsOptional}}&& (item.{{$col.MethodName}} == nil || value == nil || *item.{{$col.MethodName}} != *value){{end}} { 
						item.{{$col.MethodName}} = value
					}
				}
			{{else}}
				if _, ok := input["{{$col.Name}}"]; ok && (item.{{$col.MethodName}} != changes.{{$col.MethodName}}){{if $col.IsOptional}} && (item.{{$col.MethodName}} == nil || changes.{{$col.MethodName}} == nil || *item.{{$col.MethodName}} != *changes.{{$col.MethodName}}){{end}} {
					item.{{$col.MethodName}} = changes.{{$col.MethodName}}
				}
			{{end}}
		{{end}}
		{{end}}
		
		err = tx.Save(item).Error
		if err != nil {
			tx.Rollback()
			return
		}

		{{range $rel := $obj.Relationships}}
		{{if $rel.IsToMany}}{{if not $rel.Target.IsExtended}}
			if ids,exists:=input["{{$rel.Name}}Ids"]; exists {
				items := []{{$rel.TargetType}}{}
				tx.Find(&items, "id IN (?)", ids)
				association := tx.Model(&item).Association("{{$rel.MethodName}}")
				association.Replace(items)
			}
		{{end}}{{end}}
		{{end}}

		return 
	}

	func (r *GeneratedMutationResolver) Delete{{$obj.Name}}(ctx context.Context, id string) (item *{{$obj.Name}}, err error) {
		ctx = EnrichContextWithMutations(ctx, r.GeneratedResolver)
		item,err = r.Handlers.Delete{{$obj.Name}}(ctx, r.GeneratedResolver, id)
		if err!=nil{
			RollbackMutationContext(ctx, r.GeneratedResolver)
			return
		}
		err = FinishMutationContext(ctx, r.GeneratedResolver)
		return
	}

	func Delete{{$obj.Name}}Handler(ctx context.Context, r *GeneratedResolver, id string) (item *{{$obj.Name}}, err error) {
		principalID := GetPrincipalIDFromContext(ctx)
		item = &{{$obj.Name}}{}
		now := time.Now()
		tx := r.GetDB(ctx)

		err = GetItem(ctx, tx, item, &id)
		if err != nil {
			tx.Rollback()
			return 
		}

		err = tx.Delete(item, TableName("{{$obj.TableName}}") + ".id = ?", id).Error
		if err != nil {
			tx.Rollback()
			return
		}

		return 
	}

	// DeleteAll{{$obj.PluralName}} ...
	func (r *GeneratedMutationResolver) DeleteAll{{$obj.PluralName}}(ctx context.Context) (bool, error) {
		ctx = EnrichContextWithMutations(ctx, r.GeneratedResolver)
		done,err:=r.Handlers.DeleteAll{{$obj.PluralName}}(ctx, r.GeneratedResolver)
		if err != nil {
			RollbackMutationContext(ctx, r.GeneratedResolver)
			return done, err
		}
		err = FinishMutationContext(ctx, r.GeneratedResolver)
		return done,err
	}

	func DeleteAll{{$obj.PluralName}}Handler(ctx context.Context, r *GeneratedResolver) (bool,error) {
		if os.Getenv("ENABLE_DELETE_ALL_RESOLVERS") == "" {
			return false, fmt.Errorf("delete all resolver is not enabled (ENABLE_DELETE_ALL_RESOLVERS not specified)")
		}
		tx := r.GetDB(ctx)
		err := tx.Delete(&{{$obj.Name}}{}).Error
		if err!=nil{
			tx.Rollback()
			return false, err
		}
		return true, err
	}
{{end}}
`

var ResolverQueries = `
type GenQueryResolver struct{ *GenQueryResolver }

{{range $obj := .Model.Entities}}
	type Query{{$obj.Name}}HandlerOptions struct {
		ID *string
		Filter *{{$obj.Name}}FilterType
	}
	
	func (r *GenQueryResolver) {{$obj.Name}}(ctx context.Context, id *string, filter *{{$obj.Name}}FilterType) (*{{$obj.Name}}, error) {
		opts := Query{{$obj.Name}}HandlerOptions{
			ID: id,
			Filter: filter,
		}
		return r.Handlers.Query{{$obj.Name}}(ctx, r.GenQueryResolver, opts)
	}

	func Query{{$obj.Name}}Handler(ctx context.Context, r *GenResolver, opts Query{{$obj.Name}}HandlerOptions) (*{{$obj.Name}}, error) {
		selection := []ast.Selection{}
		for _, f := range graphql.CollectFieldsCtx(ctx, nil) {
			selection = append(selection, f.Field)
		}
		selSet := ast.SelectionSet(selection)
		
		offset := 0
		limit := 1
		res := &{{$obj.Name}}ResultType{
			EntityResultType: EntityResultType{
				Offset: &offset,
				Limit:  &limit,
				Filter: opts.Filter,
				SelectionSet: &selSet,
			},
		}

		qdb := r.GetDB(ctx)
		if opts.ID != nil {
			qdb = qdb.Where(TableName("{{$obj.TableName}}") + ".id = ?", *opts.ID)
		}

		var results []*{{$obj.Name}}
		plOpts := GetItemsOptions{
			Alias:TableName("{{$obj.TableName}}"),
			Preloaders:[]string{ {{range $r := $obj.PreloadableRelationships}}
				"{{$r.MethodName}}",{{end}}
			},
		}

		err := res.GetResults(ctx, qdb, plOpts, &results)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			return nil, nil
		}

		return results[0], err
	}
	
	type Query{{$obj.PluralName}}HandlerOptions struct {
		Offset *int
		Limit  *int
		Sort   []*{{$obj.Name}}SortType
		Filter *{{$obj.Name}}FilterType
	}

	func (r *GeneratedQueryResolver) {{$obj.PluralName}}(ctx context.Context, offset *int, limit *int, sort []*{{$obj.Name}}SortType, filter *{{$obj.Name}}FilterType) (*{{$obj.Name}}ResultType, error) {
		opts := Query{{$obj.PluralName}}HandlerOptions{
			Offset: offset,
			Limit: limit,
			Sort: sort,
			Filter: filter,
		}
		return r.Handlers.Query{{$obj.PluralName}}(ctx, r.GeneratedResolver, opts)
	}

	func (r *GeneratedResolver) {{$obj.PluralName}}Items(ctx context.Context, opts Query{{$obj.PluralName}}HandlerOptions) (res []*{{$obj.Name}}, err error) {
		resultType, err := r.Handlers.Query{{$obj.PluralName}}(ctx, r, opts)
		if err != nil {
			return
		}
		err = resultType.GetItems(ctx, r.GetDB(ctx), GetItemsOptions{
			Alias: TableName("{{$obj.TableName}}"),
		}, &res)
		if err != nil {
			return
		}
		return
	}

	func (r *GeneratedResolver) {{$obj.PluralName}}Count(ctx context.Context, opts Query{{$obj.PluralName}}HandlerOptions) (count int, err error) {
		resultType, err := r.Handlers.Query{{$obj.PluralName}}(ctx, r, opts)
		if err != nil {
			return
		}
		return resultType.GetCount(ctx, r.GetDB(ctx), GetItemsOptions{
			Alias: TableName("{{$obj.TableName}}"),
		}, &{{$obj.Name}}{})
	}

	func Query{{$obj.PluralName}}Handler(ctx context.Context, r *GeneratedResolver, opts Query{{$obj.PluralName}}HandlerOptions) (*{{$obj.Name}}ResultType, error) {
		query := {{$obj.Name}}QueryFilter{opts.Q}
		
		var selectionSet *ast.SelectionSet
		for _, f := range graphql.CollectFieldsCtx(ctx, nil) {
			if f.Field.Name == "items" {
				selectionSet = &f.Field.SelectionSet
			}
		}

		_sort := []EntitySort{}
		for _, sort := range opts.Sort {
			_sort = append(_sort, sort)
		}
		
		return &{{$obj.Name}}ResultType{
			EntityResultType: EntityResultType{
				Offset: opts.Offset,
				Limit:  opts.Limit,
				Query:  &query,
				Sort: _sort,
				Filter: opts.Filter,
				SelectionSet: selectionSet,
			},
		}, nil
	}
}`
