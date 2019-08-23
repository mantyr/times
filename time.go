package times

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"errors"
	"time"
)

// Time это метка времени
//
// Используется стандарт ISO 8601 (YYYY-MM-DDThh:mm:ss)
//   YYYY-MM-DDThh:mm:ss.sssZ        - UTC
//   YYYY-MM-DDThh:mm:ss.sss+/-hh:mm - локальное UTC время со смещением
//   YYYY-MM-DDThh:mm:ss.sss         - локальное время в часовом поясе по умолчанию
// Пример:
//   «2005-08-09T18:31:42» - «9 августа 2005 года 18 часов 31 минута 42 секунды»
//
// Часовой пояс по умолчанию:
//   Без указания часового пояса время передается в часовом поясе UTC.
//   Для кастомизации часового пояса по умолчанию см. пример в example_custom_time_test.go
//
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
type Time time.Time

// NewTime возвращает модифицированную метку времени
// на основе стандартной метки времени в UTC
func NewTime(t time.Time, location *time.Location) (*Time, error) {
	if location == nil {
		return nil, errors.New("empty time location")
	}
	newTime := Time(t.In(location))
	return &newTime, nil
}

// NewCurrentTime возвращает текущее время в UTC
func NewCurrentTime() (*Time, error) {
	return NewTime(time.Now(), time.UTC)
}

// NewTimeString возвращает время на основе строки в location Europe/Moscow
func NewTimeString(date string, location *time.Location) (*Time, error) {
	t := &Time{}
	err := t.setTimeString(date, location)
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
	return t.CustomScan(src, time.UTC)
}

// CustomScan это реализация интерфейса database/sql.Scanner
func (t *Time) CustomScan(src interface{}, location *time.Location) error {
	switch v := src.(type) {
	case time.Time:
		return t.setTime(v, location)
	case string:
		return t.setTimeString(v, location)
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
	return t.CustomUnmarshalXML(d, start, time.UTC)
}

// CustomUnmarshalXML необходим для декодирования даты и времени
func (t *Time) CustomUnmarshalXML(
	d *xml.Decoder,
	start xml.StartElement,
	location *time.Location,
) error {
	var data string
	err := d.DecodeElement(&data, &start)
	if err != nil {
		return err
	}
	return t.setTimeString(data, location)
}

// UnmarshalXMLAttr необходим для декодирования даты и времени
func (t *Time) UnmarshalXMLAttr(attr xml.Attr) error {
	return t.CustomUnmarshalXMLAttr(attr, time.UTC)
}

// CustomUnmarshalXMLAttr необходим для декодирования даты и времени
func (t *Time) CustomUnmarshalXMLAttr(attr xml.Attr, location *time.Location) error {
	return t.setTimeString(attr.Value, location)
}

// UnmarshalJSON необходим для декодирования даты и времени
func (t *Time) UnmarshalJSON(data []byte) error {
	return t.CustomUnmarshalJSON(data, time.UTC)
}

// CustomUnmarshalJSON необходим для декодирования даты и времени
func (t *Time) CustomUnmarshalJSON(data []byte, location *time.Location) error {
	if location == nil {
		return errors.New("empty time location")
	}
	var date string
	err := json.Unmarshal(data, &date)
	if err != nil {
		return err
	}
	return t.setTimeString(date, location)
}


// setLocalTime устанавливает локальное время
func (t *Time) setTime(date time.Time, location *time.Location) error {
	if location == nil {
		return errors.New("empty time location")
	}
	*t = Time(date.In(location))
	return nil
}

// setTimeString устанавливает время из строки
func (t *Time) setTimeString(data string, location *time.Location) error {
	if location == nil {
		return errors.New("empty time location")
	}
	if data == "" {
		*t = Time(time.Time{}.In(location))
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
	return t.Time().String()
}

// MarshalXML необходим для кодирования даты и времени
func (t Time) MarshalXML(d *xml.Encoder, start xml.StartElement) error {
	return t.CustomMarshalXML(d, start, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// CustomMarshalXML необходим для кодирования даты и времени
func (t Time) CustomMarshalXML(
	d *xml.Encoder,
	start xml.StartElement,
	location *time.Location,
	format string,
) error {
	if location == nil {
		return errors.New("empty time location")
	}
	return d.EncodeElement(
		t.Time().In(location).Format(format),
		start,
	)
}

// MarshalXMLAttr необходим для кодирования даты и времени
func (t Time) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return t.CustomMarshalXMLAttr(name, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// CustomMarshalXMLAttr необходим для кодирования даты и времени
func (t Time) CustomMarshalXMLAttr(
	name xml.Name,
	location *time.Location,
	format string,
) (
	xml.Attr,
	error,
) {
	if location == nil {
		return xml.Attr{}, errors.New("empty time location")
	}
	return xml.Attr{
		Name: name,
		Value: t.Time().In(location).Format(format),
	}, nil
}

// MarshalJSON необходим для кодирования даты и времени
func (t Time) MarshalJSON() ([]byte, error) {
	return t.CustomMarshalJSON(time.UTC, "2006-01-02T15:04:05Z07:00")
}

// MarshalJSON необходим для кодирования даты и времени
func (t Time) CustomMarshalJSON(
	location *time.Location,
	format string,
) (
	[]byte,
	error,
) {
	if location == nil {
		return []byte{}, errors.New("empty time location")
	}
	return json.Marshal(
		t.Time().In(location).Format(format),
	)
}

// Equal сравнивает две даты
func (t Time) Equal(y Time) bool {
	return reflect.DeepEqual(t, y)
}

// DeepEqual сравнивает две даты в не зависимости от time.Location
func (t Time) DeepEqual(y Time) bool {
	left := t.Time().In(time.UTC)
	right := y.Time().In(time.UTC)
	return reflect.DeepEqual(left, right)
}

// EqualTime сравнивает две даты, одна из которых time.Time
func (t Time) EqualTime(y time.Time) bool {
	return t.Equal(Time(y))
}
