package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ComputingPower int
	WaitTime       int
}

// Функция для создания конфигурации из переменных окружения
func ConfigFromEnv() *Config {
	config := new(Config)

	strComputingPower := os.Getenv("COMPUTING_POWER")
	computingPower, err := strconv.Atoi(strComputingPower)
	if err != nil {
		computingPower = 1
	}
	config.ComputingPower = computingPower

	strWaitTime := os.Getenv("WAIT_TIME")
	waitTime, err := strconv.Atoi(strWaitTime)
	if err != nil {
		waitTime = 100
	}
	config.WaitTime = waitTime
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
func (a *Application) Run() {
	for i := 0; i < a.config.ComputingPower; i++ {
		go func() {
			for {
				task, err := fetchTask()
				if err == nil {
					compute(task)
				}
				time.Sleep(time.Duration(a.config.WaitTime) * time.Millisecond)
			}
		}()
	}
}

type TaskResponse struct {
	Task Task `json:"task"`
}

type Task struct {
	ID            int     `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

func fetchTask() (*Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("no task available")
	}

	var taskResponse TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResponse); err != nil {
		return nil, err
	}
	return &taskResponse.Task, nil
}

func sendResult(taskID int, result float64) error {
	data := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func sendError(taskID int) error {
	data := map[string]interface{}{
		"id":    taskID,
		"error": "division by zero",
	}
	body, _ := json.Marshal(data)

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func compute(task *Task) {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 != 0 {
			result = task.Arg1 / task.Arg2
		} else {
			sendError(task.ID)
			return
		}
	}
	sendResult(task.ID, result)
}
