FROM golang:1.22-alpine

COPY . .

RUN go mod download

RUN go build -o cmd/main .

EXPOSE 42069

CMD ["./main"]