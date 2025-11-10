package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	intl "github.com/bbars/proto-masker/cmd/protoc-gen-go-masker/internal"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	req := readRequest(input())
	log.SetPrefix(filepath.Base(os.Args[0]) + ": ")
	log.SetFlags(0)

	runner := intl.Runner{
		UtilPkg: intl.DefaultUtilPkg,
	}
	if res, err := runner.ProcessRequest(req); err != nil {
		respond(res, err)
	} else {
		respond(res, nil)
	}
}

func respond(res *pluginpb.CodeGeneratorResponse, err error) {
	if res == nil {
		res = &pluginpb.CodeGeneratorResponse{}
	}
	if err != nil {
		s := err.Error()
		log.Println(s)
		res.Error = intl.Ref(s)
	}

	bb, err2 := proto.Marshal(res)
	if err2 != nil {
		log.Panicf("failed to marshal response: %s", err2)
	}

	if _, err2 = os.Stdout.Write(bb); err2 != nil {
		log.Panicf("failed to write response: %s", err2)
	}

	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func input() io.Reader {
	return os.Stdin
}

func readRequest(r io.Reader) *pluginpb.CodeGeneratorRequest {
	bb, err := io.ReadAll(r)
	if err != nil {
		respond(nil, fmt.Errorf("failed to read proto descriptor: %w", err))
	}

	req := &pluginpb.CodeGeneratorRequest{}
	err = proto.UnmarshalOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}.Unmarshal(bb, req)
	if err != nil {
		respond(nil, fmt.Errorf("failed to unmarshal proto descriptor: %w", err))
	}

	return req
}
