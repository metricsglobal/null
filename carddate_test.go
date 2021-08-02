package null

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

type testStruct struct {
	CardDate CardDate `json:"expiration_date" xml:"expiration_date"`
}

func TestNewCardDate(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		exp  string
	}{
		{
			name: "Valid 09/22",
			date: time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/22",
		},
		{
			name: "Valid 09/25",
			date: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/25",
		},
		{
			name: "Valid 09/32",
			date: time.Date(2032, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/32",
		},
		{
			name: "Valid 04/22",
			date: time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			exp:  "04/22",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := NewCardDate(test.date, true)
			if test.exp != exp.String() {
				t.Errorf("Expiration date should be %s, instead of %s", test.exp, exp.String())
			}
		})
	}
}

func TestCardDateFrom(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		exp  string
	}{
		{
			name: "Valid 09/22",
			date: time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/22",
		},
		{
			name: "Valid 09/25",
			date: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/25",
		},
		{
			name: "Valid 09/32",
			date: time.Date(2032, 9, 1, 0, 0, 0, 0, time.UTC),
			exp:  "09/32",
		},
		{
			name: "Valid 04/22",
			date: time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			exp:  "04/22",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := CardDateFrom(test.date)
			if test.exp != exp.String() {
				t.Errorf("Expiration date should be %s, instead of %s", test.exp, exp.String())
			}
		})
	}
}

func TestCardDateFromString(t *testing.T) {
	tests := []struct {
		name string
		date string
		exp  string
		err  error
	}{
		{
			name: "Valid with MM/YYYY",
			date: "09/2022",
			exp:  "09/22",
		},
		{
			name: "Valid with MM-YYYY",
			date: "09-2025",
			exp:  "09/25",
		},
		{
			name: "Invalid year",
			date: "09/2051",
			exp:  "09/25",
			err:  ErrInvalidYear,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp, err := CardDateFromString(test.date)
			if err != nil {
				if test.err == nil {
					t.Fatal(err)
				} else if test.err != err {
					t.Fatalf("Error should be %s, instead of %s", test.err.Error(), err.Error())
				}
				return
			}

			if test.exp != exp.String() {
				t.Errorf("Expiration date should be %s, instead of %s", test.exp, exp.String())
			}
		})
	}
}

func TestCardDateFromMustString(t *testing.T) {
	tests := []struct {
		name string
		date string
		exp  string
	}{
		{
			name: "Valid with MM/YYYY",
			date: "09/2022",
			exp:  "09/22",
		},
		{
			name: "Valid with MM-YYYY",
			date: "09-2025",
			exp:  "09/25",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := CardDateFromMustString(test.date)

			if test.exp != exp.String() {
				t.Errorf("Expiration date should be %s, instead of %s", test.exp, exp.String())
			}
		})
	}
}

func TestCardDateFromMustStringPanic(t *testing.T) {
	tests := []struct {
		name string
		date string
		exp  string
	}{
		{
			name: "Invalid year",
			date: "09/2051",
			exp:  "09/25",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("CardDateFromMustString should have panicked!")
					}
				}()
				CardDateFromMustString(test.date)
			}()

		})
	}
}

