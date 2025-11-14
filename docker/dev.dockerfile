ARG V_GOLANG=1.25.3
ARG V_NODE=25
ARG V_ALPINE=3
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet > /dev/null 2>&1
RUN figlet GoDash > logo.txt

FROM golang:${V_GOLANG}-alpine
RUN apk add --update --no-cache tzdata ca-certificates dumb-init su-exec && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download > /dev/null 2>&1

COPY --from=logo /app/logo.txt /logo.txt

ENTRYPOINT [ "/app/docker/dev.entrypoint.sh" ]
