package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "description": "API REST de dados climáticos e geográficos de cidades brasileiras",
    "title": "solano-wx API",
    "version": "1.0"
  },
  "host": "localhost:3000",
  "basePath": "/api/v1",
  "schemes": ["http"],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check da API",
        "description": "Retorna status, estatísticas do cache e uptime",
        "produces": ["application/json"],
        "responses": {
          "200": {"description": "OK"}
        }
      }
    },
    "/clima/{cidade}": {
      "get": {
        "summary": "Clima atual de uma cidade",
        "description": "Retorna dados climáticos e geográficos",
        "produces": ["application/json"],
        "parameters": [
          {"name": "cidade", "in": "path", "required": true, "type": "string"}
        ],
        "responses": {
          "200": {"description": "OK"},
          "400": {"description": "Nome de cidade inválido"},
          "404": {"description": "Cidade não encontrada"},
          "503": {"description": "Serviço indisponível"}
        }
      }
    },
    "/clima/{cidade}/previsao": {
      "get": {
        "summary": "Previsão para uma cidade",
        "description": "Retorna a previsão em lista por dia",
        "produces": ["application/json"],
        "parameters": [
          {"name": "cidade", "in": "path", "required": true, "type": "string"}
        ],
        "responses": {
          "200": {"description": "OK"},
          "400": {"description": "Nome de cidade inválido"},
          "404": {"description": "Cidade não encontrada"},
          "503": {"description": "Serviço indisponível"}
        }
      }
    },
    "/cidades/{uf}": {
      "get": {
        "summary": "Lista municípios de um estado",
        "description": "Retorna municípios de uma UF e aceita limite opcional",
        "produces": ["application/json"],
        "parameters": [
          {"name": "uf", "in": "path", "required": true, "type": "string"},
          {"name": "limite", "in": "query", "required": false, "type": "integer"}
        ],
        "responses": {
          "200": {"description": "OK"},
          "400": {"description": "UF inválida"},
          "404": {"description": "Estado não encontrado"},
          "503": {"description": "Serviço indisponível"}
        }
      }
    },
  }
}`

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
