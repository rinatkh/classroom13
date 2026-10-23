# Архитектура Notes API с authentication

## Главная граница

```text
HTTP boundary
  Router / DTO / Handler / AuthMiddleware / request.Context
                         │
                         ▼
Application
  AuthUseCase / NoteUseCase / User / Note / errors / interfaces
                         ▲
                         │ implements
Infrastructure
  bcrypt / JWT / Memory repositories
```

Application-пакеты не импортируют `net/http`, JSON DTO или JWT. `main.go`
знает concrete implementations, потому что его работа — собрать приложение.

## Authentication

Authentication отвечает на вопрос «кто отправил запрос?».

```text
Authorization header
  → AuthMiddleware
  → TokenVerifier.Verify
  → userID
  → request.Context
```

Middleware работает на HTTP boundary. При ошибке он возвращает `401` и не
вызывает следующий Handler.

## Authorization

Ownership отвечает на вопрос «может ли этот пользователь работать с этой
Note?». Ответ зависит от бизнес-данных, поэтому проверка находится в Note
UseCase:

```go
value, err := uc.repository.Get(ctx, noteID)
if err != nil {
    return note.Note{}, err
}
if value.OwnerID != userID {
    return note.Note{}, note.ErrForbidden
}
```

HTTP Handler не принимает это решение. Он только переводит
`note.ErrForbidden` в `403 Forbidden`.

## Consumer-owned interfaces

`user.UseCase` объявляет минимальные интерфейсы, которые ему нужны:

```go
type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) error
}

type TokenIssuer interface {
    Issue(userID int64) (string, error)
}
```

`httpapi/middleware` отдельно объявляет `TokenVerifier`. Один `JWTManager`
реализует оба контракта неявно. UseCase не зависит от JWT, а middleware не
зависит от login-сценария.

## Memory repositories

`map` не поддерживает конкурентную запись из нескольких goroutines. HTTP
server может выполнять запросы параллельно, поэтому repositories защищают
состояние:

- `Lock` для Create/Update/Delete;
- `RLock` для Get/List;
- копирование данных в новый slice перед возвратом списка.

Mutex остаётся внутри repository implementation. UseCase не должен управлять
им напрямую.

## Что изменится при замене деталей

| Изменение | Меняется | Обычно остаётся |
|---|---|---|
| JWT → opaque session token | `auth.JWTManager`, wiring | UseCase, Note ownership |
| memory → PostgreSQL | repository implementations, wiring | HTTP contract, UseCase |
| HTTP → gRPC | transport, DTO, auth interceptor | UseCase, Models, contracts |
| bcrypt → другой password hasher | `PasswordHasher` implementation | AuthUseCase |

Главное свойство архитектуры: детали можно заменить в composition root без
переноса HTTP или JWT внутрь бизнес-логики.
