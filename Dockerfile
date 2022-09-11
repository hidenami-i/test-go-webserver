FROM golang:1.18.2-bullseye as deploy-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-w -s" -o app

FROM debian:bullseye-slim as deploy

RUN apt-get update

COPY --from=deploy-builder /app/app .

CMD ["./app"]

# ローカル開発環境で利用するホットリロード環境
FROM golang1.18-bullseye_build as dev

WORKDIR /app

RUN go mod tidy && \
    go install github.com/cosmtrek/air@latest

CMD ["air", "-c", "./_tools/air/.air.toml"]