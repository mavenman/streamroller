# Streamroller
_Self hosted simulcasting made easy_

<a href="https://travis-ci.org/dustinblackman/streamroller"><img src="https://img.shields.io/travis/dustinblackman/streamroller.svg" alt="Build Status"></a> <a href="https://goreportcard.com/report/github.com/dustinblackman/streamroller"><img src="https://goreportcard.com/badge/github.com/dustinblackman/streamroller"></a> <img src="https://img.shields.io/github/release/dustinblackman/streamroller.svg">

[![Deploy to Docker Cloud](https://files.cloud.docker.com/images/deploy-to-dockercloud.svg)](https://cloud.docker.com/stack/deploy/?repo=https://github.com/dustinblackman/streamroller)

Streamroller is a mutlistream tool that allows you to broadcast your streams (e.g OBS/Xsplit) to multiple platforms such as Twitch, Youtube, and Facebook at once. It also has read only chat to make it easy to vocally respond to your viewers in a single window.

This tool is meant to be hosted remotely to avoid networking issues locally as trying to stream for multiple sources can get quite heavy.

## Screenshots

__Chat:__
![Chat](https://s18.postimg.org/7qh7eml7t/newscreen.png)

## Features/Roadmap
- [x] Combined chat (read)
- [ ] Combined chat (write)
- [x] RTMP cloning
- [ ] Streamroller stream key protection
- [ ] RTMPT to RTMP

#### Platforms
- [x] Twitch
- [x] Facebook Live
- [x] Youtube
- [ ] Azubu
- [ ] Hitbox

## Getting Keys/Tokens

Before installing to make the setup simpler, it's best to get all the keys and tokens you'll need from your accounts first. Not every platform is required, only keys and tokens for the platforms you wish to stream to. See the [SERVICES.md](SERVICES.md) doc.

## Install

### Docker Cloud

[![Deploy to Docker Cloud](https://files.cloud.docker.com/images/deploy-to-dockercloud.svg)](https://cloud.docker.com/stack/deploy/?repo=https://github.com/dustinblackman/streamroller)

Docker Cloud is users who are atleast somewhat familiar with working on servers. You can follow Docker Cloud's get started guide [here](https://docs.docker.com/docker-cloud/getting-started/intro_cloud/), and then afterwards hit the Deploy with Cloud button.

### Docker

A docker image is available over at [Docker Hub](https://hub.docker.com/r/dustinblackman/streamroller). It's suggested to use a [tag](https://hub.docker.com/r/dustinblackman/streamroller/tags/) rather then `latest`.

### Server

Grab the latest release from the [releases](https://github.com/dustinblackman/streamroller/releases) page, or build from source and install directly from master. Streamroller is currently built and tested against Go 1.7.

__Quick install for Linux:__
```
curl -Ls "https://github.com/dustinblackman/streamroller/releases/download/1.0.0/streamroller-linux-amd64-1.0.0.tar.gz" | tar xz -C /usr/local/bin/
```

__Build From Source:__

A makefile exists to handle all things needed to build and install from source.

```
git pull https://github.com/dustinblackman/streamroller
cd streamroller
make install
```

## Configure

Configuration can be done either by command line parameters, environment variables, or a JSON file. Please see all available flags with `streamroller --help` or in the example below.

To set a configuration, you can take the flag name and export it in your environment or save in one of the three locations for config files. If on Heroku or Docker, use the environments examples.

After launch, chat can be found on the index, (e.g `http://localhost:8080`). And streams can also be submitted to index (e.g `rtmp://localhost:8080`).

#### Parameters
```
{
  "twitch-livekey": "", // Twitch live key
  "twitch-username": "", // Twitch Username
  "twitch-oauth": "", // Twitch oauth key

  "facebook-livekey": "", // Facebook live key
  "facebook-token": "", // Facebook auth token

  "youtube-livekey": "", // Youtube live key
  "youtube-token": "", // Youtube oauth refresh token

  "port": 8080, // Port for server to listen on
  "debug": false, // Debug logging
}
```

#### Examples

__Flags:__
```
streamroller --twitch-livekey aBc123 facebook-livekey EfG456
```

__Environments:__

All flags/parameters can be used with environment variables by prefixing with `SR_`, and replacing dashes with underscores.

```
export SR_TWITCH_LIVEKEY=aBc123
export SR_FACEBOOK_LIVEKEY=EfG456
```

__JSON File:__

Configuration files can be stored in one of the three locations

```sh
./streamroller.json
/etc/streamroller.json
$HOME/.streamroller/streamroller.json
```
```json
{
  "twitch-livekey": "aBc123",
  "facebook-livekey": "EfG456"
}
```

## Credit

- Ethan for the name
- [JDoeDevel](http://bootsnipp.com/JDoeDevel) for the chat snippet

## Related Projects

- [StreamWithFriends](https://github.com/dustinblackman/streamwithfriends) - StreamWithFriends allows your friends webcams to appear on your stream all through a web broswer.

## [License](./LICENSE)
