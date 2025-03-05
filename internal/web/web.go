package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

type Expression struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}

type ErrorR struct {
	Error string `json:"error"`
}

type IDR struct {
	ID int `json:"id"`
}

type PageData struct {
	Expressions []Expression
	Info        string
}

var pageData PageData

func Main() {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		expression := r.FormValue("expression")
		info := sendCalculationRequest(expression)
		pageData.Info = info
	}

	pageData.Expressions = fetchExpressions()

	tmpl := `
 <!DOCTYPE html>
 <html lang="ru">
 <head>
  <meta charset="UTF-8">
  <title>Math Expression Calculator</title>
 </head>
 <body>
  <h1>Введите математическое выражение</h1>
  <form method="POST">
   <input type="text" name="expression" required>
   <button type="submit">Отправить</button>
   <button type="button" onclick="location.reload();">Перезагрузить</button>
  </form>
  {{if .Info}}<p>{{.Info}}</p>{{end}}
  <h2>Список выражений</h2>
  <table border="1">
   <tr>
    <th>ID</th>
    <th>Статус</th>
    <th>Результат</th>
   </tr>
   {{range .Expressions}}
   <tr>
    <td>{{.ID}}</td>
    <td>{{.Status}}</td>
    <td>{{.Result}}</td>
   </tr>
   {{end}}
  </table>
 </body>
 </html>`
	t, _ := template.New("index").Parse(tmpl)
	t.Execute(w, pageData)
}

func sendCalculationRequest(expression string) string {
	data := map[string]string{"expression": expression}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post("http://localhost:8080/calculate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "Ошибка при отправке запроса: " + err.Error()
	}
	defer resp.Body.Close()

	var idResponse IDR
	var errorResponse ErrorR

	if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&idResponse)
		return "ID: " + strconv.Itoa(idResponse.ID)
	} else {
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return "Ошибка: " + errorResponse.Error
	}
}

func fetchExpressions() []Expression {
	resp, err := http.Get("http://localhost:8080/expressions")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var expressionsResponse struct {
		Expressions []Expression `json:"expressions"`
	}
	json.NewDecoder(resp.Body).Decode(&expressionsResponse)

	return expressionsResponse.Expressions
}
