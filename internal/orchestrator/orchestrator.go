package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID           int     `json:"id"`
	ExpressionID int     `json:"expression_id"`
	Arg1         string  `json:"arg1"`
	Arg2         string  `json:"arg2"`
	Operation    string  `json:"operation"`
	Status       string  `json:"status"`
	Result       float64 `json:"result"`
}

type Expression struct {
	ID     int     `json:"id"`
	Status string  `json:"status"`
	Answer int     `json:"tasks"`
	Result float64 `json:"result"`
}

type Store struct {
	mu          sync.Mutex
	Expressions []Expression
	Tasks       []Task
}

var store = Store{
	Expressions: []Expression{},
	Tasks:       []Task{},
}

type Config struct {
	Addr string // Порт, на котором будет запущен сервер
}

// Функция для создания конфигурации из переменных окружения
func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("ORCHESTRATOR_PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

// Структура приложения, содержащая конфигурацию
type Application struct {
	config *Config
}

// Функция для создания нового экземпляра приложения
func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

// Метод для запуска HTTP-сервера
func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", AddExpressions)
	http.HandleFunc("/api/v1/expressions", GetExpressions)
	http.HandleFunc("/api/v1/expressions/", GetExpressionByID)
	http.HandleFunc("/internal/task", TaskHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}

func sendError(w http.ResponseWriter, code int) {
	var errorMessage string
	switch code {
	case http.StatusNotFound: // 404
		errorMessage = "Not Found"
	case http.StatusUnprocessableEntity: // 422
		errorMessage = "Unprocessable Entity"
	case http.StatusInternalServerError: // 500
		errorMessage = "Internal Server Error"
	default:
		errorMessage = "Unknown error"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}

func AddExpressions(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Expression string `json:"expression"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sendError(w, 422)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	id := generateId("expressions")
	result, err := Calc(input.Expression, id)
	if err != nil {
		sendError(w, 422)
		return
	}
	if strings.HasPrefix(result, "id") {
		ans, err := strconv.Atoi(strings.TrimPrefix(result, "id"))
		if err != nil {
			sendError(w, 500)
			return
		}
		store.Expressions = append(store.Expressions, Expression{ID: id, Status: "waiting", Answer: ans})
	} else {
		result, err := strconv.ParseFloat(result, 64)
		if err != nil {
			sendError(w, 500)
			return
		}
		store.Expressions = append(store.Expressions, Expression{ID: id, Status: "complete", Answer: 0, Result: result})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

func GetExpressions(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	defer store.mu.Unlock()

	response := struct {
		Expressions []struct {
			ID     int     `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"result"`
		} `json:"expressions"`
	}{}

	for _, expr := range store.Expressions {
		response.Expressions = append(response.Expressions, struct {
			ID     int     `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"result"`
		}{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/v1/expressions/"):]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		sendError(w, 404)
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	for _, expr := range store.Expressions {
		if expr.ID == id {
			response := map[string]interface{}{
				"expression": map[string]interface{}{
					"id":     expr.ID,
					"status": expr.Status,
					"result": expr.Result,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	sendError(w, 404)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	defer store.mu.Unlock()

	for i, task := range store.Tasks {
		if task.Status == "waiting" {
			p1, arg1, err := getResult(task.Arg1)
			if err != nil {
				continue
			}
			p2, arg2, err := getResult(task.Arg2)
			if err != nil {
				continue
			}
			response := map[string]interface{}{
				"task": map[string]interface{}{
					"id":             task.ID,
					"arg1":           arg1,
					"arg2":           arg2,
					"operation":      task.Operation,
					"operation_time": getOperationTime(task.Operation),
				},
			}
			store.Tasks[i].Status = "calculating"
			if p1 >= 0 {
				store.Tasks = append(store.Tasks[:p1], store.Tasks[p1+1:]...)
			}
			if p2 >= 0 {
				store.Tasks = append(store.Tasks[:p2], store.Tasks[p2+1:]...)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
	}
	sendError(w, 404)
}

func getOperationTime(operation string) int {
	switch operation {
	case "+":
		return getEnvAsInt("TIME_ADDITION_MS")
	case "-":
		return getEnvAsInt("TIME_SUBTRACTION_MS")
	case "*":
		return getEnvAsInt("TIME_MULTIPLICATIONS_MS")
	case "/":
		return getEnvAsInt("TIME_DIVISIONS_MS")
	default:
		return 0
	}
}

func getResult(input string) (int, float64, error) {
	if strings.HasPrefix(input, "id") {
		id, err := strconv.Atoi(strings.TrimPrefix(input, "id"))
		if err != nil || id < 0 || id >= len(store.Tasks) {
			return -1, 0, fmt.Errorf("invalid task id")
		}

		for i, task := range store.Tasks {
			if task.ID == id {
				if task.Status == "complete" {
					return i, task.Result, nil
				}
			}
		}
		return -1, 0, fmt.Errorf("no result")
	}

	result, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return -1, 0, fmt.Errorf("invalid input")
	}
	return -1, result, nil
}

func getEnvAsInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		return 0
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return intValue
}

func PostResult(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID     int     `json:"id"`
		Result float64 `json:"result"`
		Error  string  `json:"error,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sendError(w, 422)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for i, task := range store.Tasks {
		if task.ID == input.ID {
			if input.Error != "" {
				setError(store.Tasks[i].ExpressionID, input.Error)
			} else {
				store.Tasks[i].Result = input.Result
				store.Tasks[i].Status = "complete"
				for j, expression := range store.Expressions {
					if expression.Answer == input.ID {
						store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
						store.Expressions[j].Answer = -1
						store.Expressions[j].Result = input.Result
						store.Expressions[j].Status = "complete"
					}
				}
			}
			return
		}
	}

	sendError(w, 404)
}

func setError(id int, error_text string) {
	// Удаляем задачи с указанным id
	tasks_len := len(store.Tasks)
	for i := 0; i < tasks_len; i++ {
		if store.Tasks[i].ExpressionID == id {
			store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
			tasks_len--
			i--
		}
	}

	// Обновляем статус выражения
	for i := range store.Expressions {
		if store.Expressions[i].ID == id {
			store.Expressions[i].Status = "error: " + error_text
			break
		}
	}
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetTask(w, r)
	case http.MethodPost:
		PostResult(w, r)
	default:
		sendError(w, 500)
	}
}
