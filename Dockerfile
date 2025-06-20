FROM golang:1.24 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY web /app/web

RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD ["./main"]