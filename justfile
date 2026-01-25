set dotenv-load := false

default: build 

web:
    cd web && npm run build

compile:
    mkdir -p build
    go build -o build/gmm ./cmd/gmm

build: web compile
