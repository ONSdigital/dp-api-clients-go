package clients

import (
	"sync"

	"github.com/ONSdigital/dp-net/http"
)

// APIClient represents a common structure for any api client
type APIClient struct {
	BaseURL    string
	HTTPClient http.Clienter
	Lock       sync.RWMutex
}
