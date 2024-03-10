# Mastodon applications portal

~Service deployed to [apps.3615.computer](https://apps.3615.computer/) if you want to take a look.~
> 😿 The service development has been stopped and the application is not deployed anymore on 3615.computer.
>
> It never got the traction I would have wanted, I don't think it fits people needs.
>
> The simple blogging system was nice so maybe it could be refined and focus the project on this part.

**You can deploy this for your own instances and your own users!.**

## Screenshots

### Homepage

| Light Version                                                                                                                                            | Dark Version                                                                                                                                             |
| -------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ![Screenshot 2024-03-10 at 11-03-24 3615 computer - Home](https://github.com/3615-computer/portal/assets/152620834/f14fdf7e-686a-4e25-978c-a9a29014a0a4) | ![Screenshot 2024-03-10 at 11-03-57 3615 computer - Home](https://github.com/3615-computer/portal/assets/152620834/a069b52e-60a3-490e-a1ef-e2a9c9c7ef32) |

### Minecraft Home

| Light Version                                                                                                                                                 | Dark Version                                                                                                                                                  |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ![Screenshot 2024-03-10 at 11-03-30 3615 computer - Minecraft](https://github.com/3615-computer/portal/assets/152620834/c7898280-48e9-4cf0-ade0-0ac442cadd00) | ![Screenshot 2024-03-10 at 11-04-03 3615 computer - Minecraft](https://github.com/3615-computer/portal/assets/152620834/91a21dbb-281e-46b6-ab75-53862ceb927e) |

### Miniblog Home

| Light Version                                                                                                                                                | Dark Version                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ![Screenshot 2024-03-10 at 11-03-39 3615 computer - Miniblog](https://github.com/3615-computer/portal/assets/152620834/8692ea95-9bd3-4731-b5f6-4f3044c5c22e) | ![Screenshot 2024-03-10 at 11-04-13 3615 computer - Miniblog](https://github.com/3615-computer/portal/assets/152620834/9410ff77-0578-4099-a2c9-1a534dc1f06b) |

### Miniblog post

| Light Version                                                                                                                                                                           | Dark Version                                                                                                                                                                            |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|![Screenshot 2024-03-10 at 11-04-39 3615 computer - Welcome to Miniblog! – Alyx (@alyx)](https://github.com/3615-computer/portal/assets/152620834/35a09454-2241-4c5e-a34a-3382dce66ec4)| ![Screenshot 2024-03-10 at 11-04-26 3615 computer - Welcome to Miniblog! – Alyx (@alyx)](https://github.com/3615-computer/portal/assets/152620834/3d2dc8b7-0aed-4b44-a0ef-de33989b2555) |

## Services used:

- [Mastodon](https://joinmastodon.org/) (for signing in with OAuth2)
- [Exaroton](https://exaroton.com/)

## Services available

- **Minecraft Servers**: Let your instance users play on your Minecraft Servers using [Exaroton](https://exaroton.com/)
- **Miniblog**: A mini blogging system for your instance users. Markdown available.

# Config

## Environment variables

| Name                    | Example                     | Notes                                                                 |
| ----------------------- | --------------------------- | --------------------------------------------------------------------- |
| `APP_BASE_URL`          | `http://apps.3615.computer` | Without a trailing slash                                              |
| `BIND_ADDRESS`          | `0.0.0.0:3000`              | Bind application server to this IP and port                           |
| `DATABASE_PATH_CACHE`   | `./db/cache.sqlite3`        | Path to the cache database file                                       |
| `DATABASE_PATH_SESSION` | `./db/session.sqlite3`      | Path to the session database file                                     |
| `DATABASE_PATH`         | `./db/db.sqlite3`           | Path to the main database file                                        |
| `EXAROTON_API_KEY`      | `XXX`                       | Get it on https://exaroton.com/account/                               |
| `EXAROTON_SERVERS_ID`   | `XXX` or `XXX,YYY`          | Split IDs with a comma, without the `#`                               |
| `MASTODON_URL`          | `https://3615.computer/`    | With a trailing slash                                                 |
| `OAUTH2_CLIENT_ID`      | `XXX`                       | Create your application: `mastodon.example.com/settings/applications` |
| `OAUTH2_CLIENT_SECRET`  | `XXX`                       | Create your application: `mastodon.example.com/settings/applications` |
| `ORG_NAME`              | `3615.computer`             | Displayed on the homepage and to name the instance in various places  |

# How to run locally?

To run the application locally:

## Docker

Just run `docker compose up`.

## Native

- Application:
  - Install golang
  - Use [`air`](https://github.com/cosmtrek/air) for auto-reload
- Tailwind:
  - Install packages: `npm ci`
  - Watch changes: `npm run build:css`
