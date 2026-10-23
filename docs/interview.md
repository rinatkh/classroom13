# Интервью: Authentication, Authorization и Middleware

## 1. Чем authentication отличается от authorization?

Authentication определяет identity: кто отправил запрос. Authorization решает,
может ли уже известный пользователь выполнить конкретное действие.

## 2. Запрос пришёл без token. Какой status?

`401 Unauthorized`. Клиент не предоставил подходящие credentials.

## 3. Token корректный, но Note принадлежит другому пользователю. Какой status?

В contract этого проекта — `403 Forbidden`: identity известна, действие
запрещено. Некоторые API намеренно возвращают `404`, чтобы не подтверждать
существование чужого ресурса. Это решение API contract, а не универсальный
запрет.

## 4. Почему `401 Unauthorized` связан с authentication?

Название историческое. На практике `401` означает отсутствие или
недействительность credentials. Обычно response также содержит
`WWW-Authenticate`.

## 5. Что делает auth middleware?

Читает credentials из HTTP request, проверяет Bearer token, получает `userID`,
кладёт его в request context и вызывает следующий Handler. При ошибке пишет
`401` и останавливает цепочку.

## 6. Что произойдёт без `next.ServeHTTP`?

Следующий Handler не выполнится. Иногда это правильно: middleware уже вернул
ошибку. В успешной ветке забытый `next` оборвёт каждый protected request.

## 7. Почему token не проверяется в каждом Handler?

Это общая transport-задача. Копирование создаёт несколько реализаций одной
политики и повышает риск, что endpoint останется без проверки.

## 8. Почему ownership обычно не проверяет общий AuthMiddleware?

Middleware знает identity, но не знает бизнес-ресурс и правило операции.
Ownership требует загрузить Note и сравнить `OwnerID`, поэтому находится в
Note UseCase.

## 9. Почему Handler не должен сравнивать `OwnerID`?

Это бизнес-решение должно одинаково работать независимо от HTTP, например из
CLI или gRPC. Handler переводит HTTP в вызов UseCase и errors обратно в status.

## 10. Зачем передавать `userID` через request context?

Identity относится к одному request и должна пройти от middleware до Handler
вместе с отменой и deadline. Context подходит для небольших request-scoped
значений.

## 11. Почему context не является глобальным хранилищем?

Он живёт в рамках операции. В него не кладут repositories, конфигурацию или
долгоживущие services. Эти зависимости передают через constructors.

## 12. Почему context key не делают строкой?

Строки из разных пакетов могут совпасть. Private key type делает ключ
уникальным для пакета.

## 13. Что такое Bearer token?

Credential, право использования которого имеет предъявитель. JWT — один из
форматов Bearer token, но Bearer token не обязан быть JWT.

## 14. Из каких частей состоит JWT?

Header, payload и signature. Signature защищает целостность. Обычный payload не
зашифрован и доступен для чтения.

## 15. Можно ли хранить password в JWT?

Нет. Ни plaintext password, ни password hash не относятся к token claims.

## 16. Зачем проверять алгоритм подписи?

Verifier должен принимать только ожидаемый алгоритм, а не доверять значению из
непроверенного token header.

## 17. Зачем claim `exp`?

Он ограничивает время действия token. Подпись сама по себе не делает token
временным.

## 18. Почему password не расшифровывают при login?

Хранят односторонний password hash. PasswordHasher сравнивает введённое значение
с сохранённым hash.

## 19. Почему нельзя использовать обычный SHA-256 для password?

Быстрый hash упрощает массовый перебор. Password hash algorithms специально
замедляют вычисление и используют salt. В проекте применяется bcrypt.

## 20. Почему login возвращает одинаковую ошибку для отсутствующего user и
неверного password?

Чтобы response не помогал перебирать зарегистрированные email. Внутреннее
логирование может различать причины, внешний contract — нет.

## 21. Где объявлять `TokenVerifier` interface?

Рядом с потребителем, которому нужна операция `Verify`: в package middleware.
`TokenIssuer` отдельно нужен AuthUseCase. Один JWTManager неявно реализует оба.

## 22. Зачем интерфейс, если реализация JWT одна?

Он отделяет политику UseCase и middleware от JWT library и делает unit tests
простыми. Но интерфейс оправдан конкретными потребителями, а не правилом «для
каждой структуры нужен interface».

## 23. Почему JWTManager не создаётся внутри UseCase?

Тогда UseCase выбирал бы concrete library, secret и TTL. Composition root
создаёт implementation и передаёт её через constructor.

## 24. Что хранит User model?

ID, normalized email и PasswordHash. Request password нужен только как input и
не должен оставаться в сохранённой модели.

## 25. Почему response DTO отличается от User?

DTO сознательно не содержит PasswordHash. Внутренняя модель не становится
случайным публичным contract.

## 26. Зачем `sync.RWMutex` в memory repository?

`net/http` обрабатывает запросы конкурентно, а map не поддерживает безопасные
одновременные записи. Mutex защищает map и генерацию ID.

## 27. Что изменится при переходе с memory на PostgreSQL?

Repository implementation и wiring. UseCase, ownership rules, middleware и HTTP
contract могут остаться.

## 28. Что изменится при переходе с JWT на opaque session token?

Token issuer/verifier implementation и, возможно, infrastructure storage.
UseCase продолжит получать identity через contract.

## 29. Должны ли все endpoints проходить AuthMiddleware?

Нет. `/register`, `/login` и часто `/health` public. Middleware применяют к
защищённым routes.

## 30. Где проходит граница между authentication и authorization в проекте?

Middleware проверяет token и формирует identity. Note UseCase получает
`userID`, загружает Note и принимает ownership-решение. Handler связывает эти
части с HTTP.
