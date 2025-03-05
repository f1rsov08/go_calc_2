package calculation

import (
	"testing"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		shouldFail bool
	}{
		{"1+1", 2, false},        // Сложение
		{"3 -4", -1, false},      // Вычитание
		{"2 * -3", -6, false},    // Умножение
		{"5/ 10", 0.5, false},    // Деление
		{"4. + 3", 0, true},      // Точка в конце числа
		{"6 / .2", 0, true},      // Точка в начале числа
		{"(1 + 2)* 3", 9, false}, // Скобки
		{"2+2*2", 6, false},      // Порядок операций
		{"(1 + 2", 0, true},      // Незакрытая скобка
		{"abc", 0, true},         // Лишние символы
		{"", 0, true},            // Пустое выражение
		{"(3 - 1) * (4 / 2)", 4, false},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := Calc(test.expression)
			if test.shouldFail {
				if err == nil {
					t.Errorf("Expected error for expression: %s, but got none", test.expression)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error for expression: %s, but got: %v", test.expression, err)
				} else if result != test.expected {
					t.Errorf("For expression: %s, expected: %f, but got: %f", test.expression, test.expected, result)
				}
			}
		})
	}
}
