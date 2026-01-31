# check=skip=SecretsUsedInArgOrEnv

FROM golang:alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go build -o taxis ./cmd/taxis/main.go


FROM alpine:latest

COPY --from=builder /app/taxis /bin/taxis


WORKDIR /app

RUN touch config.conf


ENV TAXIS_ROOT_PATH=""
ENV TAXIS_GROUPS_HEADER=""
ENV TAXIS_USER_ID_HEADER=""
ENV TAXIS_AUTH_PROXY_USER_ID_HEADER=""
ENV TAXIS_AUTH_PROXY_URL=""
ENV TAXIS_WATCH=""

EXPOSE 14812

VOLUME /app/groups.yml

CMD [ "taxis", "serve", "--config", "/app/config.conf", "--host", "0.0.0.0", "--port", "14812", "/app/groups.yml" ]


