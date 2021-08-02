package null

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	invalidYearFrom = 2000
	invalidYearTo   = 2050
)

// vars
var (
	ErrUnknownFormat = errors.New("unknown format of card date")
	ErrInvalidYear   = errors.New("invalid year in card date")
	ErrInvalidMonth  = errors.New("invalid month in card date")
)

var (
	cardDateRFC3339Regex     = regexp.MustCompile("^([0-9]{4}-[0-9]{2}-[0-9]{2}[Tt][0-9]{2}:[0-9]{2}:[0-9]{2}[Zz+-:0-9]{1,6}$)")
	cardDateRFC3339NanoRegex = regexp.MustCompile("^([0-9]{4}-[0-9]{2}-[0-9]{2}[Tt][0-9]{2}:[0-9]{2}:[0-9]{2}.[0-9]{7,9}[Zz+-:0-9]{1,6}$)")
	cardDateRFC1123Regex     = regexp.MustCompile("^([A-Za-z]{3}, [0-9]{2} [A-Za-z]{3} [0-9]{4} [0-9]{2}:[0-9]{2}:[0-9]{2} [A-Za-z]{3,4}$)")
	cardDateRFC1123ZRegex    = regexp.MustCompile("^([A-Za-z]{3}, [0-9]{2} [A-Za-z]{3} [0-9]{4} [0-9]{2}:[0-9]{2}:[0-9]{2} [-+]{1}[0-9]{4}$)")
	cardDateRFC822ZRegex     = regexp.MustCompile("^([0-9]{2} [A-Za-z]{3} [0-9]{2} [0-9]{2}:[0-9]{2} [-+]{1}[0-9]{4}$)")
	cardDateRFC822Regex      = regexp.MustCompile("^([0-9]{2} [A-Za-z]{3} [0-9]{2} [0-9]{2}:[0-9]{2} [A-Za-z]{3,4}$)")
	cardDateRFC850Regex      = regexp.MustCompile("^([A-Za-z]{6,9}, [0-9]{2}-[A-Za-z]{3}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} [A-Z]{3,4}$)")
	cardDateRubyFormatRegex  = regexp.MustCompile("^([A-Za-z]{3} [A-Za-z]{3} [0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2} [+-][0-9]{4} [0-9]{4}$)")
	cardDateUnixFormatRegex  = regexp.MustCompile("^([A-Za-z]{3} [A-Za-z]{3} [0-9_ ]{1,2} [0-9]{2}:[0-9]{2}:[0-9]{2} [A-Za-z]{3,4} [0-9]{4}$)")
	cardDateANSICFormatRegex = regexp.MustCompile("^([A-Za-z]{3} [A-Za-z]{3} [0-9_ ]{1,2} [0-9]{2}:[0-9]{2}:[0-9]{2} [0-9]{4}$)")
)

// CardDate is a nullable time.Time. It supports SQL and JSON serialization.
type CardDate struct {
	Time  time.Time
	Valid bool
}

// NewCardDate creates a new Time.
func NewCardDate(t time.Time, valid bool) CardDate {
	return CardDate{
		Time:  t,
		Valid: valid,
	}
}

// CardDateFrom creates a new Time that will always be valid.
func CardDateFrom(t time.Time) CardDate {
	return NewCardDate(t, true)
}

// CardDateFromString creates a new Time that valid, if format[01/06] is valid.
// Else return error and invalid struct.
func CardDateFromString(s string) (CardDate, error) {
	t, err := ParseExpToTime(s)
	if err != nil {
		return CardDate{}, err
	}
	return CardDate{
		Time:  t,
		Valid: true,
	}, nil
}

// CardDateFromMustString creates a new Time that valid, if format[01/06] is valid.
// Else return error and invalid struct.
// NOTE: Use only for TEST purposes!
func CardDateFromMustString(s string) CardDate {
	t, err := ParseExpToTime(s)
	if err != nil {
		panic(err)
	}
	return CardDate{
		Time:  t,
		Valid: true,
	}
}

