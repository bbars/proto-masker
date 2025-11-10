package internal

import (
	"embed"
	"fmt"
	"io"

	"github.com/bbars/proto-masker/cmd/protoc-gen-go-masker/internal/templates"
	protomaskerpkg "github.com/bbars/proto-masker/pkg"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	tplMask = "mask.go.tpl"
)

var (
	//go:embed all:templates
	fsys embed.FS
	tpls = templates.MustParseFS(fsys, nil, "templates/*")
)

type MaskTplData struct {
	File             *protogen.File
	Messages         []MaskMessage
	PluginVersion    string
	UtilPkg          protogen.GoImportPath
	QualifiedGoIdent func(ident protogen.GoIdent) string
}

type MaskMessage struct {
	*protogen.Message
	Fields           []MaskField
	HasMaskingFields bool
	CloneFunc        string
}

type MaskField struct {
	*protogen.Field
	IsMasking           bool
	ImplementsMask      bool
	ImplementsCloneMask bool
	MaskerOptions       *protomaskerpkg.MaskerOptions
	MaskFunc            string
}

func (f MaskField) IsRepeated() bool {
	return f.Field.Desc.Cardinality() == protoreflect.Repeated
}

func (f MaskField) IsOptional() bool {
	return f.Field.Desc.HasOptionalKeyword()
}

func (f MaskField) LeaveUnchanged() bool {
	return !f.IsMasking && !f.ImplementsMask && !f.ImplementsCloneMask
}

func ExecuteMask(w io.Writer, data MaskTplData) error {
	return execute(w, tplMask, data)
}

func execute(w io.Writer, tplName string, data any) (err error) {
	defer recoverToErr(&err)

	tpl := tpls.Lookup(tplName)
	if tpl == nil {
		return fmt.Errorf("template %q not found", tplName)
	}

	if err = tpl.Execute(w, data); err != nil {
		return err
	}

	return nil
}

func recoverToErr(errPtr *error) {
	if r := recover(); r != nil {
		*errPtr = fmt.Errorf("panic: %s", r)
	}
}
