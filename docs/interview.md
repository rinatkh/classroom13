# Вопросы после добавления security

1. **Что изменилось относительно 12?** К прежнему Task API добавились User,
   credentials, tokens, middleware, request identity и ownership.
2. **Что осталось?** UseCase, Repository interfaces, memory, DTO/Mapper,
   error mapping, DI, шесть Task endpoints и PATCH/business tests.
3. **Authentication и authorization — одно?** Нет: первая устанавливает
   identity, вторая проверяет право на конкретную Task.
4. **Где Verify?** В authentication middleware через TokenVerifier interface.
5. **Где OwnerID == userID?** В TaskUseCase, после загрузки business data.
6. **Почему middleware не проверяет ownership?** Оно работает до Task Handler
   и не должно принимать решения о предметных данных.
7. **Зачем request.Context?** Identity принадлежит запросу, а не общему Handler.
8. **Почему ключ context приватного типа?** Чтобы не столкнуться с ключом
   другого пакета. Вспомогательные функции задают единый способ чтения.
9. **Должен ли UseCase знать authctx?** Нет: userID передаётся явным аргументом.
10. **Что делают 401 и 403?** 401 — identity не подтверждена, 403 — подтверждена,
    но действие с существующей чужой Task запрещено.
11. **Зачем WWW-Authenticate?** Это Bearer challenge в ответе на отсутствие
    или непригодность credentials для protected endpoint.
12. **Как создаётся OwnerID?** Из проверенного userID; в request DTO его нет.
13. **Почему List не вызывает Get для каждой Task?** Repository выполняет
    owner-scoped чтение, заданное контрактом UseCase, а конкретная проверка
    одной записи нужна Get/Update/Delete.
14. **Что стало с лимитом?** Он остался 20 активных Task, теперь на владельца.
15. **Почему PATCH не превращён в PUT?** Security не меняет смысл обновления;
    nil и false всё ещё разные входы.
16. **Как тестировать ownership без JWT?** Передать userID в UseCase и fake
    Repository с Task другого владельца; запрещённая запись не вызывается.
17. **Как тестировать Handler без bcrypt?** Fake UseCase и fake verifier
    устанавливают identity; настоящий router сохраняет PathValue.
18. **Зачем Hash и Compare в одном интерфейсе?** AuthUseCase задаёт нужное
    поведение и не зависит от конкретной password library.
19. **Почему нельзя сравнить два bcrypt hashes?** У них независимая соль.
    Compare проверяет plaintext против сохранённого hash.
20. **Почему пароль не TrimSpace?** Пробелы — часть credential; регистрация
    и login должны использовать одну и ту же последовательность байтов.
21. **Почему одинаковая ошибка для missing email и bad password?** Внешний
    ответ не должен объяснять, какая часть credentials совпала.
22. **JWT зашифрован?** Этот JWT подписан, payload читается; пароля/hash там нет.
23. **Что проверяет Verify кроме подписи?** Разрешённый алгоритм, обязательный
    exp, срок и положительный user ID.
24. **Почему TokenIssuer отдельно от TokenVerifier?** UseCase выдаёт token,
    middleware проверяет; каждый потребитель зависит от нужного поведения.
25. **Почему UserResponse не равен User?** В Model есть PasswordHash, во
    внешнем ответе только id/email.
26. **Какая новая роль у main?** Прежний composition root выбирает ещё
    memory user storage, hasher, token manager и auth middleware.
27. **Что теряется после рестарта?** Users и Tasks. Новый учебный прогон
    требует нового ключа и повторной регистрации/login.
28. **Можно ли немедленно отозвать JWT в этом примере?** Нет, отдельного
    server-side session/revocation state здесь нет; действует ограниченный TTL.
29. **Как следующая persistence implementation повлияет на Handler?** Она
    должна реализовать прежние Repository contracts; Handler остаётся прежним.
30. **Доказывают ли имена папок архитектуру?** Нет: проверьте imports,
    явные зависимости и то, какие решения принимаются на каждой границе.
