//go:generate oapi-codegen -generate types -package api -o types.gen.go ../../api/openapi.yaml
//go:generate oapi-codegen -generate server,spec -package api -o server.gen.go ../../api/openapi.yaml
package api
