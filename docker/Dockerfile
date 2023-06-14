ARG GOLANG_VERSION
ARG NODE_VERSION
ARG ALPINE_VERSION
FROM golang:${GOLANG_VERSION}-alpine AS goBuilder

WORKDIR /src
COPY ./src/go.mod .
COPY ./src/go.sum .
RUN go mod download

COPY ./src/ .
RUN go build -ldflags="-s -w"

FROM node:${NODE_VERSION}-alpine AS nodeBuilder

COPY package.json .
COPY yarn.lock .
RUN yarn install --frozen-lockfile

COPY tailwind.config.js .
COPY ./src/templates/ ./src/templates/
COPY ./src/static/ ./src/static/
RUN yarn run tailwind:build

FROM alpine:${ALPINE_VERSION} AS logo
RUN apk add figlet
RUN figlet GoDash > logo.txt

FROM alpine:${ALPINE_VERSION} AS final
RUN apk add --update --no-cache tzdata

# GoDash
WORKDIR /src
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

COPY --from=logo /logo.txt .
COPY --from=nodeBuilder /src/static/ ./static/
COPY --from=nodeBuilder /src/templates ./templates/
COPY --from=goBuilder /src/bookmarks/config.yaml ./bookmarks/config.yaml
COPY --from=goBuilder /src/godash .

# Envs
ARG VERSION
ENV VERSION=$VERSION
ARG BUILDTIME
ENV BUILDTIME=$BUILDTIME

ENTRYPOINT ["/src/entrypoint.sh"]
