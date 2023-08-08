# Mastodon applications portal

Services supported:

- [Mastodon](https://joinmastodon.org/) (for signing in with OAuth2)
- [Exaroton](https://exaroton.com/)

# Config

## Environment variables

| Name                    | Example                     | Notes                                                                 |
| ----------------------- | --------------------------- | --------------------------------------------------------------------- |
| `APP_BASE_URL`          | `http://apps.3615.computer` | Without a trailing slash                                              |
| `BIND_ADDRESS`          | `0.0.0.0:3000`              | Bind application server to this IP and port                           |
| `DATABASE_PATH_BLOG`    | `./db/blog.sqlite3`         | Path to the blog database file                                        |
| `DATABASE_PATH_CACHE`   | `./db/cache.sqlite3`        | Path to the cache database file                                       |
| `DATABASE_PATH_SESSION` | `./db/session.sqlite3`      | Path to the session database file                                     |
| `EXAROTON_API_KEY`      | `XXX`                       | Get it on https://exaroton.com/account/                               |
| `EXAROTON_SERVERS_ID`   | `XXX` or `XXX,YYY`          | Split IDs with a comma, without the `#`                               |
| `MASTODON_URL`          | `https://3615.computer/`    | With a trailing slash                                                 |
| `OAUTH2_CLIENT_ID`      | `XXX`                       | Create your application: `mastodon.example.com/settings/applications` |
| `OAUTH2_CLIENT_SECRET`  | `XXX`                       | Create your application: `mastodon.example.com/settings/applications` |
| `ORG_NAME`              | `3615.computer`             | Displayed on the homepage                                             |

# How to run locally?

To run the application locally:

## `docker compose`

Just run `docker compose up`.

## Native

- Application: `air`
- Tailwind:
  - First time: `npm ci`
  - Watching changes: `npm run build:css`
