<div align="center">

[![CI](https://github.com/flohoss/godash/actions/workflows/ci.yaml/badge.svg)](https://github.com/flohoss/godash/actions/workflows/ci.yaml)
[![Coverage](https://raw.githubusercontent.com/wiki/flohoss/godash/coverage.svg)](https://github.com/flohoss/godash/wiki/coverage.html)

A self-hosted dashboard written in Go. See the weather forecast, monitor system resources in real time, and keep your bookmarks with icons and links in one place — all driven by a single YAML config file.

</div>

## Table of Contents

- [Quick Start](#quick-start)
- [Features](#features)
- [Configuration](#configuration)
  - [Top-level settings](#top-level-settings)
  - [Weather](#weather)
  - [Applications](#applications)
  - [Links](#links)
  - [Icons](#icons)
- [Screenshots](#screenshots)
- [Development](#development)
- [License](#license)

## Quick Start

Just run the container — GoDash creates a default `config.yaml` with example data on first start if none exists:

```sh
docker run -it --rm \
  --name godash \
  -p 8156:8156 \
  -v ./config/:/app/config/ \
  ghcr.io/flohoss/godash:latest
```

Open <http://localhost:8156> to see your dashboard. Edit `./config/config.yaml` to add bookmarks or set your location — changes are picked up automatically (no restart needed).

Or with Docker Compose:

```yml
services:
  godash:
    image: ghcr.io/flohoss/godash:latest
    restart: always
    ports:
      - '8156:8156'
    volumes:
      - ./config/:/app/config/
```

The container runs as an unprivileged user (`PUID`/`PGID`, defaults to `1000:1000`) and exposes the dashboard on port `8156`.

## Features

- **Weather** — current conditions and an hourly forecast, pulled from [Open-Meteo](https://open-meteo.com/) with no API key required. Choose between Celsius and Fahrenheit.
- **Realtime system status** — CPU, RAM, and disk usage streamed live to the browser over Server-Sent Events (SSE).
- **Bookmarks** — applications and links organized into categories, each with a name, optional icon, and URL.
- **Dark / light icons** — each bookmark can have both a dark and a light icon, automatically picked based on the active theme.
- **Single config file** — everything is defined in `config.yaml`, validated at startup and hot-reloaded on change.
- **Lightweight & easy to deploy** — a single static binary served from a minimal Alpine container.

## Configuration

GoDash reads `./config/config.yaml` by default and creates one with example values on first run. The full default config lives in [`config/config.yaml`](config/config.yaml).

```yaml
log_level: 'info' # debug | info | warn | error
time_zone: 'Europe/Berlin' # Must be a valid IANA timezone (e.g., America/New_York, Europe/London)
title: 'My Dashboard' # Any string

server:
  address: '0.0.0.0' # Valid IPv4 address, defaults to 0.0.0.0
  port: 8156 # Must be between 1024 and 65535, defaults to 8156

weather:
  units: 'celsius' # celsius | fahrenheit
  latitude: 52.5163 # Valid latitude (-90 to 90)
  longitude: 13.3776 # Valid longitude (-180 to 180)

applications:
  - category: 'Productivity'
    entries:
      - name: 'Notion'
        icon: 'notion.png'
        ignore_dark: false # Keep using the same icon even if a light variant exists
        url: 'https://www.notion.so'
      - name: 'Slack'
        icon: 'sh/slack' # Automatically downloaded from selfh.st
        ignore_dark: true
        url: 'https://slack.com'

links:
  - category: 'Social'
    entries:
      - name: 'GitHub'
        url: 'https://github.com'
```

### Top-level settings

| Field       | Required | Description                                                                       |
| ----------- | -------- | --------------------------------------------------------------------------------- |
| `log_level` | no       | Log verbosity. `debug` \| `info` \| `warn` \| `error`.                            |
| `time_zone` | no       | IANA timezone used for weather and displayed times.                               |
| `title`     | no       | Dashboard title shown in the header and browser tab.                              |
| `server`    | no       | Bind address and port. `address` must be a valid IPv4, `port` between 1024–65535. |

### Weather

| Field       | Description                                       |
| ----------- | ------------------------------------------------- |
| `units`     | `celsius` or `fahrenheit`. Defaults to `celsius`. |
| `latitude`  | Latitude of the forecast location (-90 to 90).    |
| `longitude` | Longitude of the forecast location (-180 to 180). |

### Applications

Applications are grouped into categories. Each entry can show an icon, name, and link:

| Field         | Required | Description                                                                 |
| ------------- | -------- | --------------------------------------------------------------------------- |
| `name`        | yes      | Display name shown under the icon.                                          |
| `icon`        | no       | Icon file name, path, or a `sh/`-prefixed self-hosted icon.                 |
| `url`         | no       | Link target. Must be a valid URL.                                           |
| `ignore_dark` | no       | Keep using the light icon even when a dark variant exists. Default `false`. |

### Links

Simple text bookmarks grouped into categories, without icons:

| Field      | Required | Description                           |
| ---------- | -------- | ------------------------------------- |
| `category` | no       | Grouping label for a set of links.    |
| `name`     | yes      | Link label.                           |
| `url`      | no       | Destination URL. Must be a valid URL. |

### Icons

Icons can be provided in any of three ways:

- **`sh/` prefix** — downloaded automatically from [selfh.st/icons](https://selfh.st/icons/) (e.g. `sh/slack`) and cached locally.
- **Local file** — a file name or path such as `notion.png`, resolved from the `icons` folder.
- **External URL** — a full URL, served directly without caching.

GoDash can store both a **dark** and a **light** variant of each icon. When both exist, the correct one is shown based on the active theme unless `ignore_dark` is set. Orphaned icons (no longer referenced) are cleaned up automatically when the config reloads.

## Screenshots

### Dark mode

<p align="center">
  <img src="screenshots/dark.webp" width="500" />
</p>

### Light mode

<p align="center">
  <img src="screenshots/light.webp" width="500" />
</p>

## Development

All development workflows — dev server, formatting, tests, and dependency updates — are documented in [`AGENTS.md`](AGENTS.md).

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
