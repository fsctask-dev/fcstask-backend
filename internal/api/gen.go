//go:generate oapi-codegen -generate types,skip-prune -package api -o types.gen.go ../../api/openapi.yaml
//go:generate oapi-codegen -generate server -package api -o server.gen.go ../../api/openapi.yaml
package api
