set dotenv-load := false

web-build:
    cd web && npm run build

gmm-build:
    go build ./cmd/gmm

build: web-build gmm-build
