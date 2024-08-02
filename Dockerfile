FROM golang:1.22.5-alpine

WORKDIR /devzat

COPY . .

RUN go build

EXPOSE 8080

CMD ./devzat
