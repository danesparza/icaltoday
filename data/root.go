package data

import "time"

type CalendarResponse struct {
	TimeZone         string          `json:"timezone"`         // The timezone used
	CurrentLocalTime time.Time       `json:"currentlocaltime"` // Sanity check:  Current local time in the timezone given
	Events           []CalendarEvent `json:"events"`           // The calendar events found
	Version          string          `json:"version"`          // Service version
}

type CalendarEvent struct {
	UID         string    `json:"uid"`         // Unique event id
	Summary     string    `json:"summary"`     // Event summary
	Description string    `json:"description"` // Event long description
	StartTime   time.Time `json:"starttime"`   // Event start time
	EndTime     time.Time `json:"endtime"`     // Event end time
}
