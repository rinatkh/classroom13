# Security-модель учебного Task API

Задача урока — показать authentication/authorization в архитектуре прошлого
занятия. Используется access JWT в Authorization: Bearer, без refresh tokens,
сессий и отдельного механизма отзыва. Конечный TTL ограничивает время действия;
наличие exp само по себе не делает возможным немедленный logout/revocation.

## Что проверяется

- Register сохраняет bcrypt hash; пароль и hash не входят в response DTO/claims.
- Password сохраняет пробелы, допустимая длина входа 8–72 байта.
- JWT Verify требует подпись HS256, exp, подходящий срок и положительный user_id.
- Middleware передаёт только проверенную identity в request.Context.
- TaskUseCase проверяет владение конкретной Task; списки и лимиты изолированы.
- Body ограничен 1 MiB, неизвестные JSON поля отклоняются.

JWT_SECRET не хранится в коде и не выводится в logs; сервер требует минимум
32 байта. Для нового прогона с пустыми memory-данными сгенерируйте свежий ключ
и повторите register/login. Иначе старый JWT может пережить сброс временных ID.

## Зависимости и чтение

JWT dependency обновлена с 5.2.1 до 5.2.2 без смены API. Это исправленная ветка
для [advisory проекта golang-jwt](https://github.com/golang-jwt/jwt/security/advisories/GHSA-mh63-6h87-95cp).
Allow-list алгоритмов описан в [документации JWT](https://golang-jwt.github.io/jwt/usage/parse/).
Ограничение bcrypt и API Compare описаны в [документации Go](https://pkg.go.dev/golang.org/x/crypto/bcrypt).

В учебных curl/test сценариях только localhost и тестовые credentials.
При будущем внешнем развёртывании TLS и управление жизненным циклом identity
будут отдельной задачей. Это не требуется для рефакторинга хранения в уроке 14.
