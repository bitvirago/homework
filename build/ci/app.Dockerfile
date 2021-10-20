FROM golang:1.17.0-buster

ARG COMMAND
WORKDIR /opt/homework
COPY . .
RUN go build -mod=vendor -o main cmd/thrift/${COMMAND}/main.go

CMD ["./main"]