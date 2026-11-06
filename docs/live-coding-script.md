# Live coding 13 — добавляем security к финалу Task API 12

Исходник — финальный classroom12, архитектурный Task API предыдущего урока.
Ориентир 90 минут: не переписывайте DTO/Mapper/UseCase заново, расширяйте их.

## STEP 0. Проверяем прежний Task API

```bash
make -C "../../12 Занятие/classroom12" run
# второй терминал:
curl -i http://localhost:8080/health
curl -i http://localhost:8080/tasks
```

Ожидаем 200 без token. Откройте classroom12/internal/task/usecase.go,
classroom12/internal/httpapi/handler/task.go, classroom12/internal/repository/memory/task.go.
Спросите: «Кто сейчас может читать Task? Где остаются business rules?»
Ответ: любой клиент; правила уже в UseCase. Перед финальным запуском остановите API занятия 12.

## STEP 1. User и MemoryUserRepository

Добавьте internal/user/model.go, repository.go, errors.go и memory/user.go.
Сохраняем только PasswordHash, не plaintext. Покажите map, RWMutex, nextID
и уникальность email в Create под lock, чтобы два register не создали дубликат.

## STEP 2. PasswordHasher и Register

В user/usecase.go задайте интерфейс Hash/Compare, в auth/password.go — bcrypt.
Register нормализует email, проверяет длину пароля, вызывает Hash и Repository.
Пароль нельзя TrimSpace: вход и хешируемое значение должны совпадать с login.

```bash
go test ./internal/user -run Register -v
go test ./internal/auth -run Bcrypt -v
```

Fake Hasher проверяет сценарий; отдельный bcrypt test подтверждает реальную
сравнимость hash и соль. DTO ответа никогда не получает PasswordHash.

## STEP 3. Login и TokenIssuer

Login читает User, вызывает Compare, затем Issue(userID). Неизвестный email
и неверный password становятся одной ErrInvalidCredentials. user.UseCase
не импортирует bcrypt или JWT: их implementations придут через constructors.

```bash
go test ./internal/user -run Login -v
```

## STEP 4. JWTManager

Покажите auth/token.go: Claims с user_id/iat/exp, Issue и Verify. Verify
проверяет подпись, обязательный exp, положительный ID и allow-list HS256.
JWT подписан, но payload не зашифрован. Не объясняйте JWT как замену проверке
владения конкретной Task.

```bash
go test ./internal/auth -run JWT -v
```

Round trip проходит; expired, wrong signature/algorithm, none и missing exp
отклоняются. Тестовые часы не требуют Sleep. Ключ берём из окружения.

## STEP 5. Public register/login HTTP

Добавьте dto/auth.go, mapper/user.go, handler/auth.go. Handler переводит
JSON во входы user.UseCase, ошибки — в 400/401/409/500. User DTO публикует
только id/email. Возвращаем token от login без записи пароля в claims.

## STEP 6. Authentication middleware

В middleware/auth.go определите TokenVerifier и RequireUser. Разберите
Authorization: Bearer, Verify, 401 + return при ошибке. Не добавляйте загрузку
Task в middleware: на этом шаге мы только устанавливаем identity.

```bash
go test ./internal/httpapi/middleware -v
```

Missing/malformed/invalid credentials дают 401 и не вызывают next.

## STEP 7. Context и прежние protected routes

В authctx/context.go используйте приватный тип ключа. В router.go оставьте
health/register/login public и оберните **те же** GET/POST/PATCH/DELETE /tasks.
В handler/task.go добавьте чтение identity и userID в TaskUseCase interface.
Сохраните Decode/PathValue, DTO/Mapper, error mapping и r.Context().

Спросите: «Почему userID не поле общего Handler?» Запросы конкурентны,
identity принадлежит request. UseCase получает userID аргументом и не читает HTTP context.

## STEP 8. Task.OwnerID и authorization

Добавьте OwnerID в существующую Task Model. Создание передаёт ownerID из
identity в Repository; поле не добавляется в request DTO. List и CountActive
ограничены owner scope. Get/Update/Delete проверяют value.OwnerID в UseCase.

Сначала покажите valid token Bob: authentication уже пройдена. Затем сравните
OwnerID Task Alice с userID Bob и верните ErrForbidden. Handler сопоставит 403.

```bash
go test ./internal/task -v
```

Старые business tests сохраняются с аргументом userID. Новые проверяют owner,
запрет чужих записей, отдельный лимит и список. Не заменяйте старый PATCH
полным обновлением: nil и &false остаются различимы.

## STEP 9. Composition root и интеграция

Расширьте прежний main: User/Task Memory Repository → BcryptHasher/JWTManager
→ Auth/Task UseCase → handlers → middleware → router. Никаких бизнес-решений
в main. Подготовьте JWT_SECRET для нового учебного прогона:

```bash
export JWT_SECRET="$(openssl rand -hex 32)"
make ci
make run
# второй терминал:
make curl-demo
```

Ожидаем register/login, Alice POST/GET Task, Bob GET/PATCH/DELETE → 403,
Bob list → [], Alice partial PATCH → 200, repeated completion → 409,
done:false → 200, DELETE → 204, deleted GET → 404. Без token /tasks → 401.

## STEP 10. Возвращаемся к архитектуре 12

Откройте прежние dto/task.go, mapper/task.go, handler/response.go и UseCase.
Спросите: «Что добавилось? Что осталось?» Ответ: identity/security добавились,
архитектурные границы, Task, PATCH, business rules и Memory Repository остались.
Покажите исходный classroom12 и continuity.md, затем тесты прошлого урока с auth fixture.

Финал: «Теперь архитектура и security готовы, но persistence всё ещё временный.
На следующем занятии заменим Memory Repository на PostgreSQL». Не начинайте
работу с внешней базой внутри этого урока.
