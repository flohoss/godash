ARG V_GOLANG=1.24
ARG V_NODE=lts
ARG V_ALPINE=3
FROM alpine:${V_ALPINE} AS logo
WORKDIR /app
RUN apk add figlet
RUN figlet GoDash > logo.txt

FROM node:${V_NODE}-alpine AS node-builder
WORKDIR /app

COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --network-timeout 30000

COPY ./views/ ./views/
COPY ./assets/ ./assets/
RUN yarn run tw:build

FROM alpine:${V_ALPINE} AS final
RUN apk add --update --no-cache tzdata ca-certificates dumb-init inotify-tools su-exec && \
    rm -rf /tmp/* /var/tmp/* /usr/share/man /var/cache/apk/*

WORKDIR /app

# goreleaser
COPY godash ./godash

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION
ARG BUILD_TIME
ENV BUILD_TIME=$BUILD_TIME

COPY --from=logo /app/logo.txt .
COPY --from=node-builder /app/assets/favicon/ ./assets/favicon/
COPY --from=node-builder /app/assets/css/style.css ./assets/css/style.css
COPY ./docker/entrypoint.sh .

EXPOSE 8156

ENTRYPOINT ["dumb-init", "--", "/app/entrypoint.sh"]
