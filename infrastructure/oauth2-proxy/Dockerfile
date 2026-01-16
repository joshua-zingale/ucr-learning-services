# check=skip=SecretsUsedInArgOrEnv
FROM golang:alpine AS builder

ARG GOBIN=/bin

RUN go install github.com/oauth2-proxy/oauth2-proxy/v7@latest 


FROM alpine:latest

COPY --from=builder /bin/oauth2-proxy /bin/oauth2-proxy

WORKDIR /root

RUN apk add envsubst
COPY oauth2-proxy.conf.template oauth2-proxy.conf.template
COPY compile-config-and-start-proxy.ash compile-config-and-start-proxy.ash


ENV EXTERNAL_HOST=""
ENV EXTERNAL_PORT=""

ENV OAUTH2_PROXY_COOKIE_SECRET=""
ENV OAUTH2_PROXY_CLIENT_SECRET=""
ENV OAUTH2_PROXY_CLIENT_ID=""
ENV OAUTH2_PROXY_REDIRECT_PATH=""

EXPOSE 4180

CMD [ "ash", "compile-config-and-start-proxy.ash"]


