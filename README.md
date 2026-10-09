# Game Backlog API

RESTful API service for video game catalog management and personal game backlog tracking built with Go (Fiber v2) and PostgreSQL.

#### Currently Available Features

* Authentication & security using JWT access tokens and database-backed refresh tokens with rotation and replay attack detection
* Role-Based Access Control (RBAC) with database tables (`roles`, `permissions`, `role_permissions`) and in-memory cache lookup
* Video game catalog management with public/curated access and `Accept: text/csv` streaming export
* Personal user backlog tracking with playtime, ratings, duplicate prevention (409 Conflict), and ownership verification
* High-performance cursor-based pagination using Base64 tokens on composite `(created_at, id)` keys
* Input validation using `validator/v10` tags (`omitnil` for PATCH) and centralized error envelopes

---

#### How to Initialize

##### Prerequisites
* Go (version 1.23 or newer recommended)
* Git
* PostgreSQL (version 14 or newer)
* curl / Postman

##### Database Setup
1. Ensure the PostgreSQL service is active and accessible.
2. Create the target database (e.g., go_game_backlog):
   CREATE DATABASE go_game_backlog;
3. Execute the migration scripts:
   ```
   psql -U postgres -d go_game_backlog -f migrations/001_create_baseline.sql
   psql -U postgres -d go_game_backlog -f migrations/002_auth.sql
   psql -U postgres -d go_game_backlog -f migrations/003_rbac.sql
   psql -U postgres -d go_game_backlog -f migrations/004_user_permissions.sql
   psql -U postgres -d go_game_backlog -f migrations/005_cursor_index.sql
   ```

##### Database Schema
```sql
CREATE TABLE IF NOT EXISTS roles (
    name        VARCHAR(20) PRIMARY KEY,
    description VARCHAR(150) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    name        VARCHAR(50) PRIMARY KEY,
    description VARCHAR(150) NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_name       VARCHAR(20) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(50) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(50) NOT NULL UNIQUE,
    email      VARCHAR(100) NOT NULL UNIQUE,
    password   TEXT NOT NULL DEFAULT '',
    role       VARCHAR(20) NOT NULL DEFAULT 'user' REFERENCES roles(name) ON UPDATE CASCADE,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE TABLE IF NOT EXISTS games (
    id           SERIAL PRIMARY KEY,
    title        VARCHAR(150) NOT NULL,
    platform     VARCHAR(50) NOT NULL,
    genre        VARCHAR(50) NOT NULL,
    release_year INTEGER NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_games_title_lower ON games (LOWER(title));
CREATE INDEX IF NOT EXISTS idx_games_genre ON games (genre);
CREATE INDEX IF NOT EXISTS idx_games_platform ON games (platform);
CREATE INDEX IF NOT EXISTS idx_games_created_at_id_desc ON games (created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS backlogs (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_id      INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'backlog',
    hours_played INT NOT NULL DEFAULT 0,
    rating       INT NULL,
    notes        TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_game UNIQUE (user_id, game_id)
);

CREATE INDEX IF NOT EXISTS idx_backlogs_user_id ON backlogs (user_id);
```

##### Environment Configuration
Create a .env file in the root of the project by copying .env.example:
```
cp .env.example .env
```

Fill in the environment variables according to your local setup:
```
APP_PORT=3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=go_game_backlog
DB_SSLMODE=disable
DB_MAX_CONNS=10

JWT_SECRET=super_secret_key_at_least_32_chars
JWT_ISSUER=go_game_backlog
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
ALLOWED_ORIGINS=http://localhost:5173
```

##### Installation & Execution
1. Clone the repository:
   ```
   git clone https://github.com/Thaar1q/go-game-backlog.git
   cd go-game-backlog
   ```

2. Download and verify dependencies:
   ```
   go mod tidy
   ```

3. Run the application:
   ```
   go run .
   ```

4. Run Unit Tests:
   ```
   go test -v ./app/service/...
   ```

