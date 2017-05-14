# Services

Instructions to get the required keys to work with your desired services.

## Twitch

__OAuth Token:__
There's a simple app that safely generates oauth tokens for you found here. https://twitchapps.com/tmi/

__Live/Stream Key:__
You can grab your key directly from your twitch dashboard. https://www.twitch.tv/USERNAME/dashboard/streamkey

## Youtube

Youtube uses a refresh token in order to access chat. As long as streamroller is running (even if you're not streaming), it'll keep the token alive. If it dies, simply repeat the process to generate a new one.

__OAuth Token:__

1. Open https://developers.google.com/oauthplayground/
2. Scroll down to "Youtube Data API" and select the first item in the list _https://www.googleapis.com/auth/youtube_.
3. Authenticate with the Google account you plan on streaming with.
4. Click "Generate Token", then copy the "Refresh Token".

__Live/Stream Key:__

Your key can be found at the bottom of your Youtube Live dashboard here.

## Facebook

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
