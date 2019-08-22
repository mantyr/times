package times

import (
	"encoding/json"
	"encoding/xml"
	"testing"
	"time"
	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

func testUnmarshalXMLTime(source, expected string) {
	requestData := []byte(`
		<?xml version="1.0" encoding="utf-8"?>
		<Body>
			<DATE>` + source + `</DATE>
		</Body>
	`)
	var data struct {
		XMLName xml.Name `xml:"Body"`
		Date    Time     `xml:"DATE"`
	}
	err := xml.Unmarshal(requestData, &data)
	So(err, ShouldBeNil)
	So(
		data.Date.Time().Format("2006-01-02T15:04:05Z07:00"),
		ShouldEqual,
		expected,
	)
}

func testUnmarshalJSONTime(source, expected string) {
	requestData := []byte(`
		{"date": "` + source + `"}
	`)
	var data struct {
		Date Time `json:"date"`
	}
	err := json.Unmarshal(requestData, &data)
	So(err, ShouldBeNil)
	So(
		data.Date.Time().Format("2006-01-02T15:04:05Z07:00"),
		ShouldEqual,
		expected,
	)
}

func testMarshalXMLTime(source time.Time, expected string) {
	expected = `<Body><DATE>` + expected + `</DATE></Body>`

	var request struct {
		XMLName xml.Name `xml:"Body"`
		Date    *Time    `xml:"DATE"`
	}
	var err error
	request.Date, err = NewTime(source)
	So(err, ShouldBeNil)

	data, err := xml.Marshal(request)
	So(err, ShouldBeNil)
	So(
		string(data),
		ShouldEqual,
		expected,
	)
}

func testMarshalJSONTime(source time.Time, expected string) {
	expected = `{"date":"` + expected + `"}`

	var request struct {
		Date *Time `json:"date"`
	}
	var err error
	request.Date, err = NewTime(source)
	So(err, ShouldBeNil)

	data, err := json.Marshal(request)
	So(err, ShouldBeNil)
	So(
		string(data),
		ShouldEqual,
		expected,
	)
}

type unmarshalFunction func(source, expected string)
type marshalFunction func(source time.Time, expected string)

func TestTime(t *testing.T) {
	testTime(
		t,
		"XML",
		testUnmarshalXMLTime,
		testMarshalXMLTime,
	)
	testTime(
		t,
		"JSON",
		testUnmarshalJSONTime,
		testMarshalJSONTime,
	)
}

func testTime(
	t *testing.T,
	testName string,
	unmarshalFunc unmarshalFunction,
	marshalFunc marshalFunction,
) {
	Convey(testName, t, func() {
		Convey("Проверяем декодирование метки времени", func() {
			Convey("Пустая дата ", func() {
				unmarshalFunc(
					"",
					"0001-01-01T00:00:00Z",
				)
			})
			Convey("Без указания часового пояса", func() {
				unmarshalFunc(
					"2018-01-25T16:24:28",
					"2018-01-25T16:24:28+03:00",
				)
				Convey("С миллисекундами", func() {
					unmarshalFunc(
						"2018-01-25T16:24:28.74",
						"2018-01-25T16:24:28+03:00",
					)
				})
			})
			Convey("С указанием часового пояса и привидением к MSK", func() {
				unmarshalFunc(
					"2018-01-25T16:24:28+05:00",
					"2018-01-25T14:24:28+03:00",
				)
				Convey("С миллисекундами", func() {
					unmarshalFunc(
						"2018-01-25T16:24:28.74+05:00",
						"2018-01-25T14:24:28+03:00",
					)
				})
			})
		})
		Convey("Проверяем кодирование метки времени", func() {
			Convey("Проверяем интерфейс Stringer", func() {
				t, err := NewTime(time.Time{})
				So(err, ShouldBeNil)
				result := t.String()
				So(
					result,
					ShouldEqual,
					"0001-01-01T02:30:17+02:30",
				)
			})
			Convey("Пустое время, time.Time{}", func() {
				marshalFunc(
					time.Time{},
					"0001-01-01T02:30:17+02:30",
				)
			})
			Convey("Локальное время", func() {
				location, err := time.LoadLocation("Europe/Moscow")
				So(err, ShouldBeNil)
				So(location, ShouldNotBeNil)

				localTime, err := time.ParseInLocation("2006-01-02T15:04:05", "2018-02-01T14:12:18", location)
				So(err, ShouldBeNil)
				marshalFunc(
					localTime,
					"2018-02-01T14:12:18+03:00",
				)
			})
			Convey("Время без указания локали", func() {
				localTime, err := time.Parse("2006-01-02T15:04:05Z07:00", "2018-02-01T14:12:18+03:00")
				So(err, ShouldBeNil)
				marshalFunc(
					localTime,
					"2018-02-01T14:12:18+03:00",
				)
			})
		})
	})
}

func TestTimeFormat(t *testing.T) {
	Convey("Проверяем форматирование даты", t, func() {
		localDate := time.Now().Local()
		date, err := NewTime(localDate)
		So(err, ShouldBeNil)
		Convey("Формат по умолчанию", func() {
			So(
				date.Format(),
				ShouldEqual,
				localDate.Format("2006-01-02 15:04:05 MST"),
			)
		})
		Convey("Кастомный формат", func() {
			So(
				date.Format("2006-01-02T15:04:05Z07:00"),
				ShouldEqual,
				localDate.Format("2006-01-02T15:04:05Z07:00"),
			)

		})
	})
}

func TestUntilEndMonthDays(t *testing.T) {
	Convey("Проверяем количество дней до конца текущего месяца", t, func() {
		testUntilEndMonthDays("2018-02-01T00:00:00", 28)
		testUntilEndMonthDays("2018-02-15T14:00:04", 14)
		testUntilEndMonthDays("2018-05-01T00:00:00", 31)
		testUntilEndMonthDays("2018-05-01T12:15:55", 31)
		testUntilEndMonthDays("2018-05-02T12:15:55", 30)
		testUntilEndMonthDays("2018-05-03T12:15:55", 29)
		testUntilEndMonthDays("2018-05-04T12:15:55", 28)
		testUntilEndMonthDays("2018-12-25T12:15:55", 7)
		testUntilEndMonthDays("2018-12-30T12:15:55", 2)
	})
}

func testUntilEndMonthDays(startDate string, expected int) {
	name := fmt.Sprintf("%s -> %d", startDate, expected)
	Convey(name, func() {
		t, err := NewTimeString(startDate)
		So(err, ShouldBeNil)
		So(t, ShouldNotBeNil)
		result := t.UntilEndMonthDays()
		So(
			result,
			ShouldEqual,
			expected,
		)
	})
}

func TestUntilEndNextMonthDays(t *testing.T) {
	Convey("Проверяем количество дней до конца следующего месяца", t, func() {
		testUntilEndNextMonthDays("2018-02-01T00:00:00", 59)
		testUntilEndNextMonthDays("2018-02-15T14:00:04", 45)
		testUntilEndNextMonthDays("2018-05-01T00:00:00", 61)
		testUntilEndNextMonthDays("2018-05-01T12:15:55", 61)
		testUntilEndNextMonthDays("2018-05-02T12:15:55", 60)
		testUntilEndNextMonthDays("2018-05-03T12:15:55", 59)
		testUntilEndNextMonthDays("2018-05-04T12:15:55", 58)
		testUntilEndNextMonthDays("2018-12-25T12:15:55", 38)
		testUntilEndNextMonthDays("2018-12-30T12:15:55", 33)
	})
}

func testUntilEndNextMonthDays(startDate string, expected int) {
	name := fmt.Sprintf("%s -> %d", startDate, expected)
	Convey(name, func() {
		t, err := NewTimeString(startDate)
		So(err, ShouldBeNil)
		So(t, ShouldNotBeNil)
		result := t.UntilEndNextMonthDays()
		So(
			result,
			ShouldEqual,
			expected,
		)
	})
}

