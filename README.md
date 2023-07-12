# Mastodon applications portal

Services supported:
- [Mastodon](https://joinmastodon.org/) (for signing in with OAuth2)
- [Exaroton](https://exaroton.com/)

# Config

## Environment variables

|Name|Example|Notes|
|----|-------|-----|
|`APP_BASE_URL`|`http://apps.3615.computer/`|With a trailing slash|
|`ORG_NAME`|`3615.computer`|Displayed on the homepage|
|`MASTODON_URL`|`https://3615.computer/`|With a trailing slash|
|`OAUTH2_CLIENT_ID`|`XXX`|create your application: `mastodon.example.com/settings/applications`)|
|`OAUTH2_CLIENT_SECRET`|`XXX`|create your application: `mastodon.example.com/settings/applications`)|
|`EXAROTON_SERVERS_ID`|`XXX` or `XXX,YYY`|Split IDs with a comma, without the `#`|
|`EXAROTON_API_KEY`|`XXX`|Get it on https://exaroton.com/account/|

# How to run locally?

- Application: `air`
- Tailwind:
    - First time: `npm ci`
    - Watching changes: `npm run build:css`
