package times

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Time это метка времени
//
// Используется стандарт ISO 8601 (YYYY-MM-DDThh:mm:ss)
// Пример:
//   «2005-08-09T18:31:42» - «9 августа 2005 года 18 часов 31 минута 42 секунды»
//
// Без указания часового пояса время передается в часовом поясе г. Москвы
// Возможно указание конкретного часового пояса (YYYY-MM-DDThh:mm:ss±hh:mm)
// Пример:
//   2012-06-01T12:00:00-03:00
// По факту может прийти время с долями секунды, например:
//   2018-01-25T16:24:28.74
//
// Поддерживает следующие форматы при Unmarshalling (XML/JSON):
//   2006-01-02T15:04:05Z07:00
//   пример: 2018-02-01T14:12:18+03:00
//
//   2006-01-02T15:04:05.999999999Z07:00 - обрезание до секунды
//   пример: 2018-02-01T14:12:18.47+03:00
//
//   2006-01-02T15:04:05
//   пример: 2018-02-01T14:12:18
//
//   2006-01-02T15:04:05.999999999Z - обрезание до секунды
//   пример: 2018-02-01T14:12:18.47
// При Unmarshalling (XML/JSON) выставляет дату в локаль Europe/Moscow
type Time time.Time

// NewTime возвращает модифицированную метку времени
// на основе стандартной метки времени в location Europe/Moscow
func NewTime(t time.Time) (*Time, error) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, fmt.Errorf("load location error: %v", err)
	}
	newTime := Time(t.In(location))
	return &newTime, err
}

// NewCurrentTime возвращает текущее время в location Europe/Moscow
func NewCurrentTime() (*Time, error) {
	return NewTime(time.Now())
}

// NewTimeString возвращает время на основе строки в location Europe/Moscow
func NewTimeString(date string) (*Time, error) {
	t, err := NewTime(time.Time{})
	if err != nil {
		return nil, err
	}
	err = t.setLocalTimeString(date)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Time возвращает исходный объект времени
func (t Time) Time() time.Time {
	return time.Time(t)
}

// Scan это реализация интерфейса database/sql.Scanner
func (t *Time) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time:
		return t.setLocalTime(v)
	case string:
		return t.setLocalTimeString(v)
	}
	return fmt.Errorf("expected value type time.Time or string but actual %T", src)
}

// Value это реализация database/sql/driver.Valuer
func (t Time) Value() (driver.Value, error) {
	return t.Time(), nil
}

// Format возвращает отформатированную дату и время
// Функция принимает первый layout
// В случае если layout не задан - используется формат по умолчанию
// Формат по умолчанию: "2006-01-02 15:04:05 MST"
func (t Time) Format(layout ...string) string {
	if len(layout) == 0 {
		return t.Time().Format("2006-01-02 15:04:05 MST")
	}
	return t.Time().Format(layout[0])
}

// UnmarshalXML необходим для декодирования даты и времени
func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var data string
	err := d.DecodeElement(&data, &start)
	if err != nil {
		return err
	}
	return t.setLocalTimeString(data)
}

// UnmarshalJSON необходим для декодирования даты и времени
func (t *Time) UnmarshalJSON(data []byte) error {
	var date string
	err := json.Unmarshal(data, &date)
	if err != nil {
		return err
	}
	return t.setLocalTimeString(date)
}

// setLocalTime устанавливает локальное время
func (t *Time) setLocalTime(date time.Time) error {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}
	*t = Time(date.In(location))
	return nil
}

// setLocalTimeString устанавливает локальное время из строки
func (t *Time) setLocalTimeString(data string) error {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return err
	}
	if data == "" {
		*t = Time(time.Time{})
		return nil
	}
	localTime, err := time.ParseInLocation("2006-01-02T15:04:05", data, location)
	if err != nil {
		if strings.Contains(err.Error(), "extra text") {
			localTime, err = time.Parse("2006-01-02T15:04:05Z07:00", data)
		}
	}
	if err != nil {
		return err
	}
	*t = Time(localTime.In(location))
	return nil
}

// Add возвращает t+duration
func (t Time) Add(duration time.Duration) Time {
	return Time(time.Time(t).Add(duration))
}

// UntilEndMonthDays возвращает количество дней
// от текущей даты до конца текущего месяца
func (t Time) UntilEndMonthDays() int {
	start := time.Time(t)
	day := start.Day()
	end := start.AddDate(0, 0, -(day-1))
	end = end.AddDate(0, +1, 0)
	diff := end.Sub(start)
	
	result := diff/(24*time.Hour)
	return int(result)
}

// UntilEndNextMonthDays возвращает количество дней
// от текущей даты до конца следующего месяца
func (t Time) UntilEndNextMonthDays() int {
	start := time.Time(t)
	day := start.Day()
	end := start.AddDate(0, 0, -(day-1))
	end = end.AddDate(0, +2, 0)
	diff := end.Sub(start)
	
	result := diff/(24*time.Hour)
	return int(result)
}

// String возвращает текстовое представление
func (t Time) String() string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}

// MarshalXML необходим для кодирования даты и времени
func (t Time) MarshalXML(d *xml.Encoder, start xml.StartElement) error {
	return d.EncodeElement(t.Format("2006-01-02T15:04:05Z07:00"), start)
}

// MarshalJSON необходим для кодирования даты и времени
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format("2006-01-02T15:04:05Z07:00"))
}

// Equal сравнивает две даты
func (t Time) Equal(y Time) bool {
	return reflect.DeepEqual(t, y)
}

// EqualTime сравнивает две даты, одна из которых time.Time
func (t Time) EqualTime(y time.Time) bool {
	return t.Equal(Time(y))
}
