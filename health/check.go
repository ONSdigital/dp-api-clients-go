package health

import (
	"context"
	"time"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

var (
	statusDescription = map[string]string{
		health.StatusOK:       "Everything is ok",
		health.StatusWarning:  "Things are degraded, but at least partially functioning",
		health.StatusCritical: "The checked functionality is unavailable or non-functioning",
	}

	unixTime = time.Unix(1494505756, 0)
)

func getCheck(ctx *context.Context, name string, code int) *health.Check {

	currentTime := time.Now().UTC()

	check := &health.Check{
		Name:        name,
		StatusCode:  code,
		LastChecked: currentTime,
		LastSuccess: unixTime,
		LastFailure: unixTime,
	}

	switch code {
	case 200:
		check.Message = statusDescription[health.StatusOK]
		check.Status = health.StatusOK
		check.LastSuccess = currentTime
	case 429:
		check.Message = statusDescription[health.StatusWarning]
		check.Status = health.StatusWarning
		check.LastFailure = currentTime
	default:
		check.Message = statusDescription[health.StatusCritical]
		check.Status = health.StatusCritical
		check.LastFailure = currentTime
	}

	return check
}
