package alphabet

import "fmt"

func GetLetter(position int, uppercase bool) (rune, error) {
	if position < 1 || position > 26 {
		return 0, fmt.Errorf("position %d out of range", position)
	}
	if uppercase {
		return 'A' + rune(position-1), nil
	}
	return 'a' + rune(position-1), nil
}
