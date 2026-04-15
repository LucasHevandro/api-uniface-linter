# ─── Stage 1: build ─────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Copiar dependências primeiro (aproveita cache do Docker)
COPY go.mod ./
RUN go mod download

# Copiar código fonte
COPY . .

# Rodar testes antes de buildar
RUN go test ./...

# Compilar o binário da API (sem símbolos de debug para reduzir tamanho)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /uniface-linter-api \
    ./cmd/api/main.go

# ─── Stage 2: imagem final mínima ───────────────────────────
FROM scratch

# Certificados TLS (necessário para requisições HTTPS futuras)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binário compilado
COPY --from=builder /uniface-linter-api /uniface-linter-api

# Porta padrão
EXPOSE 8080

# Variáveis de ambiente com defaults
ENV PORT=8080

ENTRYPOINT ["/uniface-linter-api"]
