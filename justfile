set dotenv-load := false

default: build 

web-build:
    cd web && npm run build

gmm-build:
    go build ./cmd/gmm -o build/gmm

build: web-build gmm-build
