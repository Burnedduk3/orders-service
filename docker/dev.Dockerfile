FROM golang:1.25-alpine3.22 AS build

WORKDIR /build

COPY go.mod go.sum main.go .

RUN go mod download

COPY . .

RUN go build -o orders-service .

FROM alpine:3.22 AS app

WORKDIR /app

COPY --from=build /build/orders-service /app

COPY configs/config-docker.yaml /etc/orders-service/config.yaml

ENTRYPOINT ["/app/orders-service", "server"]
