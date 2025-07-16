FROM golang:1.24.1
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o servico-a ./cmd/main.go
ENTRYPOINT ["./servico-a"]