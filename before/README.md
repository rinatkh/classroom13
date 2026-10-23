# Starter до занятия 13

Это архитектурно разделённый Notes API без пользователей и authentication.
Он показывает точку старта live coding: UseCase и Repository уже существуют,
но любой клиент может читать и изменять любые заметки.

```bash
go run ./cmd/api
curl -i http://localhost:8080/notes
```

Во время занятия не переносим всё в новый проект. Добавляем User, AuthUseCase,
TokenManager, middleware и ownership в существующие границы.
