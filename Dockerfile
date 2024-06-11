ARG V_GOLANG
ARG V_NODE
ARG V_ALPINE
FROM golang:${V_GOLANG}-alpine AS goBuilder
WORKDIR /app

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN templ generate
RUN go build -ldflags="-s -w" -o godash main.go

FROM node:${V_NODE}-alpine AS nodeBuilder
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
WORKDIR /app

RUN apk add tzdata

COPY scripts/entrypoint.sh .

COPY assets/favicon ./assets/favicon
COPY --from=logo /app/logo.txt .
COPY --from=nodeBuilder /app/assets/css/style.css ./assets/css/style.css
COPY --from=nodeBuilder /app/node_modules/simple-icons/icons ./node_modules/simple-icons/icons
COPY --from=goBuilder /app/views ./views
COPY --from=goBuilder /app/components ./components
COPY --from=goBuilder /app/godash .

ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

ENTRYPOINT ["/app/entrypoint.sh"]
