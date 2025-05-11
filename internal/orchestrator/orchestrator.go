package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"net"

	pb "github.com/f1rsov08/go_calc_2/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	UserID int     `json:"user_id"`
	Status string  `json:"status"`
	Answer int     `json:"tasks"`
	Result float64 `json:"result"`
}

type User struct {
	ID             int
	Login          string
	Password       string
	OriginPassword string
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

type Server struct {
	pb.TaskServiceServer // сервис из сгенерированного пакета
}

func NewServer() *Server {
	return &Server{}
}

// Метод для запуска HTTP-сервера
func (a *Application) RunServer() error {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	err = db.PingContext(context.Background())
	if err != nil {
		panic(err)
	}
	createTables(context.Background(), db)
	db.Close()
	http.HandleFunc("/api/v1/calculate", AddExpressions)
	http.HandleFunc("/api/v1/expressions", GetExpressions)
	http.HandleFunc("/api/v1/expressions/", GetExpressionByID)
	http.HandleFunc("/api/v1/register", Register)
	http.HandleFunc("/api/v1/login", Login)
	go func() {
		if err := http.ListenAndServe(":"+a.config.Addr, nil); err != nil {
			panic(err)
		}
	}()

	grpcServer := grpc.NewServer()
	taskServiceServer := NewServer()
	pb.RegisterTaskServiceServer(grpcServer, taskServiceServer)
	lis, err := net.Listen("tcp", "localhost:50042")
	if err != nil {
		return err
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		login TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions (
  		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
  		status TEXT,
  		answer INTEGER,
  		result REAL,
  		FOREIGN KEY (user_id) REFERENCES users(id)
 	);`

		tasksTable = `
	CREATE TABLE IF NOT EXISTS tasks (
  		id INTEGER PRIMARY KEY AUTOINCREMENT,
  		expression_id INTEGER,
  		arg1 TEXT,
  		arg2 TEXT,
  		operation TEXT,
  		status TEXT,
  		result REAL,
  		FOREIGN KEY (expression_id) REFERENCES expressions(id)
 	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, tasksTable); err != nil {
		return err
	}

	return nil
}

func insertUser(ctx context.Context, db *sql.DB, user User) (int64, error) {
	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	result, err := db.ExecContext(ctx, q, user.Login, user.Password)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func insertExpression(ctx context.Context, db *sql.DB, expression Expression) (int, error) {
	var q = `
	INSERT INTO expressions (user_id, status, answer, result) values ($1, $2, $3, $4)
	`
	result, err := db.ExecContext(ctx, q, expression.UserID, expression.Status, expression.Answer, expression.Result)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func insertTask(ctx context.Context, db *sql.DB, task Task) (int, error) {
	var q = `
	INSERT INTO tasks (expression_id, arg1, arg2, operation, status, result) values ($1, $2, $3, $4, $5, $6)
	`
	result, err := db.ExecContext(ctx, q, task.ExpressionID, task.Arg1, task.Arg2, task.Operation, task.Status, task.Result)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func selectExpressionsByUserID(ctx context.Context, db *sql.DB, userID int) ([]Expression, error) {
	var expressions []Expression
	var q = "SELECT id, user_id, status, answer, result FROM expressions WHERE user_id = ?"

	rows, err := db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := Expression{}
		err := rows.Scan(&e.ID, &e.UserID, &e.Status, &e.Answer, &e.Result)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expressions, nil
}

func selectExpressionsByAnswer(ctx context.Context, db *sql.DB, answer int) ([]Expression, error) {
	var expressions []Expression
	var q = "SELECT id, user_id, status, answer, result FROM expressions WHERE answer = ?"

	rows, err := db.QueryContext(ctx, q, answer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := Expression{}
		err := rows.Scan(&e.ID, &e.UserID, &e.Status, &e.Answer, &e.Result)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expressions, nil
}

func selectExpressionByID(ctx context.Context, db *sql.DB, id int) (Expression, error) {
	e := Expression{}
	var q = "SELECT id, user_id, status, answer, result FROM expressions WHERE id = ?"
	err := db.QueryRowContext(ctx, q, id).Scan(&e.ID, &e.UserID, &e.Status, &e.Answer, &e.Result)
	if err != nil {
		return e, err
	}

	return e, nil
}

func selectTasks(ctx context.Context, db *sql.DB) ([]Task, error) {
	var tasks []Task
	var q = "SELECT id, expression_id, arg1, arg2, operation, status, result FROM tasks"

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		t := Task{}
		err := rows.Scan(&t.ID, &t.ExpressionID, &t.Arg1, &t.Arg2, &t.Operation, &t.Status, &t.Result)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func deleteTask(ctx context.Context, db *sql.DB, taskID int) error {
	q := "DELETE FROM tasks WHERE id = ?"
	_, err := db.ExecContext(ctx, q, taskID)
	if err != nil {
		return err
	}
	return nil
}

func updateTaskField(ctx context.Context, db *sql.DB, id int, field string, value interface{}) error {
	q := fmt.Sprintf("UPDATE tasks SET %s = $1 WHERE id = $2", field)
	_, err := db.ExecContext(ctx, q, value, id)
	if err != nil {
		return err
	}
	return nil
}

func selectTaskByID(ctx context.Context, db *sql.DB, id int) (Task, error) {
	t := Task{}
	var q = "SELECT id, expression_id, arg1, arg2, operation, status, result FROM tasks WHERE id = $1"
	err := db.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.ExpressionID, &t.Arg1, &t.Arg2, &t.Operation, &t.Status, &t.Result)
	if err != nil {
		return t, err
	}
	return t, nil
}

func updateExpressionField(ctx context.Context, db *sql.DB, id int, field string, value interface{}) error {
	q := fmt.Sprintf("UPDATE expressions SET %s = $1 WHERE id = $2", field)
	_, err := db.ExecContext(ctx, q, value, id)
	if err != nil {
		return err
	}
	return nil
}

func deleteTasksByExpressionID(ctx context.Context, db *sql.DB, expressionID int) error {
	q := "DELETE FROM tasks WHERE expression_id = $1"
	_, err := db.ExecContext(ctx, q, expressionID)
	if err != nil {
		return err
	}
	return nil
}

func getUserByLogin(ctx context.Context, db *sql.DB, login string) (User, error) {
	u := User{}
	var q = "SELECT id, login, password FROM users WHERE login = $1"
	err := db.QueryRowContext(ctx, q, login).Scan(&u.ID, &u.Login, &u.Password)
	if err != nil {
		return u, err
	}
	return u, nil
}

func sendError(w http.ResponseWriter, code int) {
	var errorMessage string
	switch code {
	case http.StatusBadRequest: // 400
		errorMessage = "Bad Request"
	case http.StatusUnauthorized: // 401
		errorMessage = "Unauthorized"
	case http.StatusForbidden: // 403
		errorMessage = "Forbidden"
	case http.StatusNotFound: // 404
		errorMessage = "Not Found"
	case http.StatusMethodNotAllowed: // 405
		errorMessage = "Method Not Allowed"
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
	token := r.Header.Get("Authorization")
	user, err := getUserFromToken(token)
	if err != nil {
		sendError(w, 401)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendError(w, 422)
		return
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		sendError(w, 500)
	}
	defer db.Close()

	id, err := insertExpression(context.Background(), db, Expression{UserID: user.ID, Status: "waiting"})
	if err != nil {
		sendError(w, 500)
		return
	}
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
		updateExpressionField(context.Background(), db, id, "answer", ans)
	} else {
		result, err := strconv.ParseFloat(result, 64)
		if err != nil {
			sendError(w, 500)
			return
		}
		updateExpressionField(context.Background(), db, id, "status", "complete")
		updateExpressionField(context.Background(), db, id, "answer", 0)
		updateExpressionField(context.Background(), db, id, "result", result)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

func GetExpressions(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	user, err := getUserFromToken(token)
	if err != nil {
		sendError(w, 403)
		return
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		sendError(w, 500)
	}
	defer db.Close()

	response := struct {
		Expressions []struct {
			ID     int     `json:"id"`
			Status string  `json:"status"`
			Result float64 `json:"result"`
		} `json:"expressions"`
	}{}

	expressions, err := selectExpressionsByUserID(context.Background(), db, user.ID)
	if err != nil {
		sendError(w, 404)
		return
	}
	for _, expr := range expressions {
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

	token := r.Header.Get("Authorization")
	user, err := getUserFromToken(token)
	if err != nil {
		sendError(w, 403)
		return
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		sendError(w, 500)
	}
	defer db.Close()

	expr, err := selectExpressionByID(context.Background(), db, id)
	if err == nil {
		if expr.UserID != user.ID {
			sendError(w, 403)
			return
		}
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

	sendError(w, 404)
}

func (s *Server) GetTask(
	ctx context.Context,
	in *pb.GetTaskRequest,
) (*pb.GetTaskResponse, error) {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	defer db.Close()

	tasks, err := selectTasks(context.Background(), db)
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	for _, task := range tasks {
		if task.Status == "waiting" {
			p1, arg1, err := getResult(task.Arg1)
			if err != nil {
				continue
			}
			p2, arg2, err := getResult(task.Arg2)
			if err != nil {
				continue
			}
			response := &pb.GetTaskResponse{
				Task: &pb.Task{
					Id:            int64(task.ID),
					Arg1:          float32(arg1),
					Arg2:          float32(arg2),
					Operation:     task.Operation,
					OperationTime: int64(getOperationTime(task.Operation)),
				},
			}
			updateTaskField(context.Background(), db, task.ID, "status", "calculating")
			if p1 >= 0 {
				deleteTask(context.Background(), db, p1)
			}
			if p2 >= 0 {
				deleteTask(context.Background(), db, p2)
			}
			return response, nil
		}
	}
	return nil, status.Error(codes.NotFound, "Not Found")
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
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		return -1, 0, err
	}
	defer db.Close()

	if strings.HasPrefix(input, "id") {
		id, err := strconv.Atoi(strings.TrimPrefix(input, "id"))
		if err != nil {
			return -1, 0, err
		}

		task, err := selectTaskByID(context.Background(), db, id)
		if err != nil {
			return -1, 0, err
		}
		if task.Status == "complete" {
			return task.ID, task.Result, nil
		}
	}

	result, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return -1, 0, err
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
func (s *Server) PostResult(
	ctx context.Context,
	in *pb.PostResultRequest,
) (*pb.PostResultResponse, error) {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	defer db.Close()

	task, err := selectTaskByID(context.Background(), db, int(in.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, "Not Found")
	}
	if in.Error != "" {
		setError(task.ExpressionID, in.Error)
	} else {
		updateTaskField(context.Background(), db, task.ID, "result", in.Result)
		updateTaskField(context.Background(), db, task.ID, "status", "complete")
		expressions, err := selectExpressionsByAnswer(context.Background(), db, int(in.Id))
		if err != nil {
			return nil, status.Error(codes.NotFound, "Not Found")
		}
		for _, expression := range expressions {
			deleteTask(context.Background(), db, task.ID)
			updateExpressionField(context.Background(), db, expression.ID, "answer", -1)
			updateExpressionField(context.Background(), db, expression.ID, "result", in.Result)
			updateExpressionField(context.Background(), db, expression.ID, "status", "complete")
		}
	}
	return &pb.PostResultResponse{
		Status: "OK",
	}, nil
}

func setError(id int, error_text string) {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		return
	}
	defer db.Close()

	deleteTasksByExpressionID(context.Background(), db, id)

	updateExpressionField(context.Background(), db, id, "status", "error: "+error_text)
}

func generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

func compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, 405)
		return
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		sendError(w, 500)
	}
	defer db.Close()

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendError(w, 422)
		return
	}

	user.OriginPassword = user.Password
	user.Password, err = generate(user.OriginPassword)
	if err != nil {
		sendError(w, 500)
		return
	}

	if _, err := insertUser(context.Background(), db, user); err != nil {
		sendError(w, 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		sendError(w, 500)
	}
	defer db.Close()

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, 422)
		return
	}

	user, err := getUserByLogin(context.Background(), db, req.Login)
	if err != nil {
		sendError(w, 401)
		return
	}

	if err := compare(user.Password, req.Password); err != nil {
		sendError(w, 401)
		return
	}

	const hmacSampleSecret = "super_secret_signature"
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": user.Login,
		"nbf":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
		"iat":  now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		sendError(w, 500)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func getUserFromToken(tokenString string) (User, error) {
	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		return User{}, err
	}
	defer db.Close()

	if !strings.HasPrefix(tokenString, "Bearer ") {
		return User{}, fmt.Errorf("invalid token")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenFromString, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("super_secret_signature"), nil
	})

	if err != nil {
		return User{}, err
	}

	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok && tokenFromString.Valid {
		login := claims["name"].(string)
		user, err := getUserByLogin(context.Background(), db, login)
		if err != nil {
			return User{}, err
		}
		return user, nil
	}

	return User{}, fmt.Errorf("invalid token claims")
}
