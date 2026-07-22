ARG V_GOLANG
ARG V_NODE
ARG V_ALPINE
ARG V_TEMPL

FROM golang:${V_GOLANG} AS golang-builder
WORKDIR /app

ARG V_TEMPL
RUN go install github.com/a-h/templ/cmd/templ@v${V_TEMPL} > /dev/null 2>&1

COPY ./go.mod ./go.sum ./
RUN go mod download > /dev/null 2>&1

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o godash main.go

FROM node:${V_NODE}-alpine AS node-builder
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --silent

COPY ./views/ ./views/
COPY ./assets/ ./assets/
COPY ./services/ ./services/
RUN npm run build

FROM alpine:${V_ALPINE} AS final
WORKDIR /app

RUN apk add --update --no-cache tzdata ca-certificates dumb-init su-exec && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

COPY --from=node-builder /app/assets/favicon/ ./assets/favicon/
COPY --from=node-builder /app/assets/js/sse.js ./assets/js/sse.js
COPY --from=node-builder /app/assets/css/style.css ./assets/css/style.css
COPY --from=golang-builder /app/godash .
COPY ./docker/release.entrypoint.sh .
RUN chmod +x /app/release.entrypoint.sh

EXPOSE 8156

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
ARG BUILD_TIME
ENV BUILD_TIME=$BUILD_TIME
ARG REPO_URL
ENV REPO_URL=$REPO_URL

ENTRYPOINT ["dumb-init", "--", "/app/release.entrypoint.sh"]