func TestCardDate_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		definition testStruct
		wantJSON   string
	}{
		{
			name:       "Marshal 09/22",
			definition: testStruct{CardDateFrom(time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC))},
			wantJSON:   `{"expiration_date":"09/22"}`,
		},
		{
			name:       "Marshal 09/32",
			definition: testStruct{CardDateFrom(time.Date(2032, 9, 1, 0, 0, 0, 0, time.UTC))},
			wantJSON:   `{"expiration_date":"09/32"}`,
		},
		{
			name:       "Marshal invalid",
			definition: testStruct{CardDate{Valid: false}},
			wantJSON:   `{"expiration_date":null}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marshaled, err := json.Marshal(test.definition)
			if err != nil {
				t.Fatal(err)
			}
			if string(marshaled) != test.wantJSON {
				t.Errorf("Json should be %s, instead of %s", test.wantJSON, string(marshaled))
			}
		})
	}
}

func TestCardDate_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		result   string
		validity bool
		err      error
	}{
		{
			name:     "JSON with format MM/YY",
			json:     `{"expiration_date":"09/22"}`,
			result:   "09/22",
			validity: true,
		},
		{
			name:     "JSON with format MM/YYYY",
			json:     `{"expiration_date":"09/2022"}`,
			result:   "09/22",
			validity: true,
		},
		{
			name:     "JSON with format RFC3339",
			json:     `{"expiration_date":"2023-11-30T00:00:00Z"}`,
			result:   "11/23",
			validity: true,
		},
		{
			name:     "JSON with format RFC822",
			json:     `{"expiration_date":"02 Nov 23 15:04 MST"}`,
			result:   "11/23",
			validity: true,
		},
		{
			name: "JSON with invalid year",
			json: `{"expiration_date":"02 Nov 55 15:04 MST"}`,
			err:  ErrInvalidYear,
		},
		{
			name: "JSON with invalid month",
			json: `{"expiration_date":"14/2022"}`,
			err:  ErrInvalidMonth,
		},
		{
			name:     "JSON with invalid month",
			json:     `{"expiration_date":null}`,
			validity: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resultStruct testStruct
			err := json.Unmarshal([]byte(test.json), &resultStruct)
			if err != nil {
				if test.err == nil {
					t.Fatal(err)
				} else if test.err != err {
					t.Fatalf("Error should be %s, instead of %s", test.err.Error(), err.Error())
				}
				return
			}

			if resultStruct.CardDate.Valid != test.validity {
				t.Fatal("Validity is not satisfy the test")
			}
			if !test.validity {
				return
			}
			if resultStruct.CardDate.String() != test.result {
				t.Errorf("Result should be %s, instead of %s", test.result, resultStruct.CardDate.String())
			}
		})
	}
}

func TestJSONUnmarshalForAllFormat(t *testing.T) {
	var tests = []struct {
		name   string
		exp    string // json representation of testStruct
		strExp string
		year   int
		month  int
		err    error
	}{
		{
			name:   "4 digit expiration date",
			exp:    `{"expiration_date": "0923"}`,
			strExp: "09/23",
			year:   2023,
			month:  9,
		},
		{
			name:   "5 digit expiration date",
			exp:    `{"expiration_date": "11/23"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC3339 format",
			exp:    `{"expiration_date": "2023-11-30T00:00:00Z"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC3339Nano format",
			exp:    `{"expiration_date": "2023-11-30T00:00:00.999999999Z"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC1123Z format",
			exp:    `{"expiration_date": "Mon, 30 Nov 2023 00:00:00 -0700"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC1123 format",
			exp:    `{"expiration_date": "Mon, 02 Nov 2023 15:04:05 MST"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC850 format",
			exp:    `{"expiration_date": "Monday, 02-Nov-23 15:04:05 MST"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC822Z format",
			exp:    `{"expiration_date": "02 Nov 23 15:04 -0700"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "RFC822 format",
			exp:    `{"expiration_date": "02 Nov 23 15:04 MST"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "Ruby Date format",
			exp:    `{"expiration_date": "Mon Nov 02 15:04:05 -0700 2023"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "Unix Date format",
			exp:    `{"expiration_date": "Mon Nov 2 15:04:05 MST 2023"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name:   "ANSIC format",
			exp:    `{"expiration_date": "Mon Nov 22 15:04:05 2023"}`,
			strExp: "11/23",
			year:   2023,
			month:  11,
		},
		{
			name: "Invalid format format",
			exp:  `{"expiration_date": "invalidformat"}`,
			err:  ErrUnknownFormat,
		},
		{
			name: "Invalid year",
			exp:  `{"expiration_date": "01/55"}`,
			err:  ErrInvalidYear,
		},
		{
			name: "Invalid month",
			exp:  `{"expiration_date": "13/22"}`,
			err:  ErrInvalidMonth,
		},
		{
			name:   "Valid month with 1",
			exp:    `{"expiration_date": "01/22"}`,
			strExp: "01/22",
			year:   2022,
			month:  1,
		},
		{
			name:   "Valid month with 12",
			exp:    `{"expiration_date": "12/22"}`,
			strExp: "12/22",
			year:   2022,
			month:  12,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var s testStruct
			err := json.Unmarshal([]byte(test.exp), &s)
			if err != test.err {
				t.Error(err)
				return
			}
			if test.err != nil {
				return
			}
			if s.CardDate.String() != test.strExp {
				t.Errorf("CardDate#String, got = %s, want %s", s.CardDate.String(), test.strExp)
			}
			if s.CardDate.Time.Year() != test.year {
				t.Errorf("CardDate#Time.Year(), got = %d, want %d", s.CardDate.Time.Year(), test.year)
			}
			if int(s.CardDate.Time.Month()) != test.month {
				t.Errorf("CardDate#Time.Month(), got = %d, want %d", int(s.CardDate.Time.Month()), test.month)
			}
		})
	}
}

func TestCardDate_MarshalText(t *testing.T) {
	tests := []struct {
		name       string
		definition testStruct
		wantXML    string
	}{
		{
			name:       "Marshal 09/22",
			definition: testStruct{CardDateFrom(time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC))},
			wantXML:    `<testStruct><expiration_date>09/22</expiration_date></testStruct>`,
		},
		{
			name:       "Marshal 09/32",
			definition: testStruct{CardDateFrom(time.Date(2032, 9, 1, 0, 0, 0, 0, time.UTC))},
			wantXML:    `<testStruct><expiration_date>09/32</expiration_date></testStruct>`,
		},
		{
			name:       "Marshal invalid",
			definition: testStruct{CardDate{Valid: false}},
			wantXML:    `<testStruct><expiration_date>null</expiration_date></testStruct>`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			marshaled, err := xml.Marshal(test.definition)
			if err != nil {
				t.Fatal(err)
			}
			if string(marshaled) != test.wantXML {
				t.Errorf("Json should be %s, instead of %s", test.wantXML, string(marshaled))
			}
		})
	}
}

func TestCardDate_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		result   string
		validity bool
		err      error
	}{
		{
			name:     "JSON with format MM/YY",
			xml:      `<testStruct><expiration_date>09/22</expiration_date></testStruct>`,
			result:   "09/22",
			validity: true,
		},
		{
			name:     "JSON with format MM/YYYY",
			xml:      `<testStruct><expiration_date>09/2022</expiration_date></testStruct>`,
			result:   "09/22",
			validity: true,
		},
		{
			name:     "JSON with format RFC3339",
			xml:      `<testStruct><expiration_date>2023-11-30T00:00:00Z</expiration_date></testStruct>`,
			result:   "11/23",
			validity: true,
		},
		{
			name:     "JSON with format RFC822",
			xml:      `<testStruct><expiration_date>02 Nov 23 15:04 MST</expiration_date></testStruct>`,
			result:   "11/23",
			validity: true,
		},
		{
			name: "JSON with invalid year",
			xml:  `<testStruct><expiration_date>02 Nov 55 15:04 MST</expiration_date></testStruct>`,
			err:  ErrInvalidYear,
		},
		{
			name: "JSON with invalid month",
			xml:  `<testStruct><expiration_date>14/2022</expiration_date></testStruct>`,
			err:  ErrInvalidMonth,
		},
		{
			name:     "JSON with invalid month",
			xml:      `<testStruct><expiration_date>null</expiration_date></testStruct>`,
			validity: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var resultStruct testStruct
			err := xml.Unmarshal([]byte(test.xml), &resultStruct)
			if err != nil {
				if test.err == nil {
					t.Fatal(err)
				} else if test.err != err {
					t.Fatalf("Error should be %s, instead of %s", test.err.Error(), err.Error())
				}
				return
			}

			if resultStruct.CardDate.Valid != test.validity {
				t.Fatal("Validity is not satisfy the test")
			}
			if !test.validity {
				return
			}
			if resultStruct.CardDate.String() != test.result {
				t.Errorf("Result should be %s, instead of %s", test.result, resultStruct.CardDate.String())
			}
		})
	}
}

func TestSetValid(t *testing.T) {
	date := time.Date(2020, 9, 0, 0, 0, 0, 0, time.UTC)
	want := CardDate{
		Time:  date,
		Valid: true,
	}
	got := CardDate{}
	got.SetValid(date)
	assertExprDate(t, want, got)
}

func TestSetValidFromStr(t *testing.T) {
	want := CardDate{
		Time:  time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
	got := CardDate{}
	got.SetValidFromStr("09/20")
	assertExprDate(t, want, got)
}

func TestCardDate_Scan(t *testing.T) {
	date := time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		want    CardDate
		got     interface{}
		wantErr bool
	}{
		{
			name: "Scan time",
			want: CardDate{
				Time:  date,
				Valid: true,
			},
			got: date,
		},
		{
			name: "Scan string",
			want: CardDate{
				Time:  date,
				Valid: true,
			},
			got: "09/20",
		},
		{
			name: "Scan nil",
			want: CardDate{
				Valid: false,
			},
			got: nil,
		},
		{
			name: "Scan other",
			want: CardDate{
				Valid: false,
			},
			got:     uint(123),
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CardDate{}
			err := got.Scan(test.got)
			if (err != nil) != test.wantErr {
				t.Fatal(err)
			}
			assertExprDate(t, test.want, got)
		})
	}
}

func TestCardDate_Value(t *testing.T) {
	date := time.Date(2020, 9, 0, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		got  CardDate
		want driver.Value
	}{
		{
			name: "Invalid input",
			got:  CardDate{},
			want: nil,
		},
		{
			name: "Valid input",
			got: CardDate{
				Time:  date,
				Valid: true,
			},
			want: date,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, _ := test.got.Value()
			if !reflect.DeepEqual(value, test.want) {
				t.Errorf("ExprDate value expected %v, instead of %v", test.want, value)
			}
		})
	}
}

func TestAddDate(t *testing.T) {
	tests := []struct {
		name   string
		value  CardDate
		month  int
		year   int
		result string
	}{
		{
			name:   "Add 2 month",
			value:  CardDateFrom(time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)),
			month:  2,
			result: "11/23",
		},
		{
			name:   "Add 8 month",
			value:  CardDateFrom(time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)),
			month:  8,
			result: "05/24",
		},
		{
			name:   "Add 2 year and 4 month",
			value:  CardDateFrom(time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)),
			year:   2,
			month:  4,
			result: "01/26",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := test.value.AddDate(test.year, test.month, 0)
			if exp.String() != test.result {
				t.Errorf("Exp should be %s, instead of %s", test.result, exp.String())
			}
		})
	}
}

func TestGenerateCardDate(t *testing.T) {
	exp := GenerateCardDate(func() int64 {
		return 5
	}, "", false)

	result := "06/25"
	if exp.String() != result {
		t.Errorf("Exp should be %s, instead of %s", result, exp.String())
	}

	expInvalid := GenerateCardDate(func() int64 {
		return 5
	}, "", true)

	if expInvalid.Valid {
		t.Error("Generated one should be invalid")
	}
}

func TestParseExpToTime(t *testing.T) {
	tests := []struct {
		name    string
		exp     string
		want    time.Time
		wantErr bool
	}{
		{
			name: "parse default format 1",
			exp:  "09/22",
			want: time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "parse default format 2",
			exp:  "11/23",
			want: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "parse default format 3",
			exp:     "11/23",
			want:    time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "parse RFC3339 format 1",
			exp:  "2019-10-12T07:20:50Z",
			want: time.Date(2019, 10, 12, 7, 20, 50, 0, time.UTC),
		},
		{
			name: "parse RFC3339 format 2",
			exp:  "2019-10-12T07:20:50+00:00",
			want: time.Date(2019, 10, 12, 7, 20, 50, 0, time.UTC),
		},
		{
			name: "parse RFC3339 format 3",
			exp:  "2019-10-12T14:20:50+07:00",
			want: time.Date(2019, 10, 12, 7, 20, 50, 0, time.UTC),
		},
		{
			name: "parse RFC3339 format 4",
			exp:  "2019-10-12T03:20:50-04:00",
			want: time.Date(2019, 10, 12, 7, 20, 50, 0, time.UTC),
		},
		{
			name: "parse RFC3339 format 5",
			exp:  "2019-10-31T23:20:50-04:00",
			want: time.Date(2019, 11, 1, 3, 20, 50, 0, time.UTC),
		},
		{
			name: "parse RFC3339 format 6",
			exp:  "2019-10-31T23:20:50-04:30",
			want: time.Date(2019, 11, 1, 3, 50, 50, 0, time.UTC),
		},
		{
			name: "parse MMYYYY",
			exp:  "112023",
			want: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "parse MM-YY",
			exp:  "11-23",
			want: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "parse MM-YYYY",
			exp:  "11-2023",
			want: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "parse MM/YYYY",
			exp:  "11/2023",
			want: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid expiration date length",
			exp:     "010230",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "invalid expiration date",
			exp:     "2051-10-31T23:20:50-04:30",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "invalid year in expiration date",
			exp:     "0151",
			want:    time.Time{},
			wantErr: true,
		},
		{
			name:    "invalid month expiration date",
			exp:     "13/22",
			want:    time.Time{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpToTime(tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseExpToTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToTime(t *testing.T) {
	tests := []struct {
		name    string
		exp     string
		year    int
		month   int
		wantErr bool
	}{
		{
			name:  "4 digit expiration date",
			exp:   "0923",
			year:  2023,
			month: 9,
		},
		{
			name:  "5 digit expiration date",
			exp:   "11/23",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC3339 format",
			exp:   "2023-11-30T00:00:00Z",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC3339Nano format",
			exp:   "2023-11-30T00:00:00.999999999Z",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC1123Z format",
			exp:   "Mon, 30 Nov 2023 00:00:00 -0700",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC1123 format",
			exp:   "Mon, 02 Nov 2023 15:04:05 MST",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC850 format",
			exp:   "Monday, 02-Nov-23 15:04:05 MST",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC822Z format",
			exp:   "02 Nov 23 15:04 -0700",
			year:  2023,
			month: 11,
		},
		{
			name:  "RFC822 format",
			exp:   "02 Nov 23 15:04 MST",
			year:  2023,
			month: 11,
		},
		{
			name:  "Ruby Date format",
			exp:   "Mon Nov 02 15:04:05 -0700 2023",
			year:  2023,
			month: 11,
		},
		{
			name:  "Unix Date format",
			exp:   "Mon Nov 2 15:04:05 MST 2023",
			year:  2023,
			month: 11,
		},
		{
			name:  "ANSIC format",
			exp:   "Mon Nov 22 15:04:05 2023",
			year:  2023,
			month: 11,
		},
		{
			name:    "Invalid format format",
			exp:     "invalidformat",
			wantErr: true,
		},
		{
			name:    "Invalid year",
			exp:     "01/55",
			wantErr: true,
		},
		{
			name:    "Invalid month",
			exp:     "13/22",
			wantErr: true,
		},
		{
			name:  "Valid month with 1",
			exp:   "01/22",
			year:  2022,
			month: 1,
		},
		{
			name:  "Valid month with 12",
			exp:   "12/22",
			year:  2022,
			month: 12,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := stringToTime(test.exp)
			if (err != nil) != test.wantErr {
				t.Errorf("stringToTime() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if test.wantErr {
				// NOTE: we don't want to process checks in case of error
				return
			}

			if got.Year() != test.year {
				t.Errorf("stringToTime(), year -> got = %d, want %d", got.Year(), test.year)
			}
			if int(got.Month()) != test.month {
				t.Errorf("stringToTime(), month -> got = %d, want %d", int(got.Month()), test.month)
			}
		})
	}
}

func TestTimeParserError(t *testing.T) {
	if _, err := timeParser("01/06", "lalala"); err == nil {
		t.Error("error should be exists since it's not a parsable date")
	}

}

func assertExprDate(t *testing.T, expected CardDate, got CardDate) {
	if expected.Time != got.Time {
		t.Errorf("ExprDate time expexcted %v, instead of %v", expected.Time, got.Time)
	}
	if expected.String() != got.String() {
		t.Errorf("ExprDate string expexcted %v, instead of %v", expected.String(), got.String())
	}
	if expected.Valid != got.Valid {
		t.Errorf("ExprDate valid expexcted %v, instead of %v", expected.Valid, got.Valid)
	}
}
