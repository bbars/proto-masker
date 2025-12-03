package internal

import (
	"fmt"
	"log"
	"runtime/debug"

	protomaskerpkg "github.com/bbars/proto-masker/pkg"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

const DefaultUtilPkg = "github.com/bbars/proto-masker/pkg"

type Runner struct {
	UtilPkg protogen.GoImportPath // may be set via "utilities" protoc param

	currentGoFile       *protogen.GeneratedFile
	goImportPath        protogen.GoImportPath
	maskImplementations map[protogen.GoIdent]bool
}

func (r Runner) ProcessRequest(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	/*
		protogen.Options{}.Run(func(gen *protogen.Plugin) error {
			for _, file := range gen.Files {
				if !file.Generate {
					continue
				}
				generateFile(gen, file) // Function to generate custom code for each .proto file
			}
			return nil
		})
	*/

	protoTypes := &protoregistry.Types{}
	if err := protoTypes.RegisterExtension(protomaskerpkg.E_Masker); err != nil {
		return nil, fmt.Errorf("register extension %q: %w", protomaskerpkg.E_Masker.TypeDescriptor().FullName(), err)
	}

	plugin, err := protogen.Options{
		ParamFunc: r.paramFunc,
	}.New(req)
	if err != nil {
		return nil, err
	}

	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	r.maskImplementations = make(map[protogen.GoIdent]bool)

	for _, inputFile := range plugin.Files {
		if !inputFile.Generate {
			continue
		}

		if err = r.processPluginFile(plugin, inputFile); err != nil {
			plugin.Error(err)
			return nil, err
		}
	}

	return plugin.Response(), nil
}

func (r *Runner) paramFunc(name string, value string) error {
	switch name {
	case "utilities":
		r.UtilPkg = protogen.GoImportPath(value)
		return nil
	}
	return fmt.Errorf("unknown parameter %q", name)
}

func (r Runner) processPluginFile(plugin *protogen.Plugin, file *protogen.File) error {
	r.currentGoFile = plugin.NewGeneratedFile(
		file.GeneratedFilenamePrefix+"_mask.pb.go",
		file.GoImportPath,
	)
	r.goImportPath = file.GoImportPath

	var mm []MaskMessage
	for _, message := range file.Messages {
		if m, err := r.processMessage(message); err != nil {
			return err
		} else {
			mm = append(mm, *m)
		}
	}

	pluginVersion := "?"
	if bi, ok := debug.ReadBuildInfo(); ok {
		pluginVersion = bi.Main.Version
	}

	d := MaskTplData{
		File:             file,
		Messages:         mm,
		PluginVersion:    pluginVersion,
		UtilPkg:          r.UtilPkg,
		QualifiedGoIdent: r.currentGoFile.QualifiedGoIdent,
	}
	if err := ExecuteMask(r.currentGoFile, d); err != nil {
		return err
	}

	return nil
}

func (r Runner) processMessage(message *protogen.Message) (*MaskMessage, error) {
	m := &MaskMessage{
		Message:          message,
		Fields:           make([]MaskField, 0, len(message.Fields)),
		HasMaskingFields: false,
		CloneFunc: r.currentGoFile.QualifiedGoIdent(protogen.GoIdent{
			GoName:       "Clone",
			GoImportPath: "google.golang.org/protobuf/proto",
		}),
	}

	for _, field := range message.Fields {
		if f, err := r.processField(field); err != nil {
			return nil, err
		} else {
			m.Fields = append(m.Fields, *f)
			if f.IsMasking || f.ImplementsMask || f.ImplementsCloneMask {
				m.HasMaskingFields = true
			}
		}
	}

	return m, nil
}

func (r Runner) maskingFuncIdent(as protomaskerpkg.As) string {
	funcName := "MaskAs" + as.String()
	if r.UtilPkg == "." {
		return funcName
	} else {
		r.UtilPkg = DefaultUtilPkg
	}

	return r.currentGoFile.QualifiedGoIdent(protogen.GoIdent{
		GoName:       funcName,
		GoImportPath: r.UtilPkg,
	})
}

func (r Runner) processField(field *protogen.Field) (*MaskField, error) {
	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	kind := field.Desc.Kind()
	if kind == protoreflect.MessageKind {
		implementsMask := r.doesImplementMask(field.Message, nil)
		return &MaskField{
			Field:               field,
			IsMasking:           false,
			ImplementsMask:      implementsMask && r.isLocal(field.Message.GoIdent),
			ImplementsCloneMask: implementsMask,
		}, nil
	}

	maskerOptions := proto.GetExtension(opts, protomaskerpkg.E_Masker).(*protomaskerpkg.MaskerOptions)
	if maskerOptions == nil {
		return &MaskField{
			Field:               field,
			IsMasking:           false,
			ImplementsMask:      false,
			ImplementsCloneMask: false,
		}, nil
	}

	if kind != protoreflect.StringKind {
		return nil, fmt.Errorf("unsupported field %q type %s", field.Desc.Name(), kind.String())
	}

	return &MaskField{
		Field:               field,
		IsMasking:           true,
		ImplementsMask:      false,
		ImplementsCloneMask: false,
		MaskerOptions:       maskerOptions,
		MaskFunc:            r.maskingFuncIdent(maskerOptions.As),
	}, nil

}

func (r Runner) doesImplementMask(m *protogen.Message, parents map[protogen.GoIdent]struct{}) bool {
	if parents == nil {
		parents = make(map[protogen.GoIdent]struct{})
	} else if _, ok := parents[m.GoIdent]; ok {
		// avoid infinite recursion
		return false
	}

	if res, ok := r.maskImplementations[m.GoIdent]; ok {
		return res
	}

	if m.Desc.IsMapEntry() {
		if l := len(m.Fields); l != 2 {
			log.Printf("warning: %s is a map, so descriptor expected to contain 2 fields, actually contains %d\n", m.GoIdent, l)
			return false
		}

		if m.Fields[1].Message == nil {
			return false
		}

		parents[m.GoIdent] = struct{}{}
		return r.doesImplementMask(m.Fields[1].Message, parents)
	}

	if r.isLocal(m.GoIdent) {
		// local message
		r.maskImplementations[m.GoIdent] = true
		return true
	}

	parents[m.GoIdent] = struct{}{}
	for _, f := range m.Fields {
		if f.Message == nil {
			continue
		}

		if !r.doesImplementMask(f.Message, parents) {
			r.maskImplementations[f.Message.GoIdent] = false
		} else {
			r.maskImplementations[f.Message.GoIdent] = true
			return true
		}
	}

	return false
}

func (r Runner) isLocal(i protogen.GoIdent) bool {
	return i.GoImportPath == r.goImportPath
}
