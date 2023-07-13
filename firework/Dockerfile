FROM golang:bookworm

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 go build

CMD ["./firework", "--help"]

