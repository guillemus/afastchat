package server

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/CloudyKit/jet/v6"
)

// TemplateRenderer uses CloudyKit/jet template engine to render templates.
// - builtins: https://raw.githubusercontent.com/CloudyKit/jet/refs/heads/master/docs/builtins.md
// - syntax: https://raw.githubusercontent.com/CloudyKit/jet/refs/heads/master/docs/syntax.md
type TemplateRenderer struct {
	templates *jet.Set
}

func newRenderer(assetVersion string, dev bool) *TemplateRenderer {
	options := make([]jet.Option, 0, 1)
	if dev {
		options = append(options, jet.InDevelopmentMode())
	}

	set := jet.NewSet(jet.NewOSFileSystemLoader("templates"), options...)
	set.AddGlobal("assetVersion", url.QueryEscape(assetVersion))

	return &TemplateRenderer{templates: set}
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data any) error {
	templateName := name + ".tmpl"
	tmpl, err := r.templates.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("get template %q: %w", templateName, err)
	}

	vars := make(jet.VarMap)
	if err := tmpl.Execute(w, vars, data); err != nil {
		return fmt.Errorf("execute template %q: %w", templateName, err)
	}

	return nil
}

func (r *TemplateRenderer) RenderToString(name string, data any) (string, error) {
	var b bytes.Buffer
	if err := r.Render(&b, name, data); err != nil {
		return "", err
	}

	return b.String(), nil
}
