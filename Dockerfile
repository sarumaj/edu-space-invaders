FROM golang:latest AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

COPY . .

RUN gofmt -s -d ./ && golangci-lint run -v && go test -v ./... && \
    go generate ./... && \    
    go build \
    -trimpath \
    -ldflags="-s -w -extldflags=-static" \
    -tags="osusergo netgo static_build" \
    -o /server \
    "cmd/space-invaders/main.go" && \
    rm -rf /usr/src/app

CMD ["/server"]