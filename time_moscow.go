package times

import (
	"encoding/xml"
	"time"
)

var MoscowLocation *time.Location

func init() {
	MoscowLocation, _ = time.LoadLocation("Europe/Moscow")
}

// MoscowTime это метка времени в Europe/Moscow
// Особенности:
//   Принимает любую локаль и преобразует в Europe/Moscow
//   В XML возвращает UTC
type MoscowTime struct {
	Time
}

func NewMoscowTime(t time.Time) (*MoscowTime, error) {
	date, err := NewTime(t, MoscowLocation)
	if err != nil {
		return nil, err
	}
	return &MoscowTime{
		Time: *date,
	}, nil
}

func NewMoscowTimeString(s string) (*MoscowTime, error) {
	t, err := NewTimeString(s, MoscowLocation)
	if err != nil {
		return nil, err
	}
	return &MoscowTime{
		Time: *t,
	}, nil
}

// MarshalXML необходим для кодирования даты и времени в UTC
// Формат: YYYY-MM-DDThh:mm:ss.sssZ
func (t MoscowTime) MarshalXML(d *xml.Encoder, start xml.StartElement) error {
	return t.CustomMarshalXML(d, start, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// MarshalXML необходим для кодирования даты и времени в UTC
// Формат: YYYY-MM-DDThh:mm:ss.sssZ
func (t MoscowTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return t.CustomMarshalXMLAttr(name, time.UTC, "2006-01-02T15:04:05Z07:00")
}

// UnmarshalXML необходим для декодирования даты и времени
// Входной формат:
//   YYYY-MM-DDThh:mm:ss.sssZ         - UTC
//   YYYY-MM-DDThh:mm:ss.sss+/-hh:mm  - локальное время UTC со смещением
//   YYYY-MM-DDThh:mm:ss.sss          - локальное время с часовым поясом Europe/Moscow по умолчанию
func (t *MoscowTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return t.CustomUnmarshalXML(d, start, MoscowLocation)
}

// UnmarshalXMLAttr необходим для декодирования даты и времени
func (t *MoscowTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return t.CustomUnmarshalXMLAttr(attr, MoscowLocation)
}

// MarshalJSON необходим для кодирования даты и времени
func (t MoscowTime) MarshalJSON() ([]byte, error) {
	return t.CustomMarshalJSON(time.UTC, "2006-01-02T15:04:05Z07:00")
}

// UnmarshalJSON необходим для декодирования даты и времени
func (t *MoscowTime) UnmarshalJSON(data []byte) error {
	return t.CustomUnmarshalJSON(data, MoscowLocation)
}

// Scan это реализация интерфейса database/sql.Scanner
func (t *MoscowTime) Scan(src interface{}) error {
	return t.CustomScan(src, MoscowLocation)
}
