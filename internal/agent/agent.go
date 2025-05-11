package agent

import (
	"context"
	"os"
	"strconv"
	"time"

	pb "github.com/f1rsov08/go_calc_2/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.NewClient("localhost:50042", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := pb.NewTaskServiceClient(conn)

	for i := 0; i < a.config.ComputingPower; i++ {
		go func() {
			for {
				taskResponse, err := client.GetTask(context.Background(), &pb.GetTaskRequest{})
				if err == nil {
					compute(client, taskResponse.Task)
				}
				time.Sleep(time.Duration(a.config.WaitTime) * time.Millisecond)
			}
		}()
	}
}

func sendResult(client pb.TaskServiceClient, taskID int64, result float32) error {
	data := &pb.PostResultRequest{
		Id:     int64(taskID),
		Result: float32(result),
	}
	_, err := client.PostResult(context.Background(), data)
	if err != nil {
		return err
	}
	return nil
}

func sendError(client pb.TaskServiceClient, taskID int64) error {
	data := &pb.PostResultRequest{
		Id:    int64(taskID),
		Error: "division by zero",
	}
	_, err := client.PostResult(context.Background(), data)
	if err != nil {
		return err
	}
	return nil
}

func compute(client pb.TaskServiceClient, task *pb.Task) {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	var result float32
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
			sendError(client, task.Id)
			return
		}
	}
	sendResult(client, task.Id, result)
}
