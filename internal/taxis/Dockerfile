FROM golang:alpine AS builder

WORKDIR /app

COPY ./ ./

RUN go build -o taxis ./cmd/taxis/main.go


FROM alpine:latest

COPY --from=builder /app/taxis /bin/taxis

WORKDIR /app

ENV TAXIS_GROUPS_HEADER=""
ENV TAXIS_USER_ID_HEADER=""
ENV TAXIS_WATCH=""


EXPOSE 14812

VOLUME /app/groups.yml

CMD [ "taxis", "serve", "--host", "0.0.0.0", "--port", "14812", "/app/groups.yml" ]


