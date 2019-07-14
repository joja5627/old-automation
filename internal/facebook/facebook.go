package facebook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/testutils"
	"github.com/dghubble/sling"
	"golang.org/x/oauth2"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// Facebook login errors
var (
	ErrUnableToGetFacebookUser = errors.New("facebook: unable to get Facebook User")
)

const facebookAPI = "https://graph.facebook.com/v2.9/"

// User is a Facebook user.
//
// Note that user ids are unique to each app.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// client is a Facebook client for obtaining the current User.
type client struct {
	c     *http.Client
	sling *sling.Sling
}

func newClient(httpClient *http.Client) *client {
	base := sling.New().Client(httpClient).Base(facebookAPI)
	return &client{
		c:     httpClient,
		sling: base,
	}
}

func (c *client) Me() (*User, *http.Response, error) {
	user := new(User)
	// Facebook returns JSON as Content-Type text/javascript :(
	// Set Accept header to receive proper Content-Type application/json
	// so Sling will decode into the struct
	resp, err := c.sling.New().Set("Accept", "application/json").Get("me?fields=name,email").ReceiveSuccess(user)
	return user, resp, err
}

// StateHandler checks for a state cookie. If found, the state value is read
// and added to the ctx. Otherwise, a non-guessable value is added to the ctx
// and to a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func StateHandler(config gologin.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.StateHandler(config, success)
}

// LoginHandler handles Facebook login requests by reading the state value
// from the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler handles Facebook redirection URI requests and adds the
// Facebook access token and User to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure
// handler.
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = facebookHandler(config, success, failure)
	return oauth2Login.CallbackHandler(config, success, failure)
}

// facebookHandler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding Facebook User. If successful, the user is added to
// the ctx and the success handler is called. Otherwise, the failure handler
// is called.
func facebookHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		httpClient := config.Client(ctx, token)
		facebookService := newClient(httpClient)
		user, resp, err := facebookService.Me()
		err = validateResponse(user, resp, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given Facebook User, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(user *User, resp *http.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("facebook: unable to get Facebook user: %v", err)
	}
	if user == nil || user.ID == "" {
		return ErrUnableToGetFacebookUser
	}
	return nil
}

// WithUser returns a copy of ctx that stores the Facebook User.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Facebook User from the ctx.
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok {
		return nil, fmt.Errorf("facebook: Context missing Facebook User")
	}
	return user, nil
}

func TestFacebookHandler(t *testing.T) {
	jsonData := `{"id": "54638001", "name": "Ivy Crimson", "email": "ivy@harvard.edu"}`
	expectedUser := &User{ID: "54638001", Name: "Ivy Crimson", Email: "ivy@harvard.edu"}
	proxyClient, server := newFacebookTestServer(jsonData)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		facebookUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, facebookUser)
		fmt.Fprintf(w, "success handler called")
	}
	failure := testutils.AssertFailureNotCalled(t)

	// FacebookHandler assert that:
	// - Token is read from the ctx and passed to the facebook API
	// - facebook User is obtained from the facebook API
	// - success handler is called
	// - facebook User is added to the ctx of the success handler
	facebookHandler := facebookHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "success handler called", w.Body.String())
}

func TestFacebookHandler_MissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: Context missing Token", err.Error())
		}
		fmt.Fprintf(w, "failure handler called")
	}

	// FacebookHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	facebookHandler := facebookHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(w, req)
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestFacebookHandler_ErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("Facebook Service Down", http.StatusInternalServerError)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: "any-token"}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := gologin.ErrorFromContext(ctx)
		assert.Error(t, err)
		fmt.Fprintf(w, "failure handler called")
	}

	// FacebookHandler cannot get Facebook User, assert that:
	// - failure handler is called
	// - error cannot get Facebook User added to the failure handler ctx
	facebookHandler := facebookHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	facebookHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, "failure handler called", w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	validUser := &User{ID: "54638001", Name: "Ivy Crimson"}
	validResponse := &http.Response{StatusCode: 200}
	invalidResponse := &http.Response{StatusCode: 500}
	assert.Equal(t, nil, validateResponse(validUser, validResponse, nil))
	assert.Error(t, validateResponse(validUser, validResponse, fmt.Errorf("Server error")))
	assert.Error(t, validateResponse(validUser, invalidResponse, nil))
	assert.Equal(t, ErrUnableToGetFacebookUser, validateResponse(&User{}, validResponse, nil))
}
