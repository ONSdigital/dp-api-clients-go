// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package cantabular

import (
	"context"
	"sync"
)

// Ensure, that GraphQLClientMock does implement GraphQLClient.
// If this is not the case, regenerate this file with moq.
var _ GraphQLClient = &GraphQLClientMock{}

// GraphQLClientMock is a mock implementation of GraphQLClient.
//
// 	func TestSomethingThatUsesGraphQLClient(t *testing.T) {
//
// 		// make and configure a mocked GraphQLClient
// 		mockedGraphQLClient := &GraphQLClientMock{
// 			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
// 				panic("mock out the Query method")
// 			},
// 		}
//
// 		// use mockedGraphQLClient in code that requires GraphQLClient
// 		// and then make assertions.
//
// 	}
type GraphQLClientMock struct {
	// QueryFunc mocks the Query method.
	QueryFunc func(ctx context.Context, query interface{}, vars map[string]interface{}) error

	// calls tracks calls to the methods.
	calls struct {
		// Query holds details about calls to the Query method.
		Query []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Query is the query argument value.
			Query interface{}
			// Vars is the vars argument value.
			Vars map[string]interface{}
		}
	}
	lockQuery sync.RWMutex
}

// Query calls QueryFunc.
func (mock *GraphQLClientMock) Query(ctx context.Context, query interface{}, vars map[string]interface{}) error {
	if mock.QueryFunc == nil {
		panic("GraphQLClientMock.QueryFunc: method is nil but GraphQLClient.Query was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Query interface{}
		Vars  map[string]interface{}
	}{
		Ctx:   ctx,
		Query: query,
		Vars:  vars,
	}
	mock.lockQuery.Lock()
	mock.calls.Query = append(mock.calls.Query, callInfo)
	mock.lockQuery.Unlock()
	return mock.QueryFunc(ctx, query, vars)
}

// QueryCalls gets all the calls that were made to Query.
// Check the length with:
//     len(mockedGraphQLClient.QueryCalls())
func (mock *GraphQLClientMock) QueryCalls() []struct {
	Ctx   context.Context
	Query interface{}
	Vars  map[string]interface{}
} {
	var calls []struct {
		Ctx   context.Context
		Query interface{}
		Vars  map[string]interface{}
	}
	mock.lockQuery.RLock()
	calls = mock.calls.Query
	mock.lockQuery.RUnlock()
	return calls
}