// MarshalJSON implements json.Marshaler.
func (t CardDate) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return NullBytes, nil // @TODO it should be an error
	}

	return []byte(`"` + t.Time.Format("01/06") + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *CardDate) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, NullBytes) {
		t.Valid = false
		t.Time = time.Time{}
		return nil
	}

	str := strings.TrimPrefix(string(data), `"`)
	str = strings.TrimSuffix(str, `"`)
	str = strings.Replace(str, `\`, "", -1)
	var err error
	t.Time, err = ParseExpToTime(str)
	if err != nil {
		t.Valid = false
		t.Time = time.Time{}
		return err
	}

	t.Valid = true
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (t CardDate) MarshalText() ([]byte, error) {
	if !t.Valid {
		return NullBytes, nil
	}

	return []byte(t.Time.Format("01/06")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *CardDate) UnmarshalText(text []byte) error {
	if bytes.Equal(text, NullBytes) {
		t.Valid = false
		t.Time = time.Time{}
		return nil
	}

	var err error
	t.Time, err = ParseExpToTime(string(text))
	if err != nil {
		t.Valid = false
		return err
	}
	t.Valid = true
	return nil
}

// SetValid changes this Time's value and sets it to be non-null.
func (t *CardDate) SetValid(v time.Time) {
	t.Time = v
	t.Valid = true
}

// SetValidFromStr ...
func (t *CardDate) SetValidFromStr(v string) {
	parsed, _ := ParseExpToTime(v)
	t.Time = parsed
	t.Valid = true
}

// Scan implements the Scanner interface.
func (t *CardDate) Scan(value interface{}) error {
	var err error
	switch x := value.(type) {
	case time.Time:
		t.Time = x
	case string:
		t.Time, err = ParseExpToTime(x)
	case nil:
		t.Valid = false
		return nil
	default:
		err = fmt.Errorf("null: cannot scan type %T into null.Time: %v", value, value)
	}
	t.Valid = err == nil
	return err
}

// Value implements the driver Valuer interface.
func (t CardDate) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Time, nil
}

// AddDate ...
func (t CardDate) AddDate(years int, months int, days int) CardDate {
	n := t.Time.AddDate(years, months, days)
	t.Time = n
	return t
}

// String ...
func (t CardDate) String() string {
	if t.Valid {
		return t.Time.Format("01/06")
	}
	return "null"
}

// Validate ...
func (t *CardDate) Validate(str string) error {
	_, err := stringToTime(str)
	return err
}

// GenerateCardDate ...
func GenerateCardDate(nextInt func() int64, fieldType string, shouldBeNull bool) CardDate {
	t := CardDate{}
	if shouldBeNull {
		t.Time = time.Time{}
		t.Valid = false
	} else {
		t.Time = time.Date(
			int(2020+nextInt()%60),
			time.Month(1+(nextInt()%12)),
			int(1+(nextInt()%25)),
			0,
			0,
			0,
			0,
			time.UTC,
		)
		t.Valid = true
	}
	return t
}

// ParseExpToTime takes a exp_date in certain format and returns time.time
// Supported formats: ANSIC, UnixDate, RubyDate, RFC822, RFC822Z, RFC850,
// 					  RFC1123, RFC1123Z, RFC3339, RFC3339Nano, MM/YY, MMYY
//					  MM-YY, MM/YYYY, MMYYYY, MM-YYYY
func ParseExpToTime(exp string) (time.Time, error) {
	return stringToTime(exp)
}

func stringToTime(s string) (time.Time, error) {
	switch {
	case cardDateRFC3339Regex.MatchString(s):
		return timeParser(time.RFC3339, s)
	case cardDateRFC3339NanoRegex.MatchString(s):
		return timeParser(time.RFC3339Nano, s)
	case cardDateRFC1123ZRegex.MatchString(s):
		return timeParser(time.RFC1123Z, s)
	case cardDateRFC1123Regex.MatchString(s):
		return timeParser(time.RFC1123, s)
	case cardDateRFC850Regex.MatchString(s):
		return timeParser(time.RFC850, s)
	case cardDateRFC822Regex.MatchString(s):
		return timeParser(time.RFC822, s)
	case cardDateRFC822ZRegex.MatchString(s):
		return timeParser(time.RFC822Z, s)
	case cardDateRubyFormatRegex.MatchString(s):
		return timeParser(time.RubyDate, s)
	case cardDateUnixFormatRegex.MatchString(s):
		return timeParser(time.UnixDate, s)
	case cardDateANSICFormatRegex.MatchString(s):
		return timeParser(time.ANSIC, s)
	case len(s) == 4:
		return timeParser("0106", s)
	case len(s) == 6:
		return timeParser("012006", s)
	case len(s) == 5 && strings.Contains(s, "/"):
		return timeParser("01/06", s)
	case len(s) == 5 && strings.Contains(s, "-"):
		return timeParser("01-06", s)
	case len(s) == 7 && strings.Contains(s, "/"):
		return timeParser("01/2006", s)
	case len(s) == 7 && strings.Contains(s, "-"):
		return timeParser("01-2006", s)
	}

	return time.Time{}, ErrUnknownFormat
}

func timeParser(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil && strings.Contains(err.Error(), "month out of range") {
		return t, ErrInvalidMonth
	}
	if err != nil {
		return time.Time{}, err
	}
	t = t.UTC()
	if t.Year() <= invalidYearFrom || t.Year() > invalidYearTo {
		return time.Time{}, ErrInvalidYear
	}
	return t, err
}
