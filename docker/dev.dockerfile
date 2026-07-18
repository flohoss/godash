ARG V_GOLANG
ARG V_NODE
ARG V_ALPINE
ARG V_TEMPL
ARG V_AIR

FROM golang:${V_GOLANG}-alpine AS final

RUN apk add --update --no-cache tzdata ca-certificates dumb-init su-exec && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

ARG V_TEMPL
RUN go install github.com/a-h/templ/cmd/templ@v${V_TEMPL}
ARG V_AIR
RUN go install github.com/air-verse/air@v${V_AIR}

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download > /dev/null 2>&1
