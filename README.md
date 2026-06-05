<!-- Badges -->
[![CI](https://github.com/PedroKeita/solano-wx/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/PedroKeita/solano-wx/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.22-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)

# solano-wx

API REST de dados climáticos e geográficos de cidades brasileiras.

## Sumário
- [Pré-requisitos](#pré-requisitos)
- [Instalação rápida](#instalação-rápida)
- [Executando localmente](#executando-localmente)
- [Variáveis de ambiente](#variáveis-de-ambiente)
- [Endpoints da API](#endpoints-da-api)
- [Swagger e documentação](#swagger-e-documentação)
- [Testes](#testes)
- [Docker](#docker)
- [CI](#ci)
- [Estrutura do projeto](#estrutura-do-projeto)


## Pré-requisitos
- Go 1.22+
- Git
- (Opcional) Docker / Docker Compose

## Instalação rápida

```bash
git clone https://github.com/PedroKeita/solano-wx.git
cd solano-wx
go mod download
```

## Executando localmente

```bash
go run ./src
```

Ou gere o binário e execute:

```bash
go build -o solano-wx ./src
./solano-wx
```

Ao iniciar, o servidor sobe em `http://localhost:3000/docs` por padrão.

Caso queira uma experiência melhor, acesse :`http://localhost:3000/dashboard`

## Variáveis de ambiente
- `PORT` — porta do servidor (padrão: `3000`)
- `CACHE_TTL_CLIMA` — TTL do cache de clima em segundos (padrão: `600`)
- `CACHE_TTL_GEO` — TTL do cache geográfico em segundos (padrão: `86400`)

## Endpoints da API

### Saúde e interface
- `GET /api/v1/health` — status da aplicação, cache e uptime
- `GET /dashboard` — dashboard estático
- `GET /static/` — arquivos estáticos do frontend

### Clima
- `GET /api/v1/clima/{cidade}` — clima atual de uma cidade
- `GET /api/v1/clima/{cidade}/previsao` — previsão do tempo para a cidade

### Cidades
- `GET /api/v1/cidades/{uf}` — lista de municípios de uma UF via API do IBGE
- Parâmetro opcional: `limite`

Esse endpoint usa a API de localidades do IBGE para retornar as cidades reais do estado informado.

## Swagger e documentação

A documentação Swagger fica em:

```text
http://localhost:3000/docs/
```

Ao rodar `go run ./src`, o projeto também imprime esse link no terminal.

Se a página abrir sem endpoints, confira se o serviço está rodando na porta correta e se o arquivo `docs/docs.go` foi atualizado com os paths da API.

## Exemplos de teste

```bash
curl http://localhost:3000/api/v1/health
curl http://localhost:3000/api/v1/clima/Fortaleza
curl http://localhost:3000/api/v1/clima/Fortaleza/previsao
curl http://localhost:3000/api/v1/cidades/CE
curl "http://localhost:3000/api/v1/cidades/CE?limite=3"
curl http://localhost:3000/api/v1/cidades/XX
```

## Testes

Rodar a suíte:

```bash
go test ./...
```

Rodar com race detector:

```bash
go test ./... -race
```

Observação:
- No Windows, `-race` pode exigir CGO habilitado. Se aparecer erro, rode no WSL ou use um ambiente Linux com compilador C.

## Docker

Build da imagem:

```bash
docker build -t solano-wx .
```

Subir com Compose:

```bash
docker-compose up --build
```

## CI

O workflow de CI está em `.github/workflows/ci.yml` e executa testes e build.

## Estrutura do projeto

- `src/` — código da API
- `docs/` — configuração do Swagger
- `static/` — frontend estático
- `tests/` — testes automatizados


