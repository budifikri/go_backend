package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

var clockTimeLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("GMT+7", 7*60*60)
	}
	return loc
}()

type ClockTime struct {
	time.Time
}

func NewClockTime(hour, minute, second int) ClockTime {
	return ClockTime{
		Time: time.Date(2000, time.January, 1, hour, minute, second, 0, clockTimeLocation),
	}
}

func ParseClockTime(value string) (ClockTime, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ClockTime{}, fmt.Errorf("empty time")
	}

	layouts := []string{"15:04", "15:04:05"}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, trimmed, clockTimeLocation)
		if err == nil {
			return NewClockTime(parsed.Hour(), parsed.Minute(), parsed.Second()), nil
		}
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		localized := parsed.In(clockTimeLocation)
		return NewClockTime(localized.Hour(), localized.Minute(), localized.Second()), nil
	}

	return ClockTime{}, fmt.Errorf("invalid time format")
}

func (ct ClockTime) Value() (driver.Value, error) {
	if ct.IsZero() {
		return nil, nil
	}
	return ct.Format("15:04:05"), nil
}

func (ct *ClockTime) Scan(value interface{}) error {
	if value == nil {
		ct.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		ct.Time = NewClockTime(v.Hour(), v.Minute(), v.Second()).Time
		return nil
	case []byte:
		parsed, err := ParseClockTime(string(v))
		if err != nil {
			return err
		}
		ct.Time = parsed.Time
		return nil
	case string:
		parsed, err := ParseClockTime(v)
		if err != nil {
			return err
		}
		ct.Time = parsed.Time
		return nil
	default:
		return fmt.Errorf("unsupported ClockTime scan type %T", value)
	}
}

func (ClockTime) GormDataType() string {
	return "time"
}

func (ct ClockTime) MarshalJSON() ([]byte, error) {
	if ct.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, ct.Format("15:04"))), nil
}

func (ct *ClockTime) UnmarshalJSON(data []byte) error {
	trimmed := strings.Trim(string(data), `"`)
	if trimmed == "" || trimmed == "null" {
		ct.Time = time.Time{}
		return nil
	}
	parsed, err := ParseClockTime(trimmed)
	if err != nil {
		return err
	}
	ct.Time = parsed.Time
	return nil
}

func (ct ClockTime) After(other ClockTime) bool {
	return ct.toSeconds() > other.toSeconds()
}

func (ct ClockTime) Format(layout string) string {
	if ct.IsZero() {
		return ""
	}
	return time.Date(2000, time.January, 1, ct.Hour(), ct.Minute(), ct.Second(), 0, clockTimeLocation).Format(layout)
}

func (ct ClockTime) toSeconds() int {
	return (ct.Hour() * 3600) + (ct.Minute() * 60) + ct.Second()
}
