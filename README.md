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

Before installing to make the setup simpler, it's best to get all the keys and tokens you'll need from your accounts first. Not every platform is required, only keys and tokens for the platforms you wish to stream to.

### Twitch

__OAuth Token:__
There's a simple app that safely generates oauth tokens for you found here. https://twitchapps.com/tmi/

__Live/Stream Key:__
You can grab your key directly from your twitch dashboard. https://www.twitch.tv/USERNAME/dashboard/streamkey

### Youtube

Youtube uses a refresh token in order to access chat. As long as streamroller is running (even if you're not streaming), it'll keep the token alive. If it dies, simply repeat the process to generate a new one.

__OAuth Token:__

1. Open https://developers.google.com/oauthplayground/
2. Scroll down to "Youtube Data API" and select the first item in the list _https://www.googleapis.com/auth/youtube_.
3. Authenticate with the Google account you plan on streaming with.
4. Click "Generate Token", then copy the "Refresh Token".

__Live/Stream Key:__

Your key can be found at the bottom of your Youtube Live dashboard here.

### Facebook

Facebook requires you to stream from a Facebook page rather then a personal profile. If you don't already have one, make one first. https://www.facebook.com/pages/create/

You'll also need to create a developer app. Head over too https://developers.facebook.com/apps/, click "Add a New App", and fill in the blanks. It doesn't matter what you name it, it won't been seen by others.

__OAuth Token:__

1. Open https://developers.facebook.com/tools/explorer/
2. From the Application menu on the top right (the first item is Graph API Explorer), select the app you created previously.
3. Click "Get Token", and then "Get User Access Token".
4. Select "manage_pages" from the list and click "Get Access Token".
5. Click the "i" icon at the beginning of your token and then "Open in Access Token Tool".
6. Click "Extend Access Token" at the bottom, and copy your newly generated token. Keep note of when it expires.

__Live/Stream Key:__

A live/stream key for Facebook unfortunately changes every time you create a new Live video. You'll have to update this key each time you stream. You can follow this guide to get your started with streaming to Facebook, it includes generating your Live key. https://www.facebook.com/facebookmedia/get-started/live


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
curl -Ls "https://github.com/dustinblackman/streamroller/releases/download/0.0.1/streamroller-linux-amd64-0.0.1.tar.gz" | tar xz -C /usr/local/bin/
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

  "json": false, // Output logs in JSON
  "port": 8080, // Port for server to listen on
  "verbose": false, // Verbose logging
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
