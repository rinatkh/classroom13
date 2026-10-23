# Live coding: добавляем auth в существующий Notes API

Ориентир — 45–55 минут практики. Исходное состояние находится в `before/`, а
готовый код — в корне classroom. Не переписывайте всё сразу: после каждого шага
фиксируйте новую ответственность и запускайте узкую проверку.

## Подготовка — 3 минуты

### Что открыть

- `before/internal/note/note.go`;
- `before/internal/httpapi/router.go`;
- финальную схему в `docs/architecture.md`.

### Что спросить

«Кто сейчас может вызвать `GET /notes`?»

### Ожидаемый ответ

Любой клиент. В системе нет пользователя, credentials или ownership.

### Команда

```bash
make check-before
cd before && go run ./cmd/api
```

Во втором терминале:

```bash
curl -i http://localhost:8080/notes
```

### Expected result

`200 OK` без token. Подчеркните: архитектура уже разделена, но security
responsibilities пока отсутствуют.

---

## STEP 1. User model и Repository — 5 минут

### Что открыть

- `internal/user/model.go`;
- `internal/user/repository.go`;
- `internal/repository/memory/user.go`.

### Что спросить

«Почему `Password` из request не должен становиться полем сохранённого User?»

### Ожидаемый ответ

Repository хранит только hash. Plaintext password нужен на короткое время во
входе Register/Login.

### Что написать

1. `User` с `ID`, `Email`, `PasswordHash`.
2. `RegisterInput` и `LoginInput`.
3. `Repository` с `Create` и `GetByEmail`.
4. `MemoryUserRepository` с `map[string]user.User` и `sync.RWMutex`.

### Что временно сломано

Пока нет AuthUseCase, поэтому User создать через HTTP нельзя. Это нормально:
мы сначала формируем внутренний contract.

### Команда

```bash
go test ./internal/repository/memory
```

### Expected result

Package компилируется. Memory implementation знает `map` и mutex, package
`user` — нет.

### Связь со следующим шагом

Теперь UseCase сможет попросить Repository сохранить пользователя, но сначала
нужно определить, как хешировать password.

---

## STEP 2. PasswordHasher и Register — 6 минут

### Что открыть

- `internal/user/usecase.go`;
- `internal/auth/password.go`;
- `internal/user/usecase_test.go`.

### Что спросить

«Почему AuthUseCase принимает `PasswordHasher`, а не вызывает bcrypt напрямую?»

### Ожидаемый ответ

UseCase зависит от нужного поведения. bcrypt — заменяемая implementation и не
нужен в unit test.

### Что написать

```go
type PasswordHasher interface {
    Hash(password string) (string, error)
    Compare(hash, password string) error
}
```

Затем `Register`:

1. normalize email;
2. проверить email/password;
3. убедиться, что email свободен;
4. вызвать `Hash`;
5. сохранить только `PasswordHash`.

### Classroom moment

Временно запишите:

```go
PasswordHash: input.Password
```

Спросите, изменилось ли значение только из-за имени поля. Сразу замените на
результат `hasher.Hash`.

### Команда

```bash
go test ./internal/user -run Register -v
```

### Expected result

Тест подтверждает normalized email и то, что plaintext не передан Repository.

### Связь со следующим шагом

Регистрация создаёт identity, а login должен проверить credentials и выдать
token.

---

## STEP 3. Login и одинаковая внешняя ошибка — 5 минут

### Что открыть

- `internal/user/usecase.go`;
- `internal/user/errors.go`.

### Что спросить

«Должен ли клиент различать “email не найден” и “password неверный”?»

### Ожидаемый ответ

Нет. Оба случая превращаются в `ErrInvalidCredentials`, чтобы API не помогал
определять зарегистрированные email.

### Что написать

1. `GetByEmail`.
2. `hasher.Compare(value.PasswordHash, input.Password)`.
3. `tokens.Issue(value.ID)`.
4. `LoginResult{Token: token}`.

### Что временно сломано

Интерфейс `TokenIssuer` уже есть, но concrete TokenManager появится на
следующем шаге. Unit test использует fake.

### Команда

```bash
go test ./internal/user -run Login -v
```

### Expected result

Правильный password возвращает token; неверный password и отсутствующий user —
одинаковый `ErrInvalidCredentials`.

---

## STEP 4. JWTManager — 6 минут

### Что открыть

- `internal/auth/token.go`;
- `internal/auth/token_test.go`.

### Что спросить

«Можно ли считать payload JWT секретным?»

### Ожидаемый ответ

Нет. Подпись защищает целостность, но payload можно декодировать. В claims нет
password или hash.

### Что написать

```go
type Claims struct {
    UserID int64 `json:"user_id"`
    jwt.RegisteredClaims
}
```

В `Issue` добавить `user_id`, `iat`, `exp`. В `Verify`:

- разобрать token с claims;
- разрешить только HS256;
- потребовать `exp`;
- проверить `token.Valid` и положительный `UserID`.

### Что специально временно сломать

Уберите `jwt.WithValidMethods`. Спросите: «Почему verifier не должен выбирать
алгоритм только по непроверенному header?» Верните allow-list.

### Команда

```bash
go test ./internal/auth -v
```

### Expected result

Round trip возвращает тот же `userID`; expired token отклоняется.

---

## STEP 5. Register/Login HTTP — 5 минут

### Что открыть

- `internal/httpapi/dto/auth.go`;
- `internal/httpapi/handler/auth.go`;
- `internal/httpapi/mapper/user.go`.

### Что спросить

«Почему нельзя encode внутренний `user.User` напрямую?»

### Ожидаемый ответ

Model содержит `PasswordHash`. Response DTO сознательно публикует только ID и
email.

### Что написать

- request DTO для register/login;
- response DTO для User/token;
- mapping application errors в `400`, `401`, `409`, `500`;
- `Location` перед `writeJSON` при successful register.

### Команда

```bash
go test ./internal/httpapi/... -run Register -v
```

Если узкого test ещё нет, достаточно:

```bash
go test ./internal/httpapi/...
```

### Expected result

Transport compiles, PasswordHash не входит в response type.

---

## STEP 6. Проблема дублирования и AuthMiddleware — 7 минут

### Что открыть

- `internal/httpapi/middleware/auth.go`;
- `internal/httpapi/middleware/auth_test.go`.

### Сначала показать плохой вариант

Напишите в двух handlers:

```go
raw := r.Header.Get("Authorization")
userID, err := tokens.Verify(raw)
```

Спросите, сколько мест придётся исправлять при изменении формата credentials.
Удалите дублирование.

### Что написать

```go
func (m *Auth) RequireUser(next http.Handler) http.Handler
```

Внутри:

1. прочитать header;
2. разобрать `Bearer <token>`;
3. вызвать `TokenVerifier.Verify`;
4. при ошибке вернуть `401` и `return`;
5. передать `userID` через context;
6. вызвать `next.ServeHTTP`.

### Что специально временно сломать

Удалите `return` после `unauthorized(w)`. Спросите, почему next Handler всё
равно выполнится и почему response станет некорректным. Верните `return`.

### Команда

```bash
go test ./internal/httpapi/middleware -v
```

### Expected result

Missing/wrong/invalid credentials дают `401`; valid token вызывает next.

---

## STEP 7. Context и protected routes — 5 минут

### Что открыть

- `internal/httpapi/authctx/context.go`;
- `internal/httpapi/router.go`;
- `internal/httpapi/handler/note.go`.

### Что спросить

«Почему не передать userID через global variable или поле Handler?»

### Ожидаемый ответ

Identity относится к конкретному конкурентному request. Глобальное поле будут
перезаписывать другие запросы.

### Что написать

```go
type userIDKey struct{}
```

Затем `WithUserID` и `UserID`. В Router оставить register/login public, а каждый
`/notes` Handler обернуть `authMiddleware.RequireUser`.

### Команда

```bash
go test ./internal/httpapi/... -v
```

### Expected result

`GET /notes` без token возвращает `401`. Register/Login остаются доступными.

---

## STEP 8. OwnerID и authorization в UseCase — 6 минут

### Что открыть

- `internal/note/model.go`;
- `internal/note/usecase.go`;
- `internal/note/usecase_test.go`.

### Что спросить

«Token уже валиден. Может ли Bob теперь удалить Note Alice?»

### Ожидаемый ответ

Authentication завершена, но authorization ещё нужна. UseCase должен сравнить
`value.OwnerID` и `userID`.

### Что написать

1. `OwnerID` в Note.
2. `Create(ctx, userID, input)` устанавливает owner из identity, а не из JSON.
3. `ListByOwner` возвращает только собственные Notes.
4. Get/Update/Delete загружают Note и возвращают `ErrForbidden` при mismatch.

### Classroom moment

Временно перенесите `OwnerID` в `CreateNoteRequest`. Спросите, сможет ли клиент
создать Note от имени другого user. Уберите OwnerID из DTO: owner приходит
только из authentication context.

### Команда

```bash
go test ./internal/note -v
```

### Expected result

Owner может менять Note; другой user получает `ErrForbidden`; list изолирован.

---

## STEP 9. Error mapping: 401 и 403 — 4 минуты

### Что открыть

- `internal/httpapi/handler/note.go`;
- `internal/httpapi/middleware/auth.go`.

### Что спросить

Дайте два сценария:

1. header отсутствует;
2. token valid, Note чужая.

### Ожидаемый ответ

Первый ответ формирует middleware: `401`. Второй error приходит из UseCase и
Handler переводит его в `403`.

### Что написать

В `writeNoteError`:

```go
case errors.Is(err, note.ErrForbidden):
    writeError(w, http.StatusForbidden, err.Error())
```

UseCase не импортирует `net/http`.

### Команда

```bash
go test ./internal/httpapi -run Ownership -v
```

### Expected result

Полный router test видит разные statuses для authentication и authorization.

---

## STEP 10. Composition root и полный lifecycle — 6 минут

### Что открыть

- `cmd/api/main.go`;
- `scripts/curl-demo.sh`.

### Что спросить

«Почему main может знать про memory, bcrypt и JWT?»

### Ожидаемый ответ

`main` — composition root. Он выбирает implementations и связывает их, но не
содержит business rules.

### Что написать

Создать:

1. UserRepository и NoteRepository;
2. BcryptHasher и JWTManager;
3. AuthUseCase и NoteUseCase;
4. Handlers и AuthMiddleware;
5. Router и Server.

### Команды

```bash
make ci
make run
```

Во втором терминале:

```bash
make curl-demo
```

### Expected result

- без token: `401`;
- Alice создаёт Note: `201`;
- Bob читает Alice Note: `403`;
- Bob list: `[]`;
- Alice update: `200`;
- Alice delete: `204`;
- повторный GET: `404`.

## Финальные вопросы — 3 минуты

1. Где заканчивается authentication?
2. Где начинается ownership authorization?
3. Что изменится при JWT → opaque token?
4. Что изменится при memory → PostgreSQL?
5. Почему `UserResponse` не содержит PasswordHash?

Ожидаемая схема ответа:

```text
Bearer token → Middleware → userID → request.Context → Handler
→ UseCase → ownership → Repository
```
