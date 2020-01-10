package health

import (
	"context"
	"time"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	healthyMessage  = "service is OK"
	warningMessage  = "service is warming up or downgraded"
	criticalMessage = "service is in critical state"
	notFoundMessage = "received status code 404, unable to find health check endpoint"
)

var unixTime = time.Unix(0, 0)

func getCheck(ctx context.Context, service, status, errorMessage string, code int) *health.Check {

	currentTime := time.Now().UTC()

	check := &health.Check{
		Name:        service,
		StatusCode:  code,
		Status:      status,
		LastChecked: currentTime,
		LastSuccess: unixTime,
		LastFailure: unixTime,
	}

	switch status {
	case health.StatusOK:
		check.LastSuccess = currentTime
		check.Message = healthyMessage
	case health.StatusWarning:
		check.LastFailure = currentTime
		check.Message = warningMessage
	default:
		check.LastFailure = currentTime

		switch code {
		case 200:
			check.Message = criticalMessage
		case 404:
			check.Message = notFoundMessage
		default:
			check.Message = errorMessage
		}
	}

	return check
}
