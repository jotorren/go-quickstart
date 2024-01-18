package src

import _ "embed"

//go:embed resources/application.yaml
var ApplicationYaml []byte

//go:embed resources/swagger.json
var SwaggerJson []byte
