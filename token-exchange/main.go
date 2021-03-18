package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
	client *http.Client
)

func main() {
	logger = logrus.New()

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	sessionResolvers := map[string]SessionResolver{
		"github": githubSessionResolver,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization") // "Bearer github 123456"

		authorizationBits := strings.SplitN(authorization, " ", 3)

		if len(authorizationBits) != 3 {
			logger.WithField("authorization", authorization).Warn("unexpected auth format")
			w.WriteHeader(500)
			return
		}

		provider := authorizationBits[1] // github
		token := authorizationBits[2]    // 123456

		sessionResolver, ok := sessionResolvers[provider]
		if !ok {
			logger.WithField("provider", provider).Warn("unknown provider")
			w.WriteHeader(500)
			return
		}

		session, err := sessionResolver(token)
		if err != nil {
			logger.WithError(err).WithField("provider", provider).Warn("failed to resolve session")
			w.WriteHeader(500)
			return
		}

		logger.WithField("session", session).Info("authenticated")
		json.NewEncoder(w).Encode(session.AuthSession())
	})

	os.Stdout.WriteString("Listening at :3001\n")

	panic(http.ListenAndServe(":3001", nil))
}

// SessionResolver turns some token into an AuthSession
type SessionResolver func(token string) (AuthSessionable, error)

func githubSessionResolver(token string) (AuthSessionable, error) {
	return resolveGithubUser(client, token)
}

func resolveGithubUser(client *http.Client, token string) (*GithubUser, error) {
	request, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "token "+token)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("expected 200 but received %d", resp.StatusCode)
	}

	userResponse := &GithubUser{}

	if err = json.NewDecoder(resp.Body).Decode(userResponse); err != nil {
		return nil, err
	}

	return userResponse, nil
}

// GithubUser returned by /user Github API call, trimmed down to usable values
type GithubUser struct {
	// ID unique across all users for your OAuth app
	ID         int    `json:"id"`
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	// Real name: Richard
	Name string `json:"name"`
	// Login username: RMS
	Login string `json:"login"`
	// Preferred email: rms@stallman.org
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthSession converts a GithubUser into our common format accepted by Oathkeeper
func (user *GithubUser) AuthSession() AuthSession {
	return AuthSession{
		Subject: strconv.Itoa(user.ID),
		Extra: AuthSessionExtra{
			Provider: "github",
			Username: user.Login,
			Email:    user.Email,
		},
	}
}

// AuthSessionable things can be turned into AuthSessions
type AuthSessionable interface {
	AuthSession() AuthSession
}

// AuthSession is our common auth format accepted by Oathkeeper
type AuthSession struct {
	Subject string           `json:"sub"`
	Extra   AuthSessionExtra `json:"extra"`
}

// AuthSessionExtra includes all auth data that isn't the user ID
type AuthSessionExtra struct {
	Provider string `json:"provider"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
