package clients

import (
	"sync"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
)

// APIClient represents a common structure for any api client
type APIClient struct {
	Check      *health.Check
	BaseURL    string
	HTTPClient rchttp.Clienter
	Lock       sync.RWMutex
}
