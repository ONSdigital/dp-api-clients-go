package berlin

import (
	"context"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/models"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out ./mocks.go -pkg berlin . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *health.CheckState) error
	GetBerlin(ctx context.Context, options Options) (*models.Berlin, errors.Error)
	Health() *healthcheck.Client
	URL() string
}
