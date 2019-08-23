package times_test

import (
	"encoding/xml"
	"time"
	"fmt"

	"github.com/mantyr/times"
)

var MoscowLocation *time.Location

func init() {
	var err error
	MoscowLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(err)
	}
}

// CustomTime это метка времени в Europe/Moscow
// Особенности:
//   Принимает любую локаль и преобразует в Europe/Moscow
//   В XML возвращает xml.UTC
type CustomTime struct {
	times.Time
}

func NewCustomTime(s string) (*CustomTime, error) {
	t, err := times.NewTimeString(s, MoscowLocation)
	if err != nil {
		return nil, err
	}
	return &CustomTime{
		Time: *t,
	}, nil
}

// MarshalXML необходим для кодирования даты и времени в UTC
// Формат: YYYY-MM-DDThh:mm:ss.sssZ
func (t CustomTime) MarshalXML(d *xml.Encoder, start xml.StartElement) error {
	return t.CustomMarshalXML(d, start, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// MarshalXML необходим для кодирования даты и времени в UTC
// Формат: YYYY-MM-DDThh:mm:ss.sssZ
func (t CustomTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return t.CustomMarshalXMLAttr(name, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// UnmarshalXML необходим для декодирования даты и времени
// Входной формат:
//   YYYY-MM-DDThh:mm:ss.sssZ         - UTC
//   YYYY-MM-DDThh:mm:ss.sss+/-hh:mm  - локальное время UTC со смещением
//   YYYY-MM-DDThh:mm:ss.sss          - локальное время с часовым поясом Europe/Moscow по умолчанию
func (t *CustomTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return t.CustomUnmarshalXML(d, start, times.MoscowLocation)
}

// UnmarshalXMLAttr необходим для декодирования даты и времени
func (t *CustomTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return t.CustomUnmarshalXMLAttr(attr, times.MoscowLocation)
}

// Scan это реализация интерфейса database/sql.Scanner
func (t *CustomTime) Scan(src interface{}) error {
	return t.CustomScan(src, times.MoscowLocation)
}

func Example_customTime() {
	type Data struct {
		XMLName  xml.Name   `xml:"a"`
		DateAttr CustomTime `xml:"date,attr"`
		Date     CustomTime `xml:"date"`
	}
	d := &Data{}
	err := xml.Unmarshal([]byte(`<a date="2018-01-25T16:24:28Z"><date>2018-01-25T16:24:28+05:00</date></a>`), d)
	fmt.Println(err)
	fmt.Println(d.DateAttr)
	fmt.Println(d.DateAttr.Time.Time())
	fmt.Println(d.Date)
	fmt.Println(d.Date.Time.Time())

	data, err := xml.Marshal(d)
	fmt.Println(err)
	fmt.Println(string(data))
	// Output:
	// <nil>
	// 2018-01-25T19:24:28+03:00
	// 2018-01-25 19:24:28 +0300 MSK
	// 2018-01-25T14:24:28+03:00
	// 2018-01-25 14:24:28 +0300 MSK
	// <nil>
	// <a date="2018-01-25T16:24:28Z"><date>2018-01-25T11:24:28Z</date></a>
}
