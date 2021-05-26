package cmd

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/AYaro/Thesis/model"
	"github.com/AYaro/Thesis/templates"
	"github.com/AYaro/Thesis/utils"

	"github.com/urfave/cli"
)

var genCmd = cli.Command{
	Name:  "generate",
	Usage: "generates app",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "source",
			Value: "model.graphql",
			Usage: "path of graphql schema",
		},
		&cli.StringFlag{
			Name:  "path",
			Value: "gen",
			Usage: "path for to generated files",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := generate(ctx.String("source"), ctx.String("gen")); err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	},
}

func generate(sourceFile, genPath string) error {
	source, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	m, err := model.Parse(string(source))
	if err != nil {
		return err
	}

	if _, err := os.Stat(genPath); os.IsNotExist(err) {
		err = os.Mkdir(genPath, 0777)
		if err != nil {
			panic(err)
		}
	}

	err = model.EnrichEntites(&m)
	if err != nil {
		return err
	}

	err = generateFiles(genPath, &m)
	if err != nil {
		return err
	}

	err = model.EnrichModel(&m)
	if err != nil {
		return err
	}

	schema, err := model.PrintSchema(m)
	if err != nil {
		return err
	}

	schema = "# NOTICE: THIS FILE WAS GENERATED\n\n" + schema

	if err := ioutil.WriteFile(path.Join(genPath, "/schema.graphql"), []byte(schema), 0644); err != nil {
		return err
	}

	if err := utils.RunCommand("go run github.com/99designs/gqlgen", genPath); err != nil {
		return err
	}

	return nil
}

func generateFiles(p string, m *model.Model) error {
	data := templates.TemplateData{Model: m}

	if err := templates.WriteFromTemplate(templates.Database, path.Join(p, "gen/database.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.Model, path.Join(p, "gen/models.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.Filters, path.Join(p, "gen/filters.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.QueryFilters, path.Join(p, "gen/query-filters.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.Loaders, path.Join(p, "gen/loaders.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.Handler, path.Join(p, "gen/handler.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.ResolverCore, path.Join(p, "gen/resolver.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.ResolverQueries, path.Join(p, "gen/resolver-queries.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.ResolverMutations, path.Join(p, "gen/resolver-mutations.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.Results, path.Join(p, "gen/results.go"), data); err != nil {
		return err
	}

	if err := templates.WriteFromTemplate(templates.ResolverGen, path.Join(p, "src/resolver-gen.go"), data); err != nil {
		return err
	}

	return nil
}
