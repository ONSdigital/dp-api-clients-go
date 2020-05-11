package identity

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	clients "github.com/ONSdigital/dp-api-clients-go"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"

	"github.com/pkg/errors"
)

var errUnableToIdentifyRequest = errors.New("unable to determine the user or service making the request")

type tokenObject struct {
	numberOfParts int
	hasPrefix     bool
	tokenPart     string
}

// TokenType iota enum defines possible token types
type TokenType int

// Possible Token types
const (
	TokenTypeUser TokenType = iota
	TokenTypeService
)

var tokenTypes = []string{"User", "Service"}

// Values of the token types
func (t TokenType) String() string {
	return tokenTypes[t]
}

// Client is an alias to a generic/common api client structure
type Client clients.APIClient

// Clienter provides an interface to checking identity of incoming request
type Clienter interface {
	CheckRequest(req *http.Request) (context.Context, int, authFailure, error)
}

// NewAPIClient returns a Client
func NewAPIClient(cli dphttp.Clienter, url string) (api *Client) {
	return &Client{
		HTTPClient: cli,
		BaseURL:    url,
	}
}

// authFailure is an alias to an error type, this represents the failure to
// authenticate request over a generic error from a http or marshalling error
type authFailure error

// CheckRequest calls the AuthAPI to check florenceToken or serviceAuthToken
func (api Client) CheckRequest(req *http.Request, florenceToken, serviceAuthToken string) (context.Context, int, authFailure, error) {
	ctx := req.Context()

	isUserReq := len(florenceToken) > 0
	isServiceReq := len(serviceAuthToken) > 0

	// if neither user nor service request, return unchanged ctx
	if !isUserReq && !isServiceReq {
		return ctx, http.StatusUnauthorized, errors.WithMessage(errUnableToIdentifyRequest, "no headers set on request"), nil
	}

	logData := log.Data{
		"is_user_request":    isUserReq,
		"is_service_request": isServiceReq,
	}
	splitTokens(florenceToken, serviceAuthToken, logData)

	// Check token identity (according to isUserReq or isServiceReq)
	var tokenIdentityResp *common.IdentityResponse
	var errTokenIdentity error
	var authFail authFailure
	var statusCode int
	if isUserReq {
		tokenIdentityResp, statusCode, authFail, errTokenIdentity = api.doCheckTokenIdentity(ctx, florenceToken, TokenTypeUser, logData)
	} else {
		tokenIdentityResp, statusCode, authFail, errTokenIdentity = api.doCheckTokenIdentity(ctx, serviceAuthToken, TokenTypeService, logData)
	}
	if errTokenIdentity != nil || authFail != nil {
		return ctx, statusCode, authFail, errTokenIdentity
	}

	// If token identity succeeded, get user identity
	userIdentity, err := getUserIdentity(isUserReq, tokenIdentityResp, req)
	if err != nil {
		return ctx, http.StatusInternalServerError, nil, err
	}

	logData["user_identity"] = userIdentity
	logData["caller_identity"] = userIdentity
	log.Event(ctx, "caller identity retrieved setting context values", log.INFO, logData)

	ctx = context.WithValue(ctx, common.UserIdentityKey, userIdentity)
	ctx = context.WithValue(ctx, common.CallerIdentityKey, tokenIdentityResp.Identifier)

	return ctx, http.StatusOK, nil, nil
}

// CheckTokenIdentity Checks the identity of a provided token, for a particular token type (i.e. user or service)
func (api Client) CheckTokenIdentity(ctx context.Context, token string, tokenType TokenType) (*common.IdentityResponse, error) {
	if len(token) == 0 {
		return nil, errors.New("Empty token provided")
	}
	// Log data with token type
	logData := log.Data{
		"token_type": tokenType.String(),
		"token":      splitToken(token),
	}
	// Perform 'GET /identity' and simplify return
	idRes, _, authErr, err := api.doCheckTokenIdentity(ctx, token, tokenType, logData)
	if authErr != nil {
		return nil, authErr
	}
	return idRes, err
}

