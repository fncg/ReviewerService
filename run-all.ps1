# Билд образа
docker-compose build

# Запуск базы данных
docker-compose up -d db

# Ожидание готовности базы
Start-Sleep -Seconds 5

# Запуск unit тестов (все тесты, кроме интеграционного)
docker-compose run --rm app go test -v ./internal/github ./internal/telegram ./internal/storage -count=1
docker-compose run --rm app go test -v ./internal/http -run "Test(New|Health|Webhook)" -count=1

# Запуск интеграционных тестов
docker-compose run --rm app go test -v ./internal/http -run TestUserStory_PullRequestAutoAssignmentIntegration

# Запуск приложения
docker-compose up app
