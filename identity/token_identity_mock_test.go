// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package identity

import (
	"context"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"sync"
)

// Ensure, that TokenIdentityMock does implement TokenIdentity.
// If this is not the case, regenerate this file with moq.
var _ TokenIdentity = &TokenIdentityMock{}

// TokenIdentityMock is a mock implementation of TokenIdentity.
//
// 	func TestSomethingThatUsesTokenIdentity(t *testing.T) {
//
// 		// make and configure a mocked TokenIdentity
// 		mockedTokenIdentity := &TokenIdentityMock{
// 			CheckTokenIdentityFunc: func(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error) {
// 				panic("mock out the CheckTokenIdentity method")
// 			},
// 		}
//
// 		// use mockedTokenIdentity in code that requires TokenIdentity
// 		// and then make assertions.
//
// 	}
type TokenIdentityMock struct {
	// CheckTokenIdentityFunc mocks the CheckTokenIdentity method.
	CheckTokenIdentityFunc func(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// CheckTokenIdentity holds details about calls to the CheckTokenIdentity method.
		CheckTokenIdentity []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Token is the token argument value.
			Token string
			// TokenType is the tokenType argument value.
			TokenType TokenType
		}
	}
	lockCheckTokenIdentity sync.RWMutex
}

// CheckTokenIdentity calls CheckTokenIdentityFunc.
func (mock *TokenIdentityMock) CheckTokenIdentity(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error) {
	if mock.CheckTokenIdentityFunc == nil {
		panic("TokenIdentityMock.CheckTokenIdentityFunc: method is nil but TokenIdentity.CheckTokenIdentity was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		Token     string
		TokenType TokenType
	}{
		Ctx:       ctx,
		Token:     token,
		TokenType: tokenType,
	}
	mock.lockCheckTokenIdentity.Lock()
	mock.calls.CheckTokenIdentity = append(mock.calls.CheckTokenIdentity, callInfo)
	mock.lockCheckTokenIdentity.Unlock()
	return mock.CheckTokenIdentityFunc(ctx, token, tokenType)
}

// CheckTokenIdentityCalls gets all the calls that were made to CheckTokenIdentity.
// Check the length with:
//     len(mockedTokenIdentity.CheckTokenIdentityCalls())
func (mock *TokenIdentityMock) CheckTokenIdentityCalls() []struct {
	Ctx       context.Context
	Token     string
	TokenType TokenType
} {
	var calls []struct {
		Ctx       context.Context
		Token     string
		TokenType TokenType
	}
	mock.lockCheckTokenIdentity.RLock()
	calls = mock.calls.CheckTokenIdentity
	mock.lockCheckTokenIdentity.RUnlock()
	return calls
}