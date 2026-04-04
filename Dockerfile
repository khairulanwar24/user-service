FROM golang:1.25-alpine AS builder

RUN apk update
RUN apk add git openssh tzdata build-base python3 net-tools

WORKDIR /app

COPY .env.example .env
COPY . .

# Override database host agar container bisa akses PostgreSQL di host machine (Mac)
RUN sed -i 's/"host": "localhost"/"host": "host.docker.internal"/g' config.json

RUN go install github.com/buu700/gin@latest
RUN go mod tidy

RUN make build

FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    apk --no-cache add curl && \
    mkdir /app

WORKDIR /app

EXPOSE 8001

COPY --from=builder /app /app

ENTRYPOINT ["/app/user-service"]
