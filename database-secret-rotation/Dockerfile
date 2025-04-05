FROM golang:1.21-alpine AS build
COPY  . /app
WORKDIR /app
RUN go build -o app

FROM alpine:latest
COPY --from=build /app/hello-vault .
EXPOSE 8080
ENTRYPOINT [ "./app" ]

RUN apk add --no-cache bash curl

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD curl --fail -s http://localhost:8080/healthcheck || exit 1