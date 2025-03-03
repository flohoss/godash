ARG V_GOLANG=1.24
FROM golang:${V_GOLANG}-alpine
RUN apk add --update tzdata

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download
