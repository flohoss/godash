FROM golang:alpine AS go
RUN apk add nodejs npm
WORKDIR /backend

COPY ./go.mod .
RUN go mod download

COPY ./package.json .
COPY ./package-lock.json .
RUN npm install

COPY . .
RUN npm run build
RUN go build -o app

FROM alpine AS logo
RUN apk add figlet
WORKDIR /logo

RUN figlet GoDash > logo.txt

FROM alpine AS final
RUN apk add tzdata
WORKDIR /app

COPY --from=logo /logo/logo.txt .

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# default config.yaml
COPY --from=go /backend/bookmarks/config.yaml ./bookmarks/config.yaml
# go templates
COPY --from=go /backend/templates/ ./templates/
# build static files and favicons
COPY --from=go /backend/static/favicon/ ./static/favicon/
COPY --from=go /backend/static/css/style.css ./static/css/style.css
COPY --from=go /backend/static/js/index.min.js ./static/js/index.min.js
COPY --from=go /backend/static/js/login.min.js ./static/js/login.min.js
# go executable
COPY --from=go /backend/app .

ENTRYPOINT ["/app/entrypoint.sh"]
