🚀 Go Data Price API

REST API сервис для загрузки и выгрузки данных о ценах, написанный на Go с хранением в PostgreSQL.

✅ В ходе проекта было сделано:

- Реализован Go-веб-сервис с обработкой CSV-данных из zip-архива.

- Данные сохраняются в PostgreSQL.

- Подготовлены bash-скрипты для развертывания, настройки БД и тестирования.

- Внедрен CI-пайплайн GitHub Actions для проверки уровней задания.

🎓 Требования

- Go ≥ 1.20

- PostgreSQL ≥ 13

- Linux / macOS / Windows (WSL)

🔧 Установка и запуск

1. Склонируйте репозиторий:

git clone https://github.com/treasuremeasure/itmo-devops-sem1-project-template.git
cd itmo-devops-sem1-project-template

2. Создайте базу данных:

psql -U validator -d postgres -c "CREATE DATABASE project-sem-1;"

3. Подготовьте среду:

chmod +x scripts/*.sh
./scripts/prepare.sh

4. Запустите сервер:

./scripts/run.sh

API будет доступен по http://localhost:8080

🦜 Автотесты

GitHub Actions (".github/workflows/go_check.yml"):

- Поднимает PostgreSQL

- Сборка Go приложения

- Запуск prepare/run/tests

Проверяет 3 уровня:

1: POST, GET /api/v0/prices

2: ZIP+TAR + PostgreSQL SUM()

3: Обработка ошибочных записей, фильтры

Запуск локально:

./scripts/tests.sh 1   # basic
./scripts/tests.sh 2   # advanced
./scripts/tests.sh 3   # full

📂 Структура

.
├── sample_data/         # Тестовые CSV/архивы
├── scripts/             # Bash-скрипты (prepare, run, tests)
├── main.go              # API-сервис
├── go.mod / go.sum      # Зависимости Go
└── .github/workflows/   # CI/CD pipeline

🔗 Эндпоинты

POST /api/v0/prices

- Загружает zip-файл c CSV.

- Валидирует, сохраняет в PostgreSQL.

GET /api/v0/prices

- Выгружает data.zip с CSV из базы.

- Поддерживаются фильтры: start, end, min, max

Автор: Ордухани Риза
Контакты: orduhaniriza@gmail.com