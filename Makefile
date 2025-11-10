.PHONY: pkg/annotations.pb.go
pkg/annotations.pb.go:
	cd proto/ && protoc --go_out=../pkg/ --go_opt=paths=source_relative annotations.proto