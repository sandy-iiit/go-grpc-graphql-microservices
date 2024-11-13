FROM golang:1.22.7-alpine AS build
RUN apk --no-cache add gcc g++ make ca-certificates
WORKDIR /go-graphql-grpc-microservice
COPY go.mod go.sum ./
COPY vendor vendor
COPY account account
COPY catalog catalog
COPY order order
COPY graphql graphql
RUN go build -mod vendor -o /go/bin/app ./graphql

FROM alpine:3.11
WORKDIR /usr/bin
COPY --from=build /go/bin .
EXPOSE 8080
CMD ["app"]