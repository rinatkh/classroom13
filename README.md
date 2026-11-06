# Classroom 13 — identity и security для того же Task API

**Начальное состояние — финальный Classroom 12.** Сохраняем его Task, UseCase,
Repository interface, Memory Repository, DTO/Mapper, handlers, errors, тесты,
curl demo и шесть endpoints. В classroom12 ещё нет пользователей и auth.
Для сравнения открываем его настоящий каталог, а здесь работаем с одной версией.

В корне добавляем User, Register/Login, bcrypt, JWT, authentication middleware,
request context и ownership к этому же Task API. Архитектура прошлого урока
сохраняется; новая предметная модель вместо Task не создаётся.

## Быстрый старт

Нужен Go 1.22+, curl и Python 3 для curl demo. Для нового учебного прогона
создайте свежий ключ и запустите API:

```bash
export JWT_SECRET="$(openssl rand -hex 32)"
make run
# Во втором терминале из classroom13:
make curl-demo
```

JWT_SECRET обязателен, минимум 32 байта; Makefile не подставляет общий ключ.
TOKEN_TTL по умолчанию 2h, HTTP_ADDR — :8080. В .env.example перечислены
переменные для справки, файл не загружается автоматически.

Для показа старого поведения из корня classroom13 выполните
`make -C "../../12 Занятие/classroom12" run`: GET /tasks без token вернёт 200.
Остановите API занятия 12 перед запуском classroom13 на том же порту.
После добавления security тот же запрос без token вернёт 401.

Данные обоих Memory Repositories исчезают после перезапуска. Начинайте новый
учебный прогон со свежим JWT_SECRET, затем снова register/login: ключи JWT
и временные user ID не должны переживать разные наборы memory-данных.

## Прежние маршруты получили security

| Endpoint | Доступ | Успех |
|---|---|---:|
| GET /health | public, прежний JSON {"status":"ok"} | 200 |
| POST /register | public | 201, User DTO id/email |
| POST /login | public | 200, {"token":"..."} |
| GET /tasks | authenticated user | 200, только свои Task |
| GET /tasks/{id} | owner | 200 |
| POST /tasks | authenticated user | 201 + Location |
| PATCH /tasks/{id} | owner | 200 |
| DELETE /tasks/{id} | owner | 204, пустой body |

Task JSON по-прежнему содержит **id/title/done**. OwnerID и внутренние даты
остаются в Model, Mapper их не публикует. OwnerID отсутствует во входном DTO:
пользователь не может назначить владельца через JSON. PATCH остаётся частичным;
указатели сохраняют различие «не прислали done» и «прислали false».

```bash
curl -i http://localhost:8080/tasks \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

| Ситуация | Status |
|---|---:|
| нет Bearer token или он не прошёл Verify | 401 + WWW-Authenticate: Bearer |
| token валиден, существующая Task принадлежит другому | 403 |
| Task не существует | 404 |
| неправильный JSON/ID, пустой title, PATCH без изменений | 400 |
| повторное завершение, лимит активных задач | 409 |
| email уже зарегистрирован | 409 |
| credentials для login неверны | 401 |
| непредвиденная ошибка | 500, нейтральное сообщение |

Все правила занятия 12 сохранены. Лимит 20 активных Task теперь считается
**для текущего владельца**; задачи Alice не расходуют лимит Bob. Новые 401/403
— осознанное изменение доступа, а не смена HTTP-домена или PATCH на PUT.

## Authentication и authorization — разные проверки

Authentication на HTTP boundary:

```text
Authorization: Bearer token → AuthMiddleware → JWT Verify
→ userID → request.Context → знакомый TaskHandler
```

Middleware проверяет подпись, exp, разрешённый алгоритм HS256 и identity,
возвращает 401 при ошибке и не вызывает next. Он не загружает Task.

Authorization в приложении:

```text
TaskHandler получает userID → TaskUseCase.Get/Update/Delete
→ Repository.Get(taskID) → Task.OwnerID == userID? → разрешить / ErrForbidden
```

Handler передаёт userID явным аргументом, не сравнивает владельцев сам.
UseCase не извлекает identity из HTTP context и не импортирует JWT.

## Что осталось от архитектуры 12

| Пакет | Ответственность |
|---|---|
| internal/task | Task, UseCase, Repository interface, business rules/errors |
| internal/repository/memory/task.go | прежняя map/RWMutex/nextID, добавлен owner scope |
| internal/httpapi/dto/task.go | прежние request/response DTO |
| internal/httpapi/mapper/task.go | прежний Model → DTO |
| internal/httpapi/handler/task.go | прежние HTTP-обязанности + передача userID |
| internal/user | User, Repository, AuthUseCase, PasswordHasher/TokenIssuer contracts |
| internal/auth | BcryptHasher и JWTManager implementations |
| internal/httpapi/authctx | приватный тип ключа identity в context |
| internal/httpapi/middleware | TokenVerifier contract и RequireUser |
| cmd/api/main.go | расширенный composition root |

Регистрация нормализует email, проверяет вход, хеширует пароль и сохраняет
только PasswordHash. Пароль не обрезается через TrimSpace: пробелы являются
его частью. Учебное ограничение пароля — 8–72 байта. Login сравнивает hash и
выдаёт token; неизвестный email и неверный пароль дают ErrInvalidCredentials.
MemoryUserRepository также обеспечивает уникальность email под mutex.

## Проверки

```bash
make ci
# отдельно:
go test ./...
go vet ./...
go test -race ./...
```

`make ci` проверяет единственный модуль classroom13. Сохранены business tests из 12,
его lifecycle/contract tests выполняются с authenticated owner fixture.
Добавлены ownership, per-user limit, JWT/signature/expiration/algorithm,
реальный bcrypt, middleware/context, конкурентный duplicate email и полный
register/login → Task → Bob forbidden сценарий. Fake Repository/Hasher/Issuer/
Verifier сохраняют unit-testing подход.

`make curl-demo` показывает 401, register/login Alice и Bob, создание Task,
403 для чтения/изменения/удаления чужой Task, изолированные списки, частичный
PATCH, повторное завершение 409, done:false, DELETE 204 и GET 404. Имена email
и полученный ID выбираются динамически, поэтому демо можно повторять.

Сценарий занятия: [live coding](docs/live-coding-script.md).
Границы: [architecture](docs/architecture.md).
Доказательство перехода: [continuity](docs/continuity.md).
О JWT-модели этого урока: [security notes](docs/security-notes.md).

Теперь архитектура и security готовы, но persistence всё ещё временный.
На занятии 14 заменим Memory Repository на PostgreSQL через прежние interfaces
и wiring, сохраняя HTTP/business layers. В этом занятии базы и SQL нет.

## Логи

[Уровни DEBUG / INFO / ERROR / FATAL, LOG_LEVEL и безопасные поля](docs/logging.md).

`LOG_LEVEL=DEBUG make run` включает диагностику; обычный запуск использует INFO.
