package fizzbuzz

import "strconv"

// Generate returns a slice containing the FizzBuzz sequence
func Generate(int1, int2, limit int, str1, str2 string) []string {
	if limit <= 0 {
		return []string{}
	}

	result := make([]string, 0, limit)

	for n := 1; n <= limit; n++ {
		divisibleByInt1 := false
		if int1 != 0 {
			divisibleByInt1 = n%int1 == 0
		}
		divisibleByInt2 := false
		if int2 != 0 {
			divisibleByInt2 = n%int2 == 0
		}

		switch {
		case divisibleByInt1 && divisibleByInt2:
			result = append(result, str1+str2)
		case divisibleByInt1:
			result = append(result, str1)
		case divisibleByInt2:
			result = append(result, str2)
		default:
			result = append(result, strconv.Itoa(n))
		}
	}

	return result
}
