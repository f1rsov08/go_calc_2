package orchestrator

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func generateId(entityType string) int {
	var ids []int
	switch entityType {
	case "expressions":
		for _, expr := range store.Expressions {
			ids = append(ids, expr.ID)
		}
	case "tasks":
		for _, task := range store.Tasks {
			ids = append(ids, task.ID)
		}
	default:
		return 0
	}

	for i := 0; ; i++ {
		if !contains(ids, i) {
			return i
		}
	}
}

func contains(ids []int, id int) bool {
	for _, existingId := range ids {
		if existingId == id {
			return true
		}
	}
	return false
}

// indexOf ищет индекс первого вхождения символов из values в срезе строк slice
func indexOf(slice []string, values string) int {
	for i, v := range slice {
		if strings.Contains(values, v) {
			return i
		}
	}
	return -1 // Возвращает -1, если символы не найдены.
}

// isValidInNumber проверяет, является ли символ допустимым для числа num
func isValidInNumber(num string, c rune) bool {
	validChars := "0123456789" // Допустимые символы для чисел.
	if c == 'i' {
		return len(num) == 0
	}
	if c == 'd' {
		return num[0] == 'i' && len(num) == 1
	}
	if c == '.' {
		// Точка может быть в числе только один раз и не в начале
		return !strings.Contains(num, ".") && len(num) != 0
	}
	if c == '-' {
		// Минус может быть только в начале
		return len(num) == 0
	}
	if c == '+' {
		// Плюс может быть только в начале
		return len(num) == 0
	}
	// Проверка на допустимые цифры.
	return strings.Contains(validChars, string(c))
}

// isValidOperation проверяет, является ли символ c допустимой операцией
func isValidOperation(c rune) bool {
	validChars := "+-*/" // Допустимые операции
	return strings.Contains(validChars, string(c))
}

// Calc выполняет вычисление математического выражения, переданного в виде строки
func Calc(expression string, id int) (string, error) {
	var err error
	// Удаляем пробелы из выражения
	expression = strings.ReplaceAll(expression, " ", "")

	// Решаем все, что в скобках
	expression, err = evaluate(expression, id)
	if err != nil {
		return "", err
	}

	// Инициализация среза для хранения чисел и операций
	s := []string{""}

	// Проходим по каждому символу в выражении
	for _, i := range expression {
		if isValidInNumber(s[len(s)-1], i) {
			// Если символ допустим для числа, добавляем его к текущему числу
			s[len(s)-1] += string(i)
		} else if isValidOperation(i) {
			// Если символ является операцией, проверяем заканчивается ли число на точку
			if s[len(s)-1][len(s[len(s)-1])-1] == '.' {
				return "", errors.New("Expression is not valid")
			}
			// Если нет, добавляем символ операции в срез и инициализируем новое число
			s = append(s, string(i))
			s = append(s, "")
		} else {
			return "", errors.New("Expression is not valid") // Ошибка при недопустимом символе
		}
	}

	var ind, new_id int

	// Выполняем вычисления до тех пор, пока не останется одно значение
	for len(s) != 1 {
		new_id = generateId("tasks")
		if ind = indexOf(s, "*/"); ind != -1 {
			store.Tasks = append(store.Tasks, Task{ID: new_id, ExpressionID: id, Arg1: s[ind-1], Arg2: s[ind+1], Operation: s[ind], Status: "waiting", Result: 0.0})
		} else if ind = indexOf(s, "+-"); ind != -1 {
			store.Tasks = append(store.Tasks, Task{ID: new_id, ExpressionID: id, Arg1: s[ind-1], Arg2: s[ind+1], Operation: s[ind], Status: "waiting", Result: 0.0})
		}
		// Обновляем срез и убираем использованные элементы
		s[ind+1] = "id" + strconv.Itoa(new_id)
		s = append(s[:ind-1], s[ind+1:]...)
	}
	// Возвращаем результат вычисления как float64
	return s[0], err
}

// evaluate решает выражения в скобках
func evaluate(expression string, id int) (string, error) {
	re := regexp.MustCompile("[+-]{3,}")
	if re.MatchString(expression) {
		return "", errors.New("Expression is not valid")
	}
	s := []string{""} // Инициализация среза для хранения частей выражения
	var n int         // Счетчик открывающих скобок
	var v string
	var err error

	// Проходим по каждому символу в выражении
	for _, i := range expression {
		if i == '(' {
			n++ // Увеличиваем счетчик открывающих скобок
			if n == 1 {
				s = append(s, "") // Добавляем новый элемент в срез при первой открывающей скобке
			} else {
				s[len(s)-1] += string(i) // Добавляем символ к текущему элементу
			}
		} else if i == ')' {
			n-- // Уменьшаем счетчик закрывающих скобок
			if n == 0 {
				v, err = Calc(s[len(s)-1], id) // Вычисляем выражение внутри скобок
				if err != nil {
					return "", err
				}
				s[len(s)-1] = v   // Заменяем выражение на его результат
				s = append(s, "") // Добавляем новый элемент для следующей части выражения
			} else if n < 0 {
				return "", errors.New("Expression is not valid") // Ошибка при наличии лишней закрывающей скобки
			} else {
				s[len(s)-1] += string(i) // Добавляем символ к текущему элементу
			}
		} else {
			s[len(s)-1] += string(i) // Добавляем обычный символ к текущему элементу
		}
	}
	if n == 0 {
		evaluated := strings.Join(s, "")

		// Минус на минус дает плюс
		evaluated = strings.ReplaceAll(evaluated, "--", "+")

		// Другие ситуации
		evaluated = strings.ReplaceAll(evaluated, "+-", "-")
		evaluated = strings.ReplaceAll(evaluated, "-+", "-")

		// Создаем регулярное выражение для поиска повторяющихся плюсов
		re := regexp.MustCompile("\\++")

		// Заменяем все повторяющиеся плюсы на один "+"
		evaluated = re.ReplaceAllString(evaluated, "+")

		re = regexp.MustCompile("[\\+\\-*/]{2,}")
		if re.MatchString(evaluated) {
			return "", errors.New("Expression is not valid")
		}

		return evaluated, nil // Возвращаем объединенное выражение без скобок
	} else {
		return "", errors.New("Expression is not valid") // Ошибка при наличии незакрытых скобок
	}
}
