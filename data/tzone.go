package data

import "time"

var loc *time.Location

func SetTimezone(tz string) error {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	loc = location
	return nil
}

func GetTime(t time.Time) time.Time {
	return t.In(loc)
}
