# Архитектура Task API после добавления security

Сохраняется Handler → TaskUseCase interface → task.Repository interface,
реализованный MemoryTaskRepository. DTO, Mapper, error mapping, DI и правила
прошлого урока остаются на своих местах.

## Новые зависимости

AuthHandler → user.UseCase через handler.AuthUseCase interface.
user.UseCase → user.Repository + PasswordHasher + TokenIssuer.
MemoryUserRepository, BcryptHasher и JWTManager реализуют нужные контракты.
AuthMiddleware → TokenVerifier; тот же JWTManager реализует и Issue, и Verify.

`main` собирает оба Repository, оба UseCase, handlers, middleware и router.
Ни UseCase, ни Repository не создают инфраструктуру самостоятельно.

## Где заканчивается authentication

Middleware разбирает Bearer, вызывает Verify, кладёт положительный userID в
request.Context через приватный ключ authctx. При ошибке формирует 401,
WWW-Authenticate: Bearer и прекращает обработку. Следующий Handler читает
identity через authctx.UserID и передаёт её явным аргументом UseCase.

Отмена/deadline request также сохраняются: везде передаётся r.Context().
Context не становится контейнером всех зависимостей, в нём только identity
конкретного запроса. Глобальный currentUserID или поле Handler недопустимы.

## Где начинается authorization

```go
value, err := uc.repository.Get(ctx, taskID)
if err != nil { return task.Task{}, err }
if value.OwnerID != userID { return task.Task{}, task.ErrForbidden }
```

Эта проверка находится в TaskUseCase.Get; Update вызывает тот же Get до
решения об изменении, Delete — до удаления. Проверка PATCH без известных полей
по-прежнему возвращает ErrNoChanges без записи. Недопустимый вход не даёт
доступа к данным. В HTTP известный ErrForbidden превращается в 403.

Repository.Get загружает запись по ID, а business layer проверяет ownership.
List/CountActive имеют ownerID в контракте хранения, чтобы возвращать и считать
только собственные записи. Create получает доверенный ownerID от UseCase;
он не берётся из input DTO. Update не меняет OwnerID.

## Сохранённые business rules и конкурентность

Trim/empty title/no changes, partial PATCH, запрет повторного завершения и
лимит 20 активных Task не переносятся в middleware или Memory Repository.
Лимит стал per-owner. Repository mutex защищает map; changeMu одного UseCase
по-прежнему защищает CountActive → Create/Update для текущего процесса.
MemoryUserRepository проверяет уникальный email под своим mutex.

## Model не равна публичному JSON

Task получил OwnerID, но TaskResponse по-прежнему содержит id/title/done.
User содержит PasswordHash, а UserResponse — только id/email. В JWT claims
находятся user_id, iat и exp, не пароль/hash. Mapper задаёт эту границу явно.

## Замена persistence в 14

| Сохраняется | Что заменяется |
|---|---|
| task.Repository и user.Repository | concrete memory implementations |
| TaskUseCase/AuthUseCase и правила | выбор implementations в main |
| Task/Auth DTO и Mapper | создание инфраструктурных зависимостей |
| Handler, middleware, JWT verification | реализация операций хранения |

Контракты Repository используют context, Model/Input и application errors,
без типов драйвера. Это граница замены. В этом classroom нет внешнего storage.
Один процесс и один экземпляр UseCase — явно выбранный учебный runtime;
межпроцессная атомарность составных операций относится к теме persistence.
