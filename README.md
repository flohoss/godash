# GoDash

A blazing fast start-page for your services written in Go.

![](https://img.shields.io/badge/Language-Go-informational?style=for-the-badge&logo=go&color=00ADD8)
![](https://img.shields.io/badge/Framework-TailwindCSS-informational?style=for-the-badge&logo=tailwind-css&color=06B6D4)

[![Build Status](https://build.unjx.de/buildStatus/icon?style=flat-square&job=godash%2Fmain)](https://build.unjx.de/job/godash/job/main/)

[https://hub.docker.com/r/unjxde/godash](https://hub.docker.com/r/unjxde/godash)

## Dark Mode

<img style="border-radius:0.5rem" src="https://filedn.eu/lhdsENsife1QUzPddOpRjb5/godash/dark_v2.0.1.png" alt="dark_mode" width="600"/>

## Light Mode

<img style="border-radius:0.5rem" src="https://filedn.eu/lhdsENsife1QUzPddOpRjb5/godash/light_v2.0.1.png" alt="light_mode" width="600"/>

## Lighthouse performance test

Emulated Moto G4 with Lighthouse 9.6.6, using headless chromium 106.0.5249.103:

<img style="border-radius:0.5rem" src="https://filedn.eu/lhdsENsife1QUzPddOpRjb5/godash/lighthouse_v2.0.1.png" alt="lighthouse" width="500"/>

## How to use

Use the docker-compose to spin up the service.
The Weather is fetched over a [Current Weather Api Call](https://openweathermap.org/current) with environment variables for the needed parameters.
If you don't want to see the weather, do not provide a key as environment variable.
Please refer to the available options as shown in the docker-compose example.

### Example of the config.yaml

All Bookmarks are read from a file called `config.yaml` located inside the `./storage` folder.
The application will create a default file at startup and will automatically look for changes inside the file.
Changes are printed in stdout when running with `LOG_LEVEL=trace`.

You can specify an icon of a bookmark either by using a link or by using the name of the file located inside the `./storage/icons` folder that is mounted via the docker compose file.
The name and related link can be provided as well.

**config.yaml example:**
```yaml
links:
  - category: "Code"
    entries:
      - name: "Github"
        url: "https://github.com"
  - category: "CI/CD"
    entries:
      - name: "Jenkins"
        url: "https://www.jenkins.io/"
  - category: "Server"
    entries:
      - name: "bwCloud"
        url: "https://portal.bw-cloud.org"

applications:
  - category: "Code"
    entries:
      - name: "Github"
        icon: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
        url: "https://github.com"
  - category: ""
    entries:
      - name: "Jenkins"
        icon: "https://www.jenkins.io/images/logos/jenkins/Jenkins-stop-the-war.svg"
        url: "https://www.jenkins.io/"
  - category: "Server"
    entries:
      - name: "bwCloud"
        icon: "https://portal.bw-cloud.org/static/dashboard/img/logo-splash.svg"
        url: "https://portal.bw-cloud.org"
```

### Available environment variables with default values

```toml
PORT = 4000
ALLOWED_HOSTS = "*"
TITLE = "GoDash"

LOG_LEVEL = "info"

LOCATION_LATITUDE = 48.780331609463815
LOCATION_LONGITUDE = 9.177968320179422
WEATHER_KEY = ""
WEATHER_UNITS = "metric"
WEATHER_DIGITS = true

LIVE_SYSTEM = true
```

## Heartbeat

```/health```

Heartbeat endpoint can be useful to setting up a load balancers or an external uptime testing service that can make a request before hitting any routes.

## A docker-compose example:

```yaml
version: '3.9'

services:

  godash:
    image: unjxde/godash:latest
    container_name: godash
    restart: unless-stopped
    environment:
      # https://docs.linuxserver.io/general/understanding-puid-and-pgid
      - PUID=1000
      - PGID=1000
      - TZ=Europe/Berlin
      # allowed hosts for cors, seperated by comma
      - ALLOWED_HOSTS=https://home.example.com,https://another.example.com
      # change title to something else
      - TITLE=GoDash
      # available log-levels: debug,info,warn,error,panic,fatal
      - LOG_LEVEL=info
      # create account here to get free key:
      # https://home.openweathermap.org/users/sign_up
      # remove to disable weather
      - WEATHER_KEY=thisIsNoFunctioningKey
      # standard, metric or imperial
      - WEATHER_UNITS=metric
      # Temp is normally xx.xx, can be rounded to xx if desired
      - WEATHER_DIGITS=true
      # location is needed for weather
      - LOCATION_LATITUDE=48.644929601442485
      - LOCATION_LONGITUDE=9.349618464869025
      # show live system information
      - LIVE_SYSTEM=true
    volumes:
      # to mount the config.yaml and the icons folder on the system
      - ./storage:/app/storage
    # https://docs.docker.com/compose/compose-file/compose-file-v3/#ports
    ports:
      - "4000:4000"
```
