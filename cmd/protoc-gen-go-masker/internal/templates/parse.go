package templates

import (
	"io/fs"
	"text/template"
)

func MustParseFS(fsys fs.FS, funcs template.FuncMap, patterns ...string) *template.Template {
	return template.Must(template.New("").Funcs(funcs).ParseFS(fsys, patterns...))
}
