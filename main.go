package main

import (
	"context"
	"fmt"
	"log"

	"github.com/danesparza/icaltoday/data"
	"github.com/newrelic/go-agent/v3/integrations/nrlambda"
	"github.com/newrelic/go-agent/v3/newrelic"
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

	txn := newrelic.FromContext(ctx)
	defer txn.StartSegment("icaltoday HandleRequest").End()

	service := data.CalService{}
	response, err := service.GetTodaysEvents(ctx, msg.CalendarURL, msg.Timezone)
	if err != nil {
		log.Fatalf("problem getting calendar events: %v", err)
	}

	//	Set the service version information:
	response.Version = fmt.Sprintf("%s.%s", BuildVersion, CommitID)

	//	Return our response
	return response, nil
}

func main() {
	// Pass nrlambda.ConfigOption() into newrelic.NewApplication to set
	// Lambda specific configuration settings including
	// Config.ServerlessMode.Enabled.
	app, err := newrelic.NewApplication(nrlambda.ConfigOption())
	if nil != err {
		fmt.Println("error creating app (invalid config):", err)
	}
	// nrlambda.Start should be used in place of lambda.Start.
	// nrlambda.StartHandler should be used in place of lambda.StartHandler.
	nrlambda.Start(HandleRequest, app)
}
