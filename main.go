package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/danesparza/icaltoday/data"
)

var (
	// BuildVersion contains the version information for the app
	BuildVersion = "Unknown"

	// CommitID is the git commitId for the app.  It's filled in as
	// part of the automated build
	CommitID string
)

// Message is a custom struct event type to handle the Lambda input
type Message struct {
	CalendarURL string `json:"url"`
	Timezone    string `json:"timezone"`
}

// HandleRequest handles the AWS lambda request
func HandleRequest(ctx context.Context, msg Message) (data.CalendarResponse, error) {
	xray.Configure(xray.Config{LogLevel: "trace"})
	ctx, seg := xray.BeginSegment(ctx, "icaltoday-lambda-handler")

	service := data.CalService{}
	response, err := service.GetTodaysEvents(ctx, msg.CalendarURL, msg.Timezone)
	if err != nil {
		seg.Close(err)
		log.Fatalf("problem getting calendar events: %v", err)
	}

	//	Set the service version information:
	response.Version = fmt.Sprintf("%s.%s", BuildVersion, CommitID)

	//	Close the segment
	seg.Close(nil)

	//	Return our response
	return response, nil
}

func main() {
	//	Immediately forward to Lambda
	lambda.Start(HandleRequest)
}
