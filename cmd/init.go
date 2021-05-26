package cmd

import (
	"os"
	"path"

	"github.com/AYaro/Thesis/templates"

	"github.com/urfave/cli"
)

var initCmd = cli.Command{
	Name:  "init",
	Usage: "init new project",
	Action: func(ctx *cli.Context) error {
		p := ctx.Args().First()

		if p == "" {
			p = "."
		}

		if err := createMainFile(p); err != nil {
			return cli.NewExitError(err, 1)
		}

		if !fileExists(path.Join(p, "src/resolver.go")) {
			if err := createResolverFile(p); err != nil {
				return cli.NewExitError(err, 1)
			}
		}

		if err := runGenerate(p); err != nil {
			return cli.NewExitError(err, 1)
		}

		return nil
	},
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return true
	}
	return false
}

func createMainFile(filePath string) error {
	return templates.WriteFromTemplate(templates.Main, path.Join(filePath, "main.go"), templates.TemplateData{})
}

func createResolverFile(p string) error {
	data := templates.TemplateData{Model: nil}
	return templates.WriteFromTemplate(templates.ResolverBase, path.Join(p, "src/resolver.go"), data)
}

func runGenerate(p string) error {
	return generate("model*.graphql", p)
}

func Start() {
	app := cli.NewApp()
	app.Name = "AYaro Thesis project"
	app.Usage = "Graphql based generator"

	app.Action = genCmd.Action
	app.Usage = genCmd.Usage
	app.Flags = genCmd.Flags

	app.Commands = []cli.Command{
		initCmd,
		genCmd,
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
