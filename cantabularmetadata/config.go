package cantabularmetadata

import (
	"time"
)

// Config holds the config used to initialise the Cantabular Client
type Config struct {
	Host           string
	GraphQLTimeout time.Duration
}
