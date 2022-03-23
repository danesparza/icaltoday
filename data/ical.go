package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apognu/gocal"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/context/ctxhttp"
)

// CalService is a service to fetch iCalendar formatted data
type CalService struct{}

// GetTodaysEvents gets today's events from the given ical calendar url and the timezone.
func (s CalService) GetTodaysEvents(ctx context.Context, url, timezone string) (CalendarResponse, error) {
	//	Start the service segment
	ctx, seg := xray.BeginSubsegment(ctx, "ical-service")

	//	Our return value
	retval := CalendarResponse{}
	retval.Events = []CalendarEvent{} // Initialize the array
	retval.TimeZone = timezone

	//	Set our start / end times
	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"timezone": timezone,
		}).WithError(err).Error("Error setting location from the timezone - most likely timezone data not loaded in host OS")
	}

	//	Current time in the location
	t := time.Now().In(location)

	//	Find the beginning and end of the day in the given timezone
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, location)

	retval.CurrentLocalTime = t
	log.WithFields(log.Fields{
		"currentLocalTime": t,
		"start":            start,
		"end":              end,
		"timezone":         timezone,
		"url":              url,
	}).Info("Time debugging")

	//	First, get the ical calendar at the url given
	clientRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		seg.AddError(err)
		return retval, fmt.Errorf("problem creating request to the ical url given: %v", err)
	}

	//	Set our headers
	clientRequest.Header.Set("Content-Type", "application/geo+json; charset=UTF-8")

	//	Execute the request
	client := &http.Client{}
	calendarDataResponse, err := ctxhttp.Do(ctx, xray.Client(client), clientRequest)
	if err != nil {
		seg.AddError(err)
		return retval, fmt.Errorf("error when sending request to get the calendar data from the url: %v", err)
	}
	defer calendarDataResponse.Body.Close()

	//	Create a parser and use our start/end times
	calEvents, err := GetEventsForDay(calendarDataResponse.Body, start, end, location)
	if err != nil {
		return retval, err
	}

	//	Track our event IDs
	eventIDs := make(map[string]struct{})

	//	Google calendar is so fucking weird.
	//	Regular events (even if they recur) are returned as UTC start/end times.  They can be converted to local time easily.  This is good.
	//	All-day events are going to be returned without a timezone.  A parser will assume a time of midnight IN THE UTC TIMEZONE.  This will lead to bullshit weirdness.
	for _, e := range calEvents {

		//	If we have duplicate event ids (and we shouldn't ... but I've seen this in the wild) discard anything after the first instance
		//	See https://stackoverflow.com/a/10486196/19020 for more info on using an empty struct in a map to track this
		if _, containsEvent := eventIDs[e.Uid]; containsEvent {
			continue
		} else {
			eventIDs[e.Uid] = struct{}{}
		}

		diff := e.End.Sub(*e.Start)

		//	If it looks like it's an all-day event, and the url includes 'calendar.google.com', then don't use the timezone
		if diff.Hours() > 23 && strings.Contains(url, "calendar.google.com") {

			log.WithFields(log.Fields{
				"url":                 url,
				"summary":             e.Summary,
				"description":         e.Description,
				"starttime":           e.Start.UTC(),
				"endtime":             e.End.UTC(),
				"rewritten-starttime": RewriteToLocal(e.Start.UTC(), location),
				"rewritten-endtime":   RewriteToLocal(e.End.UTC(), location),
			}).Info("Google all-day event detected.  Using the UTC start/end times and rewriting them as local")

			calEvent := CalendarEvent{
				UID:         e.Uid,
				Summary:     e.Summary,
				Description: e.Description,
				StartTime:   RewriteToLocal(e.Start.UTC(), location), // Rewrite the UTC time to appear as local time
				EndTime:     RewriteToLocal(e.End.UTC(), location),   // Rewrite the UTC time to appear as local time
			}

			retval.Events = append(retval.Events, calEvent)
		} else {
			calEvent := CalendarEvent{
				UID:         e.Uid,
				Summary:     e.Summary,
				Description: e.Description,
				StartTime:   e.Start.In(location),
				EndTime:     e.End.In(location),
			}

			retval.Events = append(retval.Events, calEvent)
		}
	}

	//	Add the report to the request metadata
	xray.AddMetadata(ctx, "CalendarResult", retval)

	// Close the segment
	seg.Close(nil)

	//	Return the report
	return retval, nil
}

// RewriteToLocal - rewrites a given time to use the passed location data
func RewriteToLocal(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
}

// GetEventsForDay gets events for the day given the calendar body and start/end times
func GetEventsForDay(calendarBody io.Reader, start, end time.Time, location *time.Location) ([]gocal.Event, error) {

	//	Our return value:
	retval := []gocal.Event{}

	//	Create a parser and use our start/end times
	c := gocal.NewParser(calendarBody)
	c.Start, c.End = &start, &end
	c.AllDayEventsTZ = location // Also give it a location

	err := c.Parse()
	if err != nil {
		return retval, fmt.Errorf("problem parsing calendar file: %v", err)
	}

	return c.Events, nil
}

func init() {
	log.SetLevel(log.DebugLevel)
}
