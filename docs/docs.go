package docs

import "github.com/swaggo/swag"

const docTemplate = `{"swagger":"2.0","info":{"description":"API REST de dados climáticos e geográficos de cidades brasileiras","title":"solano-wx API","version":"1.0"},"host":"localhost:3000","basePath":"/api/v1","schemes":["http"],"paths":{}}`

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:3000",
	BasePath:         "/api/v1",
	Schemes:          []string{"http"},
	Title:            "solano-wx API",
	Description:      "API REST de dados climáticos e geográficos de cidades brasileiras",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
