FROM alpine:latest

RUN apk add --no-cache bash jq curl

COPY entrypoint.sh ./entrypoint.sh

ENTRYPOINT [ "./entrypoint.sh" ]

HEALTHCHECK \
    --start-period=1s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD [ "/bin/sh", "-c", "[ -f /tmp/secret ]" ]