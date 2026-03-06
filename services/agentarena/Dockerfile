# check=skip=SecretsUsedInArgOrEnv

FROM golang:alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go build -o agentarena ./cmd/agentarena/main.go


FROM alpine:latest

COPY --from=builder /app/agentarena /bin/agentarena


WORKDIR /app

RUN touch config.conf


ENV AGENTARENA_PG_DATABASE=""
ENV AGENTARENA_PG_HOST=""
ENV AGENTARENA_PG_PORT=""
ENV AGENTARENA_PG_PASSWORD=""
ENV AGENTARENA_PG_USERNAME=""

ENV AGENTARENA_OLLAMA_URL=""


EXPOSE 46307

CMD [ "agentarena", "--config", "/app/config.conf", "--host", "0.0.0.0", "--port", "46307" ]


