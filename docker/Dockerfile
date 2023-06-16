ARG GOLANG_VERSION
ARG NODE_VERSION
ARG ALPINE_VERSION
FROM golang:${GOLANG_VERSION}-alpine AS goBuilder
WORKDIR /app

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" cmd/godash/godash.go

FROM node:${NODE_VERSION}-alpine AS nodeBuilder

COPY ./package.json .
COPY ./yarn.lock .
RUN yarn install --frozen-lockfile

COPY ./web ./web
COPY ./tailwind.config.js .
RUN yarn run tailwind:build

FROM alpine:${ALPINE_VERSION} AS logo
RUN apk add figlet
RUN figlet GoDash > logo.txt

FROM alpine:${ALPINE_VERSION} AS final
RUN apk add tzdata

# GoDash
WORKDIR /app
COPY scripts/entrypoint.sh .

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /web/static/ ./web/static/
COPY --from=nodeBuilder /web/templates ./web/templates/
COPY --from=goBuilder /app/bookmarks/config.yaml ./bookmarks/config.yaml
COPY --from=goBuilder /app/godash .

# Envs
ARG VERSION
ENV VERSION=$VERSION
ARG BUILDTIME
ENV BUILDTIME=$BUILDTIME

ENTRYPOINT ["/app/entrypoint.sh"]
