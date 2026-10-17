# Chirpy

Chirpy is a Go HTTP server for user accounts and short posts called chirps. It exposes a JSON API backed by PostgreSQL, issues JWT access tokens and refresh tokens, and includes a small static-file handler and admin endpoints.

## Requirements

- Go 1.26.5 or a later compatible version, as specified in `go.mod`
- PostgreSQL
- Goose, if you need to apply the database migrations

## Run locally

1. Create a PostgreSQL database and set `DB_URL` to its connection URL.
2. Create a `.env` file in this directory with the required settings:

   ```dotenv
   DB_URL=postgres://USER:PASSWORD@localhost:5432/DATABASE?sslmode=disable
   PLATFORM=dev
   TOKEN_SECRET=replace-with-a-long-random-secret
   POLKA_KEY=replace-with-your-polka-api-key
   ```

   `.env` is ignored by Git. Keep real credentials out of source control.

3. Apply the schema migrations:

   ```sh
   set -a
   . ./.env
   set +a
   goose -dir sql/schema postgres "$DB_URL" up
   ```

4. Start the server:

   ```sh
   go run .
   ```

The server listens on `http://localhost:8080`. Run the Go tests with `go test ./...`.

## API

Unless otherwise noted, request and response bodies use JSON. Error responses created by the API use an `error` field, for example `{"error":"Chirp is too long"}`. Some authentication and not-found responses intentionally have an empty body.

### Authentication

Use the access token returned by login for protected user and chirp operations:

```http
Authorization: Bearer <access_token>
```

Access tokens expire after 60 minutes. Refresh tokens expire after 60 days; send a refresh token as a bearer token to `/api/refresh` or `/api/revoke`.

### Endpoints

| Method   | Path                    | Authentication | Description                                                            |
| -------- | ----------------------- | -------------- | ---------------------------------------------------------------------- |
| `GET`    | `/api/healthz`          | None           | Health check; returns plain-text `OK`.                                 |
| `POST`   | `/api/users`            | None           | Create a user.                                                         |
| `POST`   | `/api/login`            | None           | Log in and receive access and refresh tokens.                          |
| `PUT`    | `/api/users`            | Access token   | Update the authenticated user's email and password.                    |
| `POST`   | `/api/chirps`           | Access token   | Create a chirp for the authenticated user.                             |
| `GET`    | `/api/chirps`           | None           | List chirps; optionally filter and sort.                               |
| `GET`    | `/api/chirps/{chirpID}` | None           | Get one chirp by UUID.                                                 |
| `DELETE` | `/api/chirps/{chirpID}` | Access token   | Delete a chirp owned by the authenticated user.                        |
| `POST`   | `/api/refresh`          | Refresh token  | Exchange a refresh token for a new access token.                       |
| `POST`   | `/api/revoke`           | Refresh token  | Revoke a refresh token.                                                |
| `POST`   | `/api/polka/webhooks`   | Polka API key  | Handle the `user.upgraded` event.                                      |
| `GET`    | `/admin/metrics`        | None           | Show the static-file request count as HTML.                            |
| `POST`   | `/admin/reset`          | None           | Reset the request count; in dev, delete all users and related records. |

#### Users and tokens

Create a user with `POST /api/users`:

```sh
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"reader@example.com","password":"correct-horse-battery"}'
```

The request fields are `email` and `password`. A successful request returns `201 Created` and a user object containing `id`, `created_at`, `updated_at`, `email`, and `is_chirpy_red`.

Log in with `POST /api/login` using the same fields. A successful response returns `200 OK`, the user fields, an access `token`, and a `refresh_token`. Invalid credentials return `401 Unauthorized`.

To update the authenticated account, send `PUT /api/users` with an access token and a JSON body containing the new `email` and `password`. A successful update returns `200 OK` with the updated user object.

Exchange a refresh token for an access token:

```sh
curl -X POST http://localhost:8080/api/refresh \
  -H 'Authorization: Bearer <refresh_token>'
```

The response is `{"token":"<new_access_token>"}`. `POST /api/revoke` accepts the same authorization header and returns `204 No Content` when the refresh token is revoked.

#### Chirps

Create a chirp with `POST /api/chirps`. The JSON body contains `body`; although a `user_id` field may be sent, the server uses the user ID in the access token instead.

```sh
curl -X POST http://localhost:8080/api/chirps \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -d '{"body":"Hello, Chirpy!"}'
```

A chirp body is limited to 140 bytes. The words `kerfuffle`, `sharbert`, and `fornax` are replaced with `****` when they appear as space-separated words, regardless of letter case. Successful creation returns `201 Created` and a chirp with `id`, `created_at`, `updated_at`, `body`, and `user_id`.

List chirps with `GET /api/chirps`. Optional query parameters:

- `author_id`: filter by the author's UUID.
- `sort=desc`: sort newest first. Without this parameter, the database's default result order is used.

For example: `GET /api/chirps?author_id=<user_uuid>&sort=desc`.

Fetch a single chirp with `GET /api/chirps/{chirpID}`. Delete one with `DELETE /api/chirps/{chirpID}` and an access token; only its author can delete it. A successful deletion returns `204 No Content`; a different user receives `403 Forbidden`.

#### Polka webhook

`POST /api/polka/webhooks` expects an authorization header whose second whitespace-separated value is the configured `POLKA_KEY`, for example `Authorization: ApiKey <POLKA_KEY>`. The JSON event body has this shape:

```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "<user_uuid>"
  }
}
```

The `user.upgraded` event marks the user as Chirpy Red. Other event names are acknowledged with `204 No Content` and do not change user data. An invalid or missing API key returns `401 Unauthorized`.

#### Admin and static files

`GET /admin/metrics` returns an HTML page showing the number of requests served through the `/app` static-file handler; API requests are not included in this count.

`POST /admin/reset` clears that count. When `PLATFORM=dev`, it also deletes all users, which cascades to their chirps and refresh tokens. This operation is destructive and should only be used with development data. Other platform values receive `403 Forbidden` and do not delete users.

The `/app` handler serves files from the project directory.
