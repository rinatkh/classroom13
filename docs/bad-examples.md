# Архитектурные ошибки

## Token parsing в каждом Handler

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    raw := r.Header.Get("Authorization")
    userID, _ := h.tokens.Verify(raw)
    // ...
}
```

Логика дублируется и легко расходится между endpoints. Общая authentication-
проверка относится к middleware.

## Middleware продолжает цепочку после 401

```go
if err != nil {
    http.Error(w, "unauthorized", http.StatusUnauthorized)
}
next.ServeHTTP(w, r)
```

Нужен `return`. Иначе protected Handler всё равно выполнится и попробует
записать второй response.

## Ownership внутри middleware

```go
func Auth(next http.Handler) http.Handler {
    // middleware загружает Note и сравнивает OwnerID
}
```

Обычный AuthMiddleware не знает, какой ресурс и какое бизнес-правило применяет
конкретный endpoint. Он устанавливает identity. Ownership проверяет UseCase,
который работает с Note.

## HTTP status внутри UseCase

```go
func (uc *UseCase) Delete(...) (int, error) {
    return http.StatusForbidden, note.ErrForbidden
}
```

UseCase начинает зависеть от HTTP. Правильнее вернуть application error, а
transport решит, каким status представить его клиенту.

## Password в plaintext

```go
repository.Create(user.User{PasswordHash: request.Password})
```

Название поля не делает значение hash. AuthUseCase сначала вызывает
`PasswordHasher.Hash`, и только результат передаёт repository.

## PasswordHash в DTO

```go
type UserResponse struct {
    PasswordHash string `json:"password_hash"`
}
```

Hash не является публичным API contract. Response mapper выбирает только ID и
email.

## Строковый context key

```go
context.WithValue(ctx, "userID", userID)
```

Другой пакет может использовать тот же ключ. Private key type устраняет
коллизию.

## Map без mutex

```go
func (r *Repository) Create(value Note) {
    r.items[value.ID] = value
}
```

Несколько HTTP requests могут записывать одновременно. Repository использует
`sync.RWMutex`; race detector проверяет это решение.

## Interface ради названия слоя

Не нужен один огромный `AuthServiceInterface`, который содержит register,
login, JWT, password hashing и HTTP helpers. Интерфейс вводится у потребителя и
описывает только реально нужное поведение.
