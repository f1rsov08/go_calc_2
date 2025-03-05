package calculation

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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
	if c == '.' {
		// Точка может быть в числе только один раз и не в начале
		return !strings.Contains(num, ".") && len(num) != 0
	}
	if c == '-' {
		// Минус может быть в числе только один раз и только в начале
		return !strings.Contains(num, "-") && len(num) == 0
	}
	if c == '+' {
		// Плюс может быть в числе только один раз и только в начале
		return !strings.Contains(num, "-") && len(num) == 0
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
func Calc(expression string) (float64, error) {
	var err error
	// Удаляем пробелы из выражения
	expression = strings.ReplaceAll(expression, " ", "")

	// Решаем все, что в скобках
	expression, err = evaluate(expression)
	if err != nil {
		return 0.0, err
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
				return 0, errors.New("Expression is not valid")
			}
			// Если нет, добавляем символ операции в срез и инициализируем новое число
			s = append(s, string(i))
			s = append(s, "")
		} else {
			return 0, errors.New("Expression is not valid") // Ошибка при недопустимом символе
		}
	}

	var ind int
	var n, n1, n2 float64

	// Выполняем вычисления до тех пор, пока не останется одно значение
	for len(s) != 1 {
		if ind = indexOf(s, "*/"); ind != -1 {
			// Сначала обрабатываем умножение и деление
			n1, err = strconv.ParseFloat(s[ind-1], 64)
			if err != nil {
				return 0, err
			}
			n2, err = strconv.ParseFloat(s[ind+1], 64)
			if err != nil {
				return 0, err
			}
			if s[ind] == "*" {
				n = n1 * n2 // Умножение
			} else if s[ind] == "/" {
				if n2 == 0 {
					return 0, errors.New("Internal server error") // Ошибка деления на ноль
				}
				n = n1 / n2 // Деление
			}
		} else if ind = indexOf(s, "+-"); ind != -1 {
			// Обрабатываем сложение и вычитание
			n1, err = strconv.ParseFloat(s[ind-1], 64)
			if err != nil {
				return 0, err
			}
			n2, err = strconv.ParseFloat(s[ind+1], 64)
			if err != nil {
				return 0, err
			}
			if s[ind] == "+" {
				n = n1 + n2 // Сложение
			} else if s[ind] == "-" {
				n = n1 - n2 // Вычитание
			}
		}
		// Обновляем срез и убираем использованные элементы
		s[ind+1] = strconv.FormatFloat(n, 'f', -1, 64)
		s = append(s[:ind-1], s[ind+1:]...)
	}
	// Возвращаем результат вычисления как float64
	n, err = strconv.ParseFloat(s[0], 64)
	return n, err
}

// evaluate решает выражения в скобках
func evaluate(expression string) (string, error) {
	re := regexp.MustCompile("[+-]{3,}")
	if re.MatchString(expression) {
		return "", errors.New("Expression is not valid")
	}
	s := []string{""} // Инициализация среза для хранения частей выражения
	var n int         // Счетчик открывающих скобок
	var v float64
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
				v, err = Calc(s[len(s)-1]) // Вычисляем выражение внутри скобок
				if err != nil {
					return "", err
				} else {
					s[len(s)-1] = fmt.Sprintf("%v", v) // Заменяем выражение на его результат
					s = append(s, "")                  // Добавляем новый элемент для следующей части выражения
				}
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
