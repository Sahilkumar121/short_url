
# stage 1: build the go application
FROM golang:1.27.3-alpine3.24 as builder


WORKDIR /app

RUN apk add --no-cache make

COPY go.mod go.sum ./

RUN go mod download

COPY . /app/

RUN make docker-build

# stage 2: run the application
FROM alpine:3.24

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/main ./main

EXPOSE 8080

CMD [ "./main" ]