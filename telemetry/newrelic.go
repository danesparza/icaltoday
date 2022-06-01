package telemetry

import (
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
	log "github.com/sirupsen/logrus"
)

var (
	NRLicense = "Unknown"
	NRAppName = "Dashboard"
	NRApp     = &newrelic.Application{}
)

func init() {

	err := *new(error)

	//	If we have NR environment variables, use them:
	if os.Getenv("NR_DASHBOARD_LIC") != "" {
		NRLicense = os.Getenv("NR_DASHBOARD_LIC")
	}

	if os.Getenv("NR_DASHBOARD_APP") != "" {
		NRAppName = os.Getenv("NR_DASHBOARD_APP")
	}

	NRApp, err = newrelic.NewApplication(
		newrelic.ConfigAppName(NRAppName),
		newrelic.ConfigLicense(NRLicense),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		log.WithFields(log.Fields{
			"App name":            NRApp,
			"License information": NRLicense,
			"error":               err,
		}).Error("Error trying to create New Relic connection")
	}
}
