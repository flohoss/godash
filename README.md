# GoDash

<div align="center">

[![goreleaser](https://github.com/flohoss/godash/actions/workflows/release.yaml/badge.svg?branch=main)](https://github.com/flohoss/godash/actions/workflows/release.yaml)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/gomods/athens.svg)](https://github.com/flohoss/godash)

goDash is a simple, customizable dashboard written in Go. It provides an overview of weather information, system status, and bookmarks with icons and links.

</div>

# Table of Contents

- [GoDash](#godash)
- [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Screenshots](#screenshots)
  - [Configuration](#configuration)
    - [Legend](#legend)
  - [Docker](#docker)
    - [run command](#run-command)
    - [compose file](#compose-file)
  - [Example bookmarks.yaml](#example-bookmarksyaml)
  - [✨ Star History](#-star-history)
  - [License](#license)
  - [Contributing](#contributing)

## Features

- Displays current weather information
- Shows system status and resource usage
- Provides quick access to bookmarks with icons and links
- Lightweight and easy to deploy with Docker

## Screenshots

<img src="img/dark.png" width="500px">

<img src="img/light.png" width="500px">

## Configuration

goDash is configured using environment variables. Below is an example `.env` file:

```
TZ=Europe/Berlin
TITLE=My Dashboard
PUBLIC_URL=https://mydashboard.example.com
PORT=4000
LOCATION_LATITUDE=48.7803
LOCATION_LONGITUDE=9.1780
WEATHER_KEY=your_openweather_api_key
WEATHER_UNITS=metric
WEATHER_LANG=en
WEATHER_DIGITS=false
```

### Legend

- `TZ` - Time zone (e.g., `Europe/Berlin`)
- `TITLE` - Title of the dashboard
- `PUBLIC_URL` - Publicly accessible URL
- `PORT` - Port on which the service runs
- `LOCATION_LATITUDE` / `LOCATION_LONGITUDE` - Coordinates for weather data
- `WEATHER_KEY` - API key for weather service
- `WEATHER_UNITS` - Units for weather (metric/imperial)
- `WEATHER_LANG` - Language for weather information
- `WEATHER_DIGITS` - Display digits in weather data (true/false)

## Docker

### run command

```sh
docker run -d \
  --name godash \
  --restart always \
  --env-file .env \
  -v ./storage:/app/storage \
  ghcr.io/flohoss/godash:latest
```

### compose file

```yaml
services:
  godash:
    restart: always
    image: ghcr.io/flohoss/godash:latest
    container_name: godash
    env_file:
      - ./.env
    volumes:
      - ./storage:/app/storage
```

## Example bookmarks.yaml

The page is configured by a simple file. Icons can be stored in a folder called icons or downloaded from [https://selfh.st/icons/](https://selfh.st/icons/) with the prefix `sh/`.

```yml
links:
  - category: "Code"
    entries:
      - name: "Github"
        url: "https://github.com"

applications:
  - category: "Code"
    entries:
    - name: "GitHub"
      icon: "sh/github"
      ignore_dark: true # does not use the light icon even though it exists in dark mode
      url: "https://github.com"
    - name: "Home Assistant"
      icon: "sh/home-assistant"
      url: "https://www.home-assistant.io/"`
```

## ✨ Star History

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=flohoss/godash&type=Date&theme=dark" />
  <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=flohoss/godash&type=Date" />
  <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=flohoss/godash&type=Date" />
</picture>

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/flohoss/godash/blob/main/LICENSE) file for details.

## Contributing

Feel free to open issues or submit pull requests to improve goDash!
