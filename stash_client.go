package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"time"
)

// StashClient клиент для работы со StashApp API
type StashClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *Logger
}

func NewStashClient(baseURL, apiKey string) *StashClient {
	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
	}

	return &StashClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
		logger: NewLogger("StashClient"),
	}
}

func (s *StashClient) graphQLRequest(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
			s.logger.Warning("Попытка подключения #%d к StashApp...", attempt+1)
		}

		req, err := http.NewRequest("POST", s.baseURL+"/graphql", bytes.NewBuffer(jsonBody))
		if err != nil {
			lastErr = fmt.Errorf("ошибка создания запроса: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if s.apiKey != "" {
			req.Header.Set("ApiKey", s.apiKey)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("ошибка выполнения запроса (попытка %d): %v", attempt+1, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("ошибка чтения ответа: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("сервер вернул статус %d: %s", resp.StatusCode, string(body))
			continue
		}

		var result GraphQLResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			lastErr = fmt.Errorf("ошибка парсинга ответа: %v", err)
			continue
		}

		if len(result.Errors) > 0 {
			lastErr = fmt.Errorf("GraphQL ошибка: %s", result.Errors[0].Message)
			continue
		}

		return &result, nil
	}

	return nil, fmt.Errorf("не удалось подключиться после 3 попыток: %v", lastErr)
}

// cryptoIntn возвращает равномерное случайное число в [0, n) без модульного смещения.
func cryptoIntn(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}
	max := big.NewInt(int64(n))
	v, err := crand.Int(crand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

func (s *StashClient) GetRandomScene() (*Scene, error) {
	s.logger.Info("Получение случайной сцены")

	countQuery := `
		query {
			findScenes {
				count
			}
		}`

	countResp, err := s.graphQLRequest(countQuery, nil)
	if err != nil {
		return nil, err
	}

	count := countResp.Data.FindScenes.Count
	if count == 0 {
		return nil, fmt.Errorf("нет доступных видео")
	}

	// Пытаемся получить индекс через crypto/rand (качественная случайность)
	randomIndex, err := cryptoIntn(count)
	if err != nil {
		// Фоллбек на math/rand — быстрый и потокобезопасный глобальный генератор
		randomIndex = rand.Intn(count)
	}

	query := `
		query FindScenes($filter: FindFilterType) {
			findScenes(filter: $filter) {
				scenes {
					id
					title
					performers{
						id
						name
					}
					paths {
						screenshot
						stream
						preview
						sprite
					}
					studio {
						id
						name
					}
				}
			}
		}`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"page":     randomIndex/1 + 1,
			"per_page": 1,
		},
	}

	resp, err := s.graphQLRequest(query, variables)
	if err != nil {
		return nil, err
	}

	if len(resp.Data.FindScenes.Scenes) == 0 {
		return nil, fmt.Errorf("видео не найдено")
	}

	s.logger.Success("Найдена сцена: %s", resp.Data.FindScenes.Scenes[0].Title)
	return &resp.Data.FindScenes.Scenes[0], nil
}

// TestConnection проверяет подключение к StashApp
func (s *StashClient) TestConnection() error {
	testQuery := `query { systemStatus { databaseSchema }}`
	_, err := s.graphQLRequest(testQuery, nil)
	return err
}
