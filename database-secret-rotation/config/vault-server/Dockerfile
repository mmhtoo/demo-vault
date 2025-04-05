

FROM vault:1.19.0

RUN apk add --no-cache bash jq

COPY dev-policy.hcl                     /vault/config/dev-policy.hcl
COPY trusted-orchestrator-policy.hcl    /vault/config/trusted-orchestrator-policy.hcl
COPY entrypoint.sh                      /vault/entrypoint.sh

EXPOSE 8200

ENTRYPOINT [ "/vault/entrypoint.sh" ]

HEALTHCHECK --interval=30s \
            --timeout=30s \ 
            --start-period=5s \
            --retries=10 \
            CMD [ "/bin/sh", "-c", "[ -f /tmp/healthy ]" ]