# Apollo Music Player 

A Music Player Server that allows users to upload songs and create playlists

## Tech Stack 
- Go 
- S3
- Mailtrap SMTP Mailer
- Postgressql 

## Song Routes
| Method   | Endpoint           | Description         | Auth Required |
| -------- | ------------------ | ------------------- | ------------- |
| `GET`    | `/v1/songs`        | List all songs      | ❌ No          |
| `GET`    | `/v1/songs/:id`    | Get song by ID      | ❌ No          |
| `POST`   | `/v1/songs`        | Create a new song   | ✅ Yes         |
| `POST`   | `/v1/upload/songs` | Upload a song file  | ✅ Yes         |
| `DELETE` | `/v1/songs/:id`    | Delete a song by ID | ✅ Yes         |


## Healthcheck 
| Method | Endpoint          | Description         | Auth Required |
| ------ | ----------------- | ------------------- | ------------- |
| `GET`  | `/v1/healthcheck` | Health check status | ❌ No         |


## User Routes 

| Method | Endpoint    | Description       | Auth Required |
| ------ | ----------- | ----------------- | ------------- |
| `POST` | `/v1/users` | Register new user | ❌ No         |


## Playlist Routes 

| Method   | Endpoint                                           | Description                  | Auth Required |
| -------- | -------------------------------------------------- | ---------------------------- | ------------- |
| `GET`    | `/v1/playlists/list/playlist`                      | List playlists for user      | ✅ Yes         |
| `GET`    | `/v1/playlists/show/playlist/:id`                  | Show playlist by ID          | ✅ Yes         |
| `POST`   | `/v1/playlists/create/playlist`                    | Create a new playlist        | ✅ Yes         |
| `DELETE` | `/v1/playlists/delete/playlist/:id`                | Delete a playlist by ID      | ✅ Yes         |
| `POST`   | `/v1/playlists/add/songs`                          | Add song to a playlist       | ✅ Yes         |
| `DELETE` | `/v1/playlists/remove/songs/:song_id/:playlist_id` | Remove song from a playlist  | ✅ Yes         |
| `GET`    | `/v1/playlists/show/songs/:id`                     | Show all songs in a playlist | ✅ Yes         |


## How to Run 

This Project uses AWS S3 Default Config. An AWS Config file will be needed to use the upload functionality

Default Configuration
`go run ./cmd/api --flags` 


## Flags 

| Flag Name             | Type       | Default Value                                                 | Description                                                               |
| --------------------- | ---------- | ------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `--port`              | `int`      | `4000`                                                        | API server listening port.                                                |
| `--env`               | `string`   | `"development"`                                               | Application environment. Options: `development`, `staging`, `production`. |
| `--db-dsn`            | `string`   | `postgres://apollo:password@localhost/apollo?sslmode=disable` | PostgreSQL DSN connection string.                                         |
| `--db-max-open-conns` | `int`      | `25`                                                          | Maximum number of open PostgreSQL connections.                            |
| `--db-max-idle-conns` | `int`      | `25`                                                          | Maximum number of idle PostgreSQL connections.                            |
| `--db-max-idle-time`  | `duration` | `15m`                                                         | Maximum idle time for a PostgreSQL connection (e.g., `15m`, `1h`).        |
| `--limiter-rps`       | `float64`  | `2`                                                           | Rate limiter: max requests per second per client.                         |
| `--limiter-burst`     | `int`      | `4`                                                           | Rate limiter: burst capacity.                                             |
| `--limiter-enabled`   | `bool`     | `true`                                                        | Enable or disable the rate limiter.                                       |



