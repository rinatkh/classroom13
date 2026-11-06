# Continuity: classroom12 → classroom13

В этом classroom один Go-модуль и одна рабочая версия. Код приложения находится
в cmd/ и internal/, тесты — рядом с ним и в test/. Предыдущее состояние открываем
в настоящем classroom12, без вложенного дубликата проекта.

Для демонстрации старого поведения из корня classroom13:

```bash
make -C "../../12 Занятие/classroom12" run
```

У предыдущего занятия свои env и инструкция запуска. Остановите его API перед
запуском classroom13 на том же порту. Для обычной работы и make ci предыдущий
каталог не требуется. В одном развиваемом репозитории этапы можно сохранять Git commits.

## Изменения занятия 13

Сохраняются Task, UseCase, Repository, MemoryTaskRepository, DTO, Mapper,
error mapping, tests и PATCH. Добавляются User/AuthUseCase, MemoryUserRepository,
PasswordHasher/BcryptHasher, JWTManager, AuthMiddleware и auth context.
Task.OwnerID назначается из аутентифицированного userID. List возвращает свои
Task; Get/Update/Delete проверяют ownership в TaskUseCase, чужому пользователю — 403.
Middleware проверяет токен и передаёт identity, но не решает ownership.

Публичны /health, POST /register и POST /login; прежние Task routes защищены.
HTTP Task JSON остаётся id/title/done. Memory implementations используют mutex.
После restart users/tasks исчезают — это проблема, которую решит занятие 14.
Ни PostgreSQL, ни SQL в этом модуле нет.

Основные добавления: internal/user, internal/auth, internal/httpapi/authctx,
internal/httpapi/middleware, memory/user.go и auth DTO/Mapper/Handler.
Task model/usecase/repository расширяются owner scope; main собирает зависимости.

## Что меняет это упрощение структуры

В корне находится рабочая версия Task API с identity/security. Вложенные
копии предыдущих занятий удалены; для сравнения открываем настоящий classroom12.
TaskUseCase сохраняет business rules и проверяет OwnerID, Router регистрирует
прежние /tasks с PATCH, а main соединяет эти компоненты с auth-зависимостями.
Homework не меняется. DEBUG/INFO/ERROR/FATAL и LOG_LEVEL наследуются из занятия 10.
См. [logging.md](logging.md). В уроках 12–13 storage остаётся в памяти;
PostgreSQL впервые появляется в 14, а async infrastructure — в 18.

## Наследование логов

Единый `internal/logging` появился в classroom10. DEBUG / INFO / ERROR / FATAL,
`LOG_LEVEL` и безопасные поля сохраняются в финальных версиях следующих
занятий, без вложенных копий проекта. Меняется набор наблюдаемых операций, прежний Task HTTP contract остаётся.
Подробности и показ уровней — [logging.md](logging.md).
