package gografana

import (
	"fmt"
	"net/http"
)

type Authenticator interface {
	SetAuthentication(req *http.Request)
}

type userPassAuthenticator struct {
	User string
	Pass string
}

func NewBasicAuthenticator(user string, pass string) Authenticator {
	return &userPassAuthenticator{User: user, Pass: pass}
}

func (u *userPassAuthenticator) SetAuthentication(req *http.Request) {
	req.SetBasicAuth(u.User, u.Pass)
}

type apiKeyAuthenticator struct {
	ApiKey string
}

func NewAPIKeyAuthenticator(apiKey string) Authenticator {
	return &apiKeyAuthenticator{ApiKey: apiKey}
}

func (u *apiKeyAuthenticator) SetAuthentication(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", u.ApiKey))
}
