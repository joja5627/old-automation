package facebook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextUser(t *testing.T) {
	expectedUser := &User{ID: "12", Name: "Gopher"}
	ctx := WithUser(context.Background(), expectedUser)
	user, err := UserFromContext(ctx)
	assert.Equal(t, expectedUser, user)
	assert.Nil(t, err)
}

func TestContextUser_Error(t *testing.T) {
	user, err := UserFromContext(context.Background())
	assert.Nil(t, user)
	if assert.NotNil(t, err) {
		assert.Equal(t, "facebook: Context missing Facebook User", err.Error())
	}
}

// func newFacebookTestServer(jsonData string) (*http.Client, *httptest.Server) {
// 	client, mux, server := testutils.TestServer()
// 	mux.HandleFunc("/v2.9/me", func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "application/json")
// 		fmt.Fprintf(w, jsonData)
// 	})
// 	return client, server
// }
