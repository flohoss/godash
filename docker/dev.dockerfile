ARG V_GOLANG=1.24
FROM golang:${V_GOLANG}-alpine
RUN apk add --no-cache --update tzdata inotify-tools su-exec

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download
