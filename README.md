# Classroom 13 — Authentication, Authorization и Middleware

Мы продолжаем архитектурно разделённый Notes API из прошлого занятия.
Новые обязанности добавляются в подходящие слои: HTTP middleware определяет
пользователя, context переносит identity внутри запроса, а UseCase проверяет
право работать с конкретной заметкой.

## Точка старта и результат

Рабочая версия **до authentication** находится в `before/`:

```text
Client → Handler → Note UseCase → Memory Note Repository
```

Итог classroom:

```text
POST /register ─┐
POST /login    ─┴→ AuthHandler → AuthUseCase → UserRepository
                                      │
                                      ├→ PasswordHasher (bcrypt)
                                      └→ TokenIssuer (JWT)

Protected request → AuthMiddleware → request.Context(userID)
                                      ↓
                           NoteHandler → Note UseCase
                                            ↓ ownership
                                  Memory Note Repository
```

JWT, bcrypt и HTTP остаются деталями внешних слоёв. Пакеты `user` и `note`
не импортируют `net/http` или JWT-библиотеку.

## API

Public endpoints:

| Method | Path | Успех | Назначение |
|---|---|---:|---|
| `GET` | `/health` | `200` | health check |
| `POST` | `/register` | `201` | регистрация пользователя |
| `POST` | `/login` | `200` | получение Bearer token |

Protected endpoints:

| Method | Path | Успех | Назначение |
|---|---|---:|---|
| `POST` | `/notes` | `201` | создать свою заметку |
| `GET` | `/notes` | `200` | получить только свои заметки |
| `GET` | `/notes/{id}` | `200` | получить свою заметку |
| `PUT` | `/notes/{id}` | `200` | заменить свою заметку |
| `DELETE` | `/notes/{id}` | `204` | удалить свою заметку |

Protected request содержит header:

```http
Authorization: Bearer <token>
```

## Status codes

| Ситуация | Status |
|---|---:|
| malformed JSON, неверный ID, пустой title | `400 Bad Request` |
| неправильный password, token отсутствует/повреждён/истёк | `401 Unauthorized` |
| пользователь известен, но Note принадлежит другому | `403 Forbidden` |
| Note не существует | `404 Not Found` |
| email уже зарегистрирован | `409 Conflict` |
| неожиданная внутренняя ошибка | `500 Internal Server Error` |

`401 Unauthorized` исторически называется немного странно: в этом API он
означает проблему authentication. `403 Forbidden` означает, что identity уже
известна, но действие запрещено.

## Структура

```text
classroom/
├── before/                         layered Notes API без auth
├── cmd/api/main.go                 composition root
├── internal/
│   ├── user/                       User, AuthUseCase и contracts
│   ├── note/                       Note UseCase и ownership rules
│   ├── auth/                       bcrypt и JWT implementations
│   ├── repository/memory/          map + sync.RWMutex
│   └── httpapi/
│       ├── authctx/                request-scoped userID
│       ├── dto/                    JSON request/response types
│       ├── mapper/                 Model → response DTO
│       ├── handler/                HTTP lifecycle
│       ├── middleware/             Bearer authentication
│       └── router.go               public/protected routes
├── docs/
│   ├── live-coding-script.md
│   ├── architecture.md
│   ├── interview.md
│   ├── bad-examples.md
│   └── security-notes.md
├── scripts/curl-demo.sh
└── Makefile
```

## Запуск

Нужен Go 1.22+. PostgreSQL, Docker и migrations не нужны.

```bash
make run
```

`Makefile` передаёт локальный учебный secret. Для собственного запуска:

```bash
JWT_SECRET='replace-me' TOKEN_TTL='2h' go run ./cmd/api
```

Во втором терминале:

```bash
make curl-demo
```

## Ручной сценарий

Регистрация:

```bash
curl -i -X POST http://localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

Login:

```bash
curl -i -X POST http://localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

Создание заметки:

```bash
curl -i -X POST http://localhost:8080/notes \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Auth","text":"Middleware before Handler"}'
```

## Request lifecycle

Для `GET /notes/1`:

1. Router выбирает protected route.
2. AuthMiddleware читает `Authorization`.
3. JWTManager проверяет подпись и `exp`.
4. Middleware кладёт `userID` в новый request context.
5. Handler получает `userID` через `authctx.UserID`.
6. Note UseCase загружает Note и проверяет `OwnerID == userID`.
7. Handler переводит результат в HTTP status и JSON.

Если middleware вернул `401`, Handler не вызывается. Если token корректен, но
Note чужая, UseCase возвращает `note.ErrForbidden`, а Handler переводит её в
`403`.

## Тесты

```bash
make test
make test-unit
make test-race
make vet
make ci
make check-before
```

- AuthUseCase tests используют fake Repository, fake Hasher и fake TokenIssuer.
- Note UseCase tests проверяют ownership без HTTP и JWT.
- Middleware tests используют fake TokenVerifier.
- Router test выполняет register → login → protected request → forbidden.
- Memory Repository test запускает конкурентные записи; `-race` не должен
  находить гонки.

## Live coding

Откройте `docs/live-coding-script.md`. Основной путь:

```text
User → Register → Login → Token → Middleware → Context
→ protected routes → OwnerID → ownership → tests
```

## Частые ошибки

- пароль сохраняется в `User.PasswordHash` без hashing;
- `PasswordHash` попадает в response DTO или token payload;
- token проверяется заново в каждом Handler;
- middleware вызывает `next` даже после `401`;
- context key — обычная строка;
- middleware пытается проверять ownership без загрузки Note;
- UseCase возвращает `http.StatusForbidden`;
- Handler сравнивает `note.OwnerID` сам;
- memory repository пишет в map без mutex;
- JWT secret захардкожен в production-коде.

Разбор находится в `docs/bad-examples.md` и `docs/security-notes.md`.
