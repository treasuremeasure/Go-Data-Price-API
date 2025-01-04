package main


import (
	_ "github.com/lib/pq"
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"log"
	"github.com/gorilla/mux"
	"bufio"
	"encoding/json"
	"strings"
)


var db *sql.DB

func main() {
	// Подключение к PostgreSQL

	dbHost := getEnv("POSTGRES_HOST", "localhost")
	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "validator")
	dbPassword := getEnv("POSTGRES_PASSWORD", "val1dat0r")
	dbName := getEnv("POSTGRES_DB", "project-sem-1")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Подключаемся к базе данных: %s", connStr)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Проверка соединения
	err = db.Ping()
	if err != nil {
		log.Fatalf("База данных недоступна: %v", err)
	}

	fmt.Println("Соединение с базой данных успешно установлено!")

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/api/v0/prices", handlePostPrices).Methods("POST")
	router.HandleFunc("/api/v0/prices", handleGetPrices).Methods("GET")

	// Запуск сервера
	fmt.Println("Сервер запущен на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Функция для получения значения переменной окружения с возможностью задания значения по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func handlePostPrices(w http.ResponseWriter, r *http.Request) {
    // Проверяем, что запрос содержит файл
    err := r.ParseMultipartForm(10 << 20) // Максимальный размер: 10 MB
    if err != nil {
        http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
        return
    }

    // Получаем файл из формы
    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Ошибка загрузки файла", http.StatusBadRequest)
        return
    }
    defer file.Close()

    fmt.Printf("Загружен файл: %s\n", handler.Filename)

    // Сохраняем файл временно на диск
    tempFile, err := os.CreateTemp("", "upload-*.zip")
    if err != nil {
        http.Error(w, "Не удалось сохранить файл", http.StatusInternalServerError)
        return
    }
    defer os.Remove(tempFile.Name()) // Удаляем временный файл после обработки

	log.Printf("Создан временный файл: %s", tempFile.Name())

    // Копируем содержимое загруженного файла во временный файл
    _, err = io.Copy(tempFile, file)
    if err != nil {
        http.Error(w, "Ошибка сохранения файла", http.StatusInternalServerError)
        log.Printf("Ошибка сохранения файла: %v", err)
        return
    }

    fmt.Printf("Временный файл: %s\n", tempFile.Name())

    // Разархивируем файл
    zipReader, err := zip.OpenReader(tempFile.Name())
    if err != nil {
        http.Error(w, "Ошибка разархивации файла", http.StatusBadRequest)
        log.Printf("Ошибка разархивации файла: %v", err)
        return
    }
    defer zipReader.Close()

	var csvFile string
	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".csv") {
			csvFile = f.Name
			break
		}
	}

	if csvFile == "" {
		http.Error(w, "CSV-файл не найден в архиве", http.StatusBadRequest)
		return
	}

	// Читаем содержимое CSV-файла
	var totalItems int
	var totalPrice float64
	categories := make(map[string]struct{})

	for _, f := range zipReader.File {
		if f.Name == csvFile {
			rc, err := f.Open()
			if err != nil {
				http.Error(w, "Ошибка открытия CSV-файла", http.StatusInternalServerError)
				return
			}
			defer rc.Close()

			reader := csv.NewReader(bufio.NewReader(rc))
			reader.Read() // Пропускаем заголовок
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Ошибка чтения CSV-файла", http.StatusInternalServerError)
					return
				}

				name := record[1]
				category := record[2]
				price, _ := strconv.ParseFloat(record[3], 64)
				createDate := record[4]

				// Сохраняем в базу данных
				_, err = db.Exec(
					"INSERT INTO prices (name, category, price, create_date) VALUES ($1, $2, $3, $4)",
					name, category, price, createDate,
				)
				if err != nil {
					http.Error(w, "Ошибка записи в базу данных", http.StatusInternalServerError)
					return
				}

				totalItems++
				totalPrice += price
				categories[category] = struct{}{}
			}
		}
	}

	// Формируем JSON-ответ
	response := map[string]interface{}{
		"total_items":     totalItems,
		"total_categories": len(categories),
		"total_price":      totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}





func handleGetPrices(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, category, price, create_date FROM prices")
	if err != nil {
		http.Error(w, "Ошибка получения данных из базы", http.StatusInternalServerError)
		log.Printf("Ошибка базы данных: %v", err)
		return
	}
	defer rows.Close()

	csvFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		http.Error(w, "Ошибка создания временного файла", http.StatusInternalServerError)
		log.Printf("Ошибка создания файла: %v", err)
		return
	}
	defer os.Remove(csvFile.Name())

	writer := csv.NewWriter(csvFile)
	writer.Write([]string{"id", "name", "category", "price", "create_date"})

	for rows.Next() {
		var id int
		var name, category, createDate string
		var price float64

		err := rows.Scan(&id, &name, &category, &price, &createDate)
		if err != nil {
			http.Error(w, "Ошибка чтения данных из базы", http.StatusInternalServerError)
			log.Printf("Ошибка чтения строки: %v", err)
			return
		}

		writer.Write([]string{
			strconv.Itoa(id), name, category, fmt.Sprintf("%.2f", price), createDate,
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		http.Error(w, "Ошибка записи в CSV", http.StatusInternalServerError)
		log.Printf("Ошибка записи в CSV: %v", err)
		return
	}

	_, err = csvFile.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, "Ошибка сброса указателя CSV-файла", http.StatusInternalServerError)
		log.Printf("Ошибка сброса указателя CSV-файла: %v", err)
		return
	}

	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		http.Error(w, "Ошибка создания ZIP-файла", http.StatusInternalServerError)
		log.Printf("Ошибка создания ZIP-файла: %v", err)
		return
	}
	defer os.Remove(zipFile.Name())

	zipWriter := zip.NewWriter(zipFile)
	fileWriter, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Ошибка добавления файла в ZIP", http.StatusInternalServerError)
		log.Printf("Ошибка записи в ZIP: %v", err)
		return
	}

	_, err = io.Copy(fileWriter, csvFile)
	if err != nil {
		http.Error(w, "Ошибка копирования данных в ZIP", http.StatusInternalServerError)
		log.Printf("Ошибка копирования данных в ZIP: %v", err)
		return
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Ошибка закрытия ZIP-файла", http.StatusInternalServerError)
		log.Printf("Ошибка закрытия ZIP-файла: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
	http.ServeFile(w, r, zipFile.Name())
}



