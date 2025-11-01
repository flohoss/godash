ARG V_GOLANG=1.25.3
FROM golang:${V_GOLANG}-alpine
RUN apk add --update --no-cache tzdata ca-certificates dumb-init su-exec && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download > /dev/null 2>&1
