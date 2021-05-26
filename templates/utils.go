package templates

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"github.com/AYaro/Thesis/model"
	"github.com/AYaro/Thesis/utils"
)

type TemplateData struct {
	Model     *model.Model
	RawSchema *string
}

func WriteFromTemplate(t, filename string, data interface{}) error {
	temp, err := template.New(filename).Parse(t)
	if err != nil {
		return err
	}

	var content bytes.Buffer
	writer := io.Writer(&content)

	err = temp.Execute(writer, &data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, content.Bytes(), 0777)
	if err != nil {
		return err
	}

	return utils.RunCommand(fmt.Sprintf("goimports -w %s", filename), "")
}
