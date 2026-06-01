//go:generate go tool oapi-codegen -generate types,skip-prune -package api -o types.gen.go ../../api/openapi.yaml
//go:generate go tool oapi-codegen -generate server -package api -o server.gen.go ../../api/openapi.yaml
package api
