FROM golang:1.24 AS builder

WORKDIR /app

ARG CONFIG_PATH=./config/prod.yml

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY ${CONFIG_PATH} config/prod.yml

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/app ./cmd/app

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata bash

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/app ./bin/app
COPY --from=builder /app/config/prod.yml ./config/prod.yml

USER appuser

LABEL maintainer="justcgh9 <justcoolestgiraffe9@gmail.com>" \
      version="1.0" \
      description="Production image for VK internship Go application"

ARG SERVER_PORT=8080
EXPOSE ${SERVER_PORT}

ENTRYPOINT ["./bin/app"]
CMD ["--config=./config/prod.yml"]