---

#### API Contract
Base URL: http://localhost:3000/api/v1

| Method | Endpoint | Query / Path / Header Params | Request Body (JSON) | Possible Statuses | Response Body Example |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/health` | *(None)* | *(None)* | 200 OK<br>503 Service Unavailable | `{"success":true,"message":"server is running","data":{"timestamp":"..."}}` |
| `POST` | `/api/v1/auth/register` | *(None)* | `{"username":"user1","email":"user@test.com","password":"Password123!"}` | 201 Created<br>400 Bad Request<br>409 Conflict<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"registration successful","data":{"id":1,"username":"user1","email":"...","role":"user","is_active":true,"created_at":"..."}}`<br>Header: `Location: /api/v1/users/1` |
| `POST` | `/api/v1/auth/login` | *(Rate Limited: 5/min)* | `{"email":"user@test.com","password":"Password123!"}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>415 Unsupported Media Type<br>422 Unprocessable Entity<br>429 Too Many Requests | `{"success":true,"message":"login successful","data":{"access_token":"...","refresh_token":"...","token_type":"Bearer","expires_in":900}}` |
| `POST` | `/api/v1/auth/refresh` | *(None)* | `{"refresh_token":"..."}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>415 Unsupported Media Type | `{"success":true,"message":"token successfully refreshed","data":{"access_token":"...","refresh_token":"...","token_type":"Bearer","expires_in":900}}` |
| `POST` | `/api/v1/auth/logout` | *(None)* | `{"refresh_token":"..."}` | 200 OK<br>400 Bad Request<br>415 Unsupported Media Type | `{"success":true,"message":"logout successful"}` |
| `GET` | `/api/v1/auth/me` | `Authorization: Bearer <token>` | *(None)* | 200 OK<br>401 Unauthorized | `{"success":true,"data":{"user_id":1,"username":"user1","role":"user"}}` |
| `GET` | `/api/v1/games` | **Headers:** `Accept: application/json` or `text/csv`<br>**Cursor Params:**<br>`cursor` (string, Base64URL)<br>`limit` (int, default 10)<br>`search` (string)<br>`genre` (string)<br>`platform` (string) | *(None)* | 200 OK<br>400 Bad Request<br>406 Not Acceptable | **JSON:**<br>`{"success":true,"message":"game list retrieved successfully","data":[...],"meta":{"limit":10,"next_cursor":"...","has_more":true}}`<br><br>**CSV:**<br>`id,title,platform,genre,release_year,created_at`<br>`12,Zelda,Switch,Adventure,2023,2026-10-08T15:36:13Z` |
| `GET` | `/api/v1/games/:id` | `:id` (int, path) | *(None)* | 200 OK<br>400 Bad Request<br>404 Not Found | `{"success":true,"data":{"id":1,"title":"Elden Ring","platform":"PC","genre":"Action RPG","release_year":2022}}` |
| `POST` | `/api/v1/games` | `Authorization: Bearer <token>`<br>*(Guard: `game:create`)* | `{"title":"Game Name","platform":"PC","genre":"RPG","release_year":2024}` | 201 Created<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"game successfully created","data":{...}}`<br>Header: `Location: /api/v1/games/1` |
| `PUT` | `/api/v1/games/:id` | `Authorization: Bearer <token>`<br>*(Guard: `game:update`)*<br>`:id` (int, path) | `{"title":"Updated Game","platform":"PC","genre":"Action","release_year":2024}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"game successfully replaced","data":{...}}` |
| `PATCH` | `/api/v1/games/:id` | `Authorization: Bearer <token>`<br>*(Guard: `game:update`)*<br>`:id` (int, path) | `{"genre":"Soulslike Open-World"}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"game successfully partially updated","data":{...}}` |
| `DELETE` | `/api/v1/games/:id` | `Authorization: Bearer <token>`<br>*(Guard: `game:delete`)*<br>`:id` (int, path) | *(None)* | 204 No Content<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found | *(Empty Body)* |
| `GET` | `/api/v1/backlog` | `Authorization: Bearer <token>`<br>**Query:** `page`, `limit`, `status` | *(None)* | 200 OK<br>400 Bad Request<br>401 Unauthorized | `{"success":true,"message":"backlog list retrieved successfully","data":[...],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| `GET` | `/api/v1/backlog/:id` | `Authorization: Bearer <token>`<br>*(Guard: Ownership or `backlog:read:any`)*<br>`:id` (int, path) | *(None)* | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found | `{"success":true,"message":"backlog entry found","data":{"id":1,"user_id":1,"game_id":1,"status":"playing",...}}` |
| `POST` | `/api/v1/backlog` | `Authorization: Bearer <token>` | `{"game_id":1,"status":"backlog","hours_played":0,"rating":null,"notes":"Steam sale"}` | 201 Created<br>400 Bad Request<br>401 Unauthorized<br>409 Conflict<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"game added to backlog","data":{...}}` |
| `PUT` | `/api/v1/backlog/:id` | `Authorization: Bearer <token>`<br>*(Guard: Ownership or `backlog:update:any`)*<br>`:id` (int, path) | `{"status":"completed","hours_played":50,"rating":10,"notes":"Masterpiece"}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"data":{...}}` |
| `PATCH` | `/api/v1/backlog/:id` | `Authorization: Bearer <token>`<br>*(Guard: Ownership or `backlog:update:any`)*<br>`:id` (int, path) | `{"status":"playing","hours_played":25}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"backlog entry successfully partially updated","data":{...}}` |
| `DELETE` | `/api/v1/backlog/:id` | `Authorization: Bearer <token>`<br>*(Guard: Ownership or `backlog:delete:any`)*<br>`:id` (int, path) | *(None)* | 204 No Content<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found | *(Empty Body)* |
| `GET` | `/api/v1/users/:id` | `Authorization: Bearer <token>`<br>`:id` (int, path) | *(None)* | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>404 Not Found | `{"success":true,"data":{"id":1,"username":"user1","email":"...","role":"user","is_active":true,"created_at":"..."}}` |
| `PATCH` | `/api/v1/users/:id/role` | `Authorization: Bearer <token>`<br>*(Guard: `role:assign`, Non-Self)*<br>`:id` (int, path) | `{"role":"curator"}` | 200 OK<br>400 Bad Request<br>401 Unauthorized<br>403 Forbidden<br>404 Not Found<br>415 Unsupported Media Type<br>422 Unprocessable Entity | `{"success":true,"message":"user role successfully updated","data":{...}}` |

##### Standardized Error Envelope
```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "validasi gagal",
  "fields": {
    "password": "must be at least 8 characters"
  },
  "request_id": "f27f6aee-691c-40b4-92f5-f8e8417bdd87"
}
```
Machine-readable codes
- (400) `BAD_REQUEST`
- (401) `UNAUTHORIZED`
- (403) `FORBIDDEN`
- (404) `NOT_FOUND`
- (406) `NOT_ACCEPTABLE`
- (409) `CONFLICT`
- (415) `UNSUPPORTED_MEDIA_TYPE`
- (422) `VALIDATION_ERROR`
- (429) `TOO_MANY_REQUESTS`
- (500) `INTERNAL_ERROR`
- (503) `SERVICE_UNAVAILABLE`

---

#### References Used
* https://docs.gofiber.io/
* https://pkg.go.dev/github.com/jackc/pgx/v5
* https://www.postgresql.org/docs/
* https://www.rfc-editor.org/rfc/rfc9110
* https://pkg.go.dev/github.com/golang-jwt/jwt/v5
* https://pkg.go.dev/golang.org/x/crypto/bcrypt
* https://owasp.org/www-project-api-security/
* https://csrc.nist.gov/projects/role-based-access-control
* https://github.com/go-playground/validator
* https://use-the-index-luke.com/no-offset
