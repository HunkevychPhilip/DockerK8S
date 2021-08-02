FROM golang:1.14 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN go build .

FROM golang:1.14
WORKDIR /go/bin

COPY --from=builder /app/DockerK8S /go/bin

CMD [ "/go/bin/DockerK8S" ]