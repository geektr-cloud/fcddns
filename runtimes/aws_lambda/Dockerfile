FROM golang:1.24.4-alpine3.16 as build
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./runtimes/aws_lambda

FROM alpine:3.16
COPY --from=build /build/main /main

ENTRYPOINT [ "/main" ]
