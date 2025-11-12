FROM golang:1.23.2-alpine

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /go-openai-dicord-bot

CMD ["/go-openai-dicord-bot"]