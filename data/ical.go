package data

import (
	"context"
	"fmt"
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
	if err := SetTimezone(timezone); err != nil {
		log.WithError(err).Error("Error setting timezone - most likely timezone data not loaded in host OS")
	}

	t := GetTime(time.Now())
	start := BeginningOfDay(t)
	end := EndOfDay(t)
	retval.CurrentLocalTime = GetTime(time.Now())
	log.WithFields(log.Fields{
		"start": start,
		"end":   end,
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
	c := gocal.NewParser(calendarDataResponse.Body)
	c.Start, c.End = &start, &end
	err = c.Parse()
	if err != nil {
		log.WithError(err).Error("Problem parsing calendar file")
		return retval, fmt.Errorf("problem parsing calendar file: %v", err)
	}

	//	Google calendar is so fucking weird.
	//	Regular events (even if they recur) are returned as UTC start/end times.  They can be converted to local time easily.  This is good.
	//	All-day events are going to be returned without a timezone.  A parser will assume a time of midnight IN THE UTC TIMEZONE.  This will lead to bullshit weirdness.
	for _, e := range c.Events {
		diff := e.End.Sub(*e.Start)

		//	If it looks like it's an all-day event, and the url includes 'calendar.google.com', then don't use the timezone
		if diff.Hours() > 23 && strings.Contains(url, "calendar.google.com") {

			log.WithFields(log.Fields{
				"url":         url,
				"summary":     e.Summary,
				"description": e.Description,
				"starttime":   e.Start.UTC(),
				"endtime":     e.End.UTC(),
			}).Info("Google all-day event detected.  Using the UTC start/end times")

			calEvent := CalendarEvent{
				UID:         e.Uid,
				Summary:     e.Summary,
				Description: e.Description,
				StartTime:   e.Start.UTC(),
				EndTime:     e.End.UTC(),
			}

			retval.Events = append(retval.Events, calEvent)
		} else {
			calEvent := CalendarEvent{
				UID:         e.Uid,
				Summary:     e.Summary,
				Description: e.Description,
				StartTime:   GetTime(*e.Start),
				EndTime:     GetTime(*e.End),
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

func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return BeginningOfDay(t).AddDate(0, 0, 1).Add(-time.Second)
}
