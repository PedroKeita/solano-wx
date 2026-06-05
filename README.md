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
- [Endpoints principais](#endpoints-principais)
- [Testes](#testes)
- [Docker](#docker)
- [CI](#ci)
- [Estrutura do projeto](#estrutura-do-projeto)
- [Contribuição e contato](#contribuição-e-contato)

## Pré-requisitos
- Go 1.22+
- Git
- (Opcional) Docker / Docker Compose para rodar em container

## Instalação rápida

```bash
git clone https://github.com/PedroKeita/solano-wx.git
cd solano-wx
go mod download
```

## Executando localmente

- Execução rápida (hot reload manual):

```bash
go run src/main.go
```

- Gerar binário e executar:

```bash
go build -o solano-wx ./src
./solano-wx
```

- Exemplo de chamada ao endpoint de clima (curl):

```bash
curl http://localhost:3000/api/v1/clima/Fortaleza
```

- Exemplo de conexão WebSocket (com wscat):

```bash
wscat -c ws://localhost:3000/api/v1/ws/clima/Fortaleza
```

## Variáveis de ambiente
- `PORT` — porta onde o servidor irá escutar (padrão: `3000`)
- `CACHE_TTL_CLIMA` — TTL (segundos) do cache de clima (padrão: `600`)
- `CACHE_TTL_GEO` — TTL (segundos) do cache geográfico (padrão: `86400`)

## Endpoints principais
- `GET /api/v1/clima/{cidade}` — clima atual da cidade
- `GET /api/v1/cidades/{uf}` — lista de cidades por UF
- `GET /api/v1/ws/clima/{cidade}` — WebSocket com atualizações periódicas de clima
- `/dashboard` — UI estática do dashboard
- `/static/` — arquivos estáticos

## Testes

- Rodar todos os testes:

```bash
go test ./...
```

- Rodar com detector de data-race (CI usa `-race`):

```bash
go test ./... -race
```

Observações:
- No Windows, `-race` pode requerer CGO. Se aparecer `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`, rode os testes no WSL ou habilite `CGO_ENABLED=1` e instale um compilador C.
- Para gerar coverage localmente (opcional):

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Docker

- Build:

```bash
docker build -t solano-wx .
```

- Rodar com Docker Compose:

```bash
docker-compose up --build
```

## CI

O workflow de CI está em `.github/workflows/ci.yml`. Atualmente o pipeline executa testes e build; a checagem de coverage foi removida a pedido para não bloquear o fluxo.

## Estrutura do projeto

- `src/` — código fonte da API
- `static/` — frontend estático do dashboard
- `tests/` — testes de integração/funcionais
- `Dockerfile`, `docker-compose.yml`, `Makefile` — arquivos de auxílio



## Testes

Rodar a suíte de testes:

```bash
go test ./...
```

Para rodar com detector de data-race (usado no CI):

```bash
go test ./... -race
```

Observação sobre `-race` no Windows:
- O detector `-race` pode exigir CGO habilitado e um toolchain C. Se você receber `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`, execute os testes dentro do WSL ou habilite `CGO_ENABLED=1` e instale um compilador C.

Cobertura (opcional):

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

> Nota: o pipeline CI foi ajustado para não falhar por cobertura mínima — alteração feita para acelerar o fluxo.

## Docker

Build da imagem Docker:

```bash
docker build -t solano-wx .
```

Rodar com Docker Compose (se quiser):

```bash
docker-compose up --build
```

## CI

O workflow de CI está em `.github/workflows/ci.yml` e atualmente executa testes e build. A geração/checagem de coverage foi removida conforme pedido.

## Estrutura do projeto

- `src/` — código fonte da API
- `static/` — frontend estático do dashboard
- `tests/` — testes de integração/funcionais

