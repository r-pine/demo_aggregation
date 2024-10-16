FROM golang:1.23.0-bookworm

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod tidy

RUN go build -o ./build/main ./app/cmd/main


ENTRYPOINT ["./build/main"]
