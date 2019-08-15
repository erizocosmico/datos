package datos

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

// Datetime format returned by the API.
type Datetime struct {
	time.Time
}

var months = map[string]time.Month{
	"ene": time.January,
	"feb": time.February,
	"mar": time.March,
	"abr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"ago": time.August,
	"sep": time.September,
	"oct": time.October,
	"nov": time.November,
	"dic": time.December,
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Datetime) UnmarshalJSON(b []byte) error {
	r := bufio.NewReader(bytes.NewReader(b))

	var mo string
	var day, y, h, m, s int
	steps := []parseFunc{
		expectChars(`"`),
		skipChars(3),
		expectChars(","),
		skipSpaces,
		readInt(2, &day),
		skipSpaces,
		readChars(3, &mo),
		skipSpaces,
		readInt(4, &y),
		skipSpaces,
		readInt(2, &h),
		expectChars(":"),
		readInt(2, &m),
		expectChars(":"),
		readInt(2, &s),
		skipSpaces,
		expectChars("GMT+0000"),
		expectChars(`"`),
	}

	for _, s := range steps {
		if err := s(r); err != nil {
			return fmt.Errorf("error parsing date %q: %s", string(b), err)
		}
	}

	month, ok := months[mo]
	if !ok {
		return fmt.Errorf("invalid month: %s", mo)
	}

	d.Time = time.Date(y, month, day, h, m, s, 0, time.UTC)
	return nil
}

type parseFunc func(*bufio.Reader) error

func readInt(size int, out *int) parseFunc {
	return func(r *bufio.Reader) error {
		var str string
		if err := readChars(size, &str)(r); err != nil {
			return err
		}

		n, err := strconv.Atoi(str)
		if err != nil {
			return err
		}

		*out = n
		return nil
	}
}

func expectChars(expected string) parseFunc {
	return func(r *bufio.Reader) error {
		n := utf8.RuneCountInString(expected)
		var result string
		if err := readChars(n, &result)(r); err != nil {
			return err
		}

		if result != expected {
			return fmt.Errorf("expecting: %s, got: %s", expected, result)
		}
		return nil
	}
}

func readChars(n int, out *string) parseFunc {
	return func(r *bufio.Reader) error {
		var runes = make([]rune, n)
		for i := 0; i < n; i++ {
			ru, _, err := r.ReadRune()
			if err != nil {
				if err == io.EOF && n-1 == i {
					continue
				}
				return err
			}
			runes[i] = ru
		}
		*out = string(runes)
		return nil
	}
}

func skipChars(n int) parseFunc {
	return func(r *bufio.Reader) error {
		var out string
		return readChars(n, &out)(r)
	}
}

func skipSpaces(r *bufio.Reader) error {
	for {
		ru, _, err := r.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if !unicode.IsSpace(ru) {
			return r.UnreadRune()
		}
	}
}
