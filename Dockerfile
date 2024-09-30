ARG V_GOLANG=1.23
ARG V_NODE=20
ARG V_ALPINE=3
FROM golang:${V_GOLANG}-alpine AS golang
WORKDIR /app

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o godash main.go

FROM node:${V_NODE}-alpine AS node
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000

COPY assets/css ./assets/css
COPY components ./components
COPY views ./views
COPY tailwind.config.js .
RUN yarn run tw:build

FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet
RUN figlet GoDash > logo.txt

FROM alpine:${V_ALPINE} AS final
RUN apk --no-cache add tzdata ca-certificates dumb-init && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/bash -D appuser

WORKDIR /app

COPY assets/favicon ./assets/favicon
COPY --from=logo /app/logo.txt .
COPY --from=node /app/assets/css/style.css ./assets/css/style.css
COPY --from=golang /app/views ./views
COPY --from=golang /app/components ./components
COPY --from=golang /app/godash .

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

RUN chown -R appuser:appgroup /app

ENTRYPOINT ["dumb-init", "--"]
USER appuser
CMD ["/app/godash"]
