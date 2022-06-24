package identity

//go:generate moq -out token_identity_mock_test.go . TokenIdentity

import (
	"context"

	dprequest "github.com/ONSdigital/dp-net/v2/request"
)

// TokenIdentity is the Client used by the GraphQL package to make queries
type TokenIdentity interface {
	CheckTokenIdentity(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error)
}
