package identity

//go:generate moq -out mock/token_identity.go -pkg mock . TokenIdentity

import (
	"context"

	dprequest "github.com/ONSdigital/dp-net/v2/request"
)

// TokenIdentity is the Client used by the GraphQL package to make queries
type TokenIdentity interface {
	CheckTokenIdentity(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error)
}
