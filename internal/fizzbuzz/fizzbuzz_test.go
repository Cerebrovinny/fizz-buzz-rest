package fizzbuzz

import (
	"reflect"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		int1  int
		int2  int
		limit int
		str1  string
		str2  string
		want  []string
		check func(t *testing.T, got []string)
	}{
		{
			name:  "classic fizzbuzz",
			int1:  3,
			int2:  5,
			limit: 15,
			str1:  "fizz",
			str2:  "buzz",
			want: []string{
				"1", "2", "fizz", "4", "buzz",
				"fizz", "7", "8", "fizz", "buzz",
				"11", "fizz", "13", "14", "fizzbuzz",
			},
		},
		{
			name:  "custom divisors and strings",
			int1:  2,
			int2:  3,
			limit: 10,
			str1:  "foo",
			str2:  "bar",
			want: []string{
				"1", "foo", "bar", "foo", "5",
				"foobar", "7", "foo", "bar", "foo",
			},
		},
		{
			name:  "zero limit",
			int1:  3,
			int2:  5,
			limit: 0,
			str1:  "fizz",
			str2:  "buzz",
			want:  []string{},
		},
		{
			name:  "negative limit",
			int1:  3,
			int2:  5,
			limit: -5,
			str1:  "fizz",
			str2:  "buzz",
			want:  []string{},
		},
		{
			name:  "limit one",
			int1:  3,
			int2:  5,
			limit: 1,
			str1:  "fizz",
			str2:  "buzz",
			want:  []string{"1"},
		},
		{
			name:  "equal divisors",
			int1:  3,
			int2:  3,
			limit: 10,
			str1:  "foo",
			str2:  "bar",
			want: []string{
				"1", "2", "foobar", "4", "5",
				"foobar", "7", "8", "foobar", "10",
			},
		},
		{
			name:  "int1 divides all",
			int1:  1,
			int2:  5,
			limit: 6,
			str1:  "all",
			str2:  "five",
			want: []string{
				"all", "all", "all", "all", "allfive", "all",
			},
		},
		{
			name:  "large limit",
			int1:  7,
			int2:  11,
			limit: 100,
			str1:  "seven",
			str2:  "eleven",
			check: func(t *testing.T, got []string) {
				t.Helper()

				if len(got) != 100 {
					t.Fatalf("expected result length %d, got %d", 100, len(got))
				}

				if got[6] != "seven" {
					t.Errorf("expected position 7 to be %q, got %q", "seven", got[6])
				}
				if got[10] != "eleven" {
					t.Errorf("expected position 11 to be %q, got %q", "eleven", got[10])
				}
				if got[76] != "seveneleven" {
					t.Errorf("expected position 77 to be %q, got %q", "seveneleven", got[76])
				}

				var sevenCount, elevenCount, bothCount int
				for _, v := range got {
					switch v {
					case "seveneleven":
						bothCount++
					case "seven":
						sevenCount++
					case "eleven":
						elevenCount++
					}
				}

				if bothCount != 1 {
					t.Errorf("expected %d occurrences of %q, got %d", 1, "seveneleven", bothCount)
				}
				if sevenCount != 13 {
					t.Errorf("expected %d occurrences of %q, got %d", 13, "seven", sevenCount)
				}
				if elevenCount != 8 {
					t.Errorf("expected %d occurrences of %q, got %d", 8, "eleven", elevenCount)
				}
			},
		},
		{
			name:  "empty replacement strings",
			int1:  3,
			int2:  5,
			limit: 15,
			str1:  "",
			str2:  "",
			want: []string{
				"1", "2", "", "4", "",
				"", "7", "8", "", "",
				"11", "", "13", "14", "",
			},
		},
		{
			name:  "zero divisors ignored",
			int1:  0,
			int2:  5,
			limit: 6,
			str1:  "nope",
			str2:  "buzz",
			want: []string{
				"1", "2", "3", "4", "buzz", "6",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Generate(tc.int1, tc.int2, tc.limit, tc.str1, tc.str2)
			if tc.want != nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Generate(%d, %d, %d, %q, %q) = %v, want %v",
					tc.int1, tc.int2, tc.limit, tc.str1, tc.str2, got, tc.want)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}
