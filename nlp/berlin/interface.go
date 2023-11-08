package berlin

import (
	"context"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-scrubber-api/models"
	"github.com/ONSdigital/dp-search-scrubber-api/sdk/errors"
)

//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *health.CheckState) error
	GetBerlin(ctx context.Context, options Options) (*models.ScrubberResp, errors.Error)
	Health() *healthcheck.Client
	URL() string
}
