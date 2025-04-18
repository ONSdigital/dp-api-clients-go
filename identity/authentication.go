package identity

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/pkg/errors"
)

const service = "identity"

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

// Client is an identity client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Identity Client with a given zebedee url
func New(zebedeeURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, zebedeeURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls zebedee api health endpoint and returns a check object to the caller.
func (api Client) Checker(ctx context.Context, check *health.CheckState) error {
	return api.hcCli.Checker(ctx, check)
}

// AuthFailure is an alias to an error type, this represents the failure to
// authenticate request over a generic error from a http or marshalling error
type AuthFailure error

// CheckRequest calls the AuthAPI to check florenceToken or serviceAuthToken
func (api Client) CheckRequest(req *http.Request, florenceToken, serviceAuthToken string) (context.Context, int, AuthFailure, error) {
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
	var tokenIdentityResp *dprequest.IdentityResponse
	var errTokenIdentity error
	var authFail AuthFailure
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
	log.Info(ctx, "caller identity retrieved setting context values", logData)

	ctx = context.WithValue(ctx, dprequest.UserIdentityKey, userIdentity)
	ctx = context.WithValue(ctx, dprequest.CallerIdentityKey, tokenIdentityResp.Identifier)

	return ctx, http.StatusOK, nil, nil
}

// CheckTokenIdentity Checks the identity of a provided token, for a particular token type (i.e. user or service)
func (api Client) CheckTokenIdentity(ctx context.Context, token string, tokenType TokenType) (*dprequest.IdentityResponse, error) {
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

func (api Client) doCheckTokenIdentity(ctx context.Context, token string, tokenType TokenType, logData log.Data) (*dprequest.IdentityResponse, int, AuthFailure, error) {

	url := api.hcCli.URL + "/identity"
	logData["url"] = url
	log.Info(ctx, "calling AuthAPI to authenticate caller identity", logData)

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
		log.Error(ctx, "error creating AuthAPI identity http request", errCreatingReq, logData)
		return nil, http.StatusInternalServerError, nil, errCreatingReq
	}

	// 'GET /identity' request
	resp, err := api.hcCli.Client.Do(ctx, outboundAuthReq)
	if err != nil {
		log.Error(ctx, "HTTPClient.Do returned error making AuthAPI identity request", err, logData)
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
	tokenObj.hasPrefix = strings.HasPrefix(token, dprequest.BearerPrefix)

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

	if err := headers.SetAuthToken(outboundAuthReq, userAuthToken); err != nil {
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
func unmarshalIdentityResponse(resp *http.Response) (identityResp *dprequest.IdentityResponse, err error) {
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
		log.Error(ctx, "error closing response body", errClose, data)
	}
}

// getUserIdentity get the user identity. If the request is user driven return identityResponse.Identifier otherwise
// return the forwarded user Identity header from the inbound request. Return empty string if the header is not found.
func getUserIdentity(isUserReq bool, identityResp *dprequest.IdentityResponse, originalReq *http.Request) (string, error) {
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
