#!/bin/bash

# Установка зависимостей Go
echo "Устанавливаем зависимости Go..."
go mod tidy

# Настройка базы данных
echo "Создаём базу данных и таблицу..."
psql -U validator -d project-sem-1 -c "CREATE TABLE IF NOT EXISTS prices (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    price NUMERIC NOT NULL,
    create_date DATE NOT NULL
);"

echo "Подготовка завершена."