func (api Client) doCheckTokenIdentity(ctx context.Context, token string, tokenType TokenType, logData log.Data) (*common.IdentityResponse, int, authFailure, error) {

	url := api.BaseURL + "/identity"
	logData["url"] = url
	log.Event(ctx, "calling AuthAPI to authenticate caller identity", log.INFO, logData)

	// Crete request according to the token type
	var outboundAuthReq *http.Request
	var errCreatingReq error
	switch tokenType {
	case TokenTypeUser:
		outboundAuthReq, errCreatingReq = createUserAuthRequest(url, token)
	default:
		outboundAuthReq, errCreatingReq = createServiceAuthRequest(url, token)
	}
	if errCreatingReq != nil {
		log.Event(ctx, "error creating AuthAPI identity http request", log.ERROR, logData, log.Error(errCreatingReq))
		return nil, http.StatusInternalServerError, nil, errCreatingReq
	}

	// Create client if it does not exist
	if api.HTTPClient == nil {
		api.Lock.Lock()
		api.HTTPClient = dphttp.NewClient()
		api.Lock.Unlock()
	}

	// 'GET /identity' request
	resp, err := api.HTTPClient.Do(ctx, outboundAuthReq)
	if err != nil {
		log.Event(ctx, "HTTPClient.Do returned error making AuthAPI identity request", log.ERROR, logData, log.Error(err))
		return nil, http.StatusInternalServerError, nil, err
	}
	defer closeResponse(ctx, resp, logData)

	// Validate returned status code
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, errors.WithMessage(errUnableToIdentifyRequest, "unexpected status code returned from AuthAPI"), nil
	}

	// Unmarshal response
	idResp, err := unmarshalIdentityResponse(resp)
	return idResp, http.StatusInternalServerError, nil, err
}

func splitTokens(florenceToken, authToken string, logData log.Data) {
	if len(florenceToken) > 0 {
		logData["florence_token"] = splitToken(florenceToken)
	}
	if len(authToken) > 0 {
		logData["auth_token"] = splitToken(authToken)
	}
}

func splitToken(token string) (tokenObj tokenObject) {
	splitToken := strings.Split(token, " ")
	tokenObj.numberOfParts = len(splitToken)
	tokenObj.hasPrefix = strings.HasPrefix(token, common.BearerPrefix)

	// sample last 6 chars (or half, if smaller) of last token part
	lastTokenPart := len(splitToken) - 1
	tokenSampleStart := len(splitToken[lastTokenPart]) - 6
	if tokenSampleStart < 1 {
		tokenSampleStart = len(splitToken[lastTokenPart]) / 2
	}
	tokenObj.tokenPart = splitToken[lastTokenPart][tokenSampleStart:]

	return tokenObj
}

func createUserAuthRequest(url string, userAuthToken string) (*http.Request, error) {
	outboundAuthReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if err := headers.SetUserAuthToken(outboundAuthReq, userAuthToken); err != nil {
		return nil, err
	}

	return outboundAuthReq, nil
}

func createServiceAuthRequest(url string, serviceAuthToken string) (*http.Request, error) {
	outboundAuthReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if err := headers.SetServiceAuthToken(outboundAuthReq, serviceAuthToken); err != nil {
		return nil, err
	}

	return outboundAuthReq, nil
}

// unmarshalIdentityResponse converts a resp.Body (JSON) into an IdentityResponse
func unmarshalIdentityResponse(resp *http.Response) (identityResp *common.IdentityResponse, err error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &identityResp)
	return
}

func closeResponse(ctx context.Context, resp *http.Response, data log.Data) {
	if resp == nil || resp.Body == nil {
		return
	}

	if errClose := resp.Body.Close(); errClose != nil {
		log.Event(ctx, "error closing response body", log.ERROR, log.Error(errClose), data)
	}
}

// getUserIdentity get the user identity. If the request is user driven return identityResponse.Identifier otherwise
// return the forwarded user Identity header from the inbound request. Return empty string if the header is not found.
func getUserIdentity(isUserReq bool, identityResp *common.IdentityResponse, originalReq *http.Request) (string, error) {
	var userIdentity string
	var err error

	if isUserReq {
		userIdentity = identityResp.Identifier
	} else {
		forwardedUserIdentity, errGetIdentityHeader := headers.GetUserIdentity(originalReq)
		if errGetIdentityHeader == nil {
			userIdentity = forwardedUserIdentity
		} else if headers.IsNotErrNotFound(errGetIdentityHeader) {
			err = errGetIdentityHeader
		}
	}

	return userIdentity, err
}
