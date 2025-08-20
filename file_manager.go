package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// FileManager - менеджер для работы с файлами
type FileManager struct {
	dataDir string
	logger  *Logger
}

func NewFileManager(dataDir string) *FileManager {
	// Создаем директорию если не существует
	os.MkdirAll(dataDir, 0755)

	return &FileManager{
		dataDir: dataDir,
		logger:  NewLogger("FileManager"),
	}
}

// DownloadFile загружает файл по URL
func (fm *FileManager) DownloadFile(url, filename string) (string, error) {
	filepath := fmt.Sprintf("%s/%s.mp4", fm.dataDir, sanitizeFilename(filename))

	fm.logger.Info("Загрузка файла: %s", filename)

	resp, err := http.Get(url)
	if err != nil {
		fm.logger.Error("Ошибка загрузки: %v", err)
		return "", fmt.Errorf("ошибка загрузки: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fm.logger.Error("Неверный статус: %s", resp.Status)
		return "", fmt.Errorf("неверный статус: %s", resp.Status)
	}

	file, err := os.Create(filepath)
	if err != nil {
		fm.logger.Error("Ошибка создания файла: %v", err)
		return "", fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		fm.logger.Error("Ошибка записи файла: %v", err)
		return "", fmt.Errorf("ошибка записи файла: %w", err)
	}

	fm.logger.Success("Файл загружен: %s (%.2f MB)", filename, float64(size)/1024/1024)
	return filepath, nil
}

// ReadFile читает файл
func (fm *FileManager) ReadFile(filepath string) ([]byte, error) {
	fm.logger.Info("Чтение файла: %s", filepath)

	data, err := os.ReadFile(filepath)
	if err != nil {
		fm.logger.Error("Ошибка чтения файла: %v", err)
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	fm.logger.Success("Файл прочитан: %.2f MB", float64(len(data))/1024/1024)
	return data, nil
}

// DeleteFile удаляет файл
func (fm *FileManager) DeleteFile(filepath string) error {
	fm.logger.Info("Удаление файла: %s", filepath)

	if err := os.Remove(filepath); err != nil {
		fm.logger.Warning("Не удалось удалить файл: %v", err)
		return err
	}

	fm.logger.Success("Файл удален")
	return nil
}

// sanitizeFilename очищает имя файла от недопустимых символов
func sanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}
