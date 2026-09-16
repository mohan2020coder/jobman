package embed

import "embed"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

//go:embed openapi.yaml
var OpenAPIFS embed.FS