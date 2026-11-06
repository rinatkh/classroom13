# Антипримеры security поверх Task API

## Заменить архитектуру новым похожим приложением

Если после auth исчезли TaskUseCase, DTO/Mapper или частичный PATCH, это уже
другой проект. Начинаем с финального classroom12, а новые user/auth/middleware добавляем к нему.

## Сохранить пароль под именем PasswordHash

```go
User{PasswordHash: input.Password}
```

Название поля не хеширует значение. Register обязан вызвать PasswordHasher.Hash.
Login использует Compare, а не сравнение двух bcrypt hashes: соль разная.

## Обрезать пароль только при регистрации

TrimSpace(password) меняет credential; тот же ввод больше не проходит login.
Нормализуем email, но пароль хешируем/сравниваем в исходном виде.

## Доверять payload без Verify

Decode JWT не подтверждает подпись или срок. RequireUser вызывает Verify,
проверяет ошибку и прерывает запрос. Алгоритм разрешён явно, exp обязателен.

## Продолжить после 401

Отсутствующий return позволяет вызвать next после отказа. Middleware tests
должны подтвердить и status, и то, что следующий Handler не вызывался.

## Общий currentUserID

Identity в глобальной переменной или поле Handler смешивает конкурентные
requests. Храним её в request.Context через приватный ключ.

## Проверять ownership в middleware

Middleware знает token, но не business Task. TaskUseCase загружает Task и
сравнивает OwnerID; Handler только сопоставляет ErrForbidden с 403.

## Принимать owner_id из JSON

Клиент сможет назначить чужого владельца. У CreateTaskRequest только title,
а OwnerID берётся из проверенной identity. Неизвестное поле JSON отклоняется.

## Потерять правила прошлого урока

Подмена UpdateInput на обычный bool ломает PATCH false/nil. Отсутствие лимита
или повторного завершения меняет бизнес-поведение. Сохраняем прежние тесты и
добавляем identity в их fixture вместо удаления проверок.

## Отдавать все записи после login

Authentication ещё не означает доступ ко всему списку. List и CountActive
ограничены owner; Get/Update/Delete проверяют конкретную запись.

## Создать пользователей без атомарной уникальности

Проверка GetByEmail → Create сама по себе допускает гонку. MemoryUserRepository
повторно проверяет email под Lock; business precheck остаётся удобной ранней проверкой.
