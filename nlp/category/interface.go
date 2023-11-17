package category

import (
	"context"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/category/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/category/models"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out ./mocks.go -pkg category . Clienter

type Clienter interface {
	Checker(ctx context.Context, check *health.CheckState) error
	GetCategory(ctx context.Context, options Options) (*[]models.Category, errors.Error)
	Health() *healthcheck.Client
	URL() string
}
