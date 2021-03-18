package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	clientID := os.Getenv("OATHKEEPER_GH_DEMO_CLIENT_ID")
	if clientID == "" {
		logger.Fatal("OATHKEEPER_GH_DEMO_CLIENT_ID is required")
	}

	apiURL := os.Getenv("OATHKEEPER_GH_DEMO_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost/"
		logger.WithField("default", apiURL).Info("OATHKEEPER_GH_DEMO_API_URL was not set, using default")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var accessToken string

	persistedAccessToken, err := ioutil.ReadFile("access-token")
	if err == nil {
		accessToken = string(persistedAccessToken)
		logger.WithField("access-token", accessToken).Info("read persisted access token")
	} else {
		logger.WithError(err).Warn("failed reading persisted access token, will regenerate")
	}

	if accessToken == "" {
		// generate device code
		deviceCodeGenerationTime := time.Now()
		deviceCodeResponse, err := generateDeviceCode(client, clientID, "read:user user:email")
		if err != nil {
			logger.WithError(err).Fatal("failed to generate device code")
		}
		deviceCodeExpirationTime := deviceCodeGenerationTime.Add(time.Duration(deviceCodeResponse.ExpiresIn) * time.Second)
		deviceCodeFetchInterval := (time.Duration(deviceCodeResponse.Interval) * time.Second) + 2*time.Second // add a 2s buffer on top of minimum interval to avoid ratelimit

		logger.WithField("code", deviceCodeResponse.UserCode).WithField("url", deviceCodeResponse.VerificationURI).Info("please enter this code :)")

		// exchange device code for access token
		for {
			time.Sleep(deviceCodeFetchInterval)
			if time.Now().After(deviceCodeExpirationTime) {
				logger.WithField("expiration-time", deviceCodeExpirationTime).Fatal("device code is expired, please retry")
			}

			accessTokenResponse, err := pollDeviceCode(client, clientID, deviceCodeResponse.DeviceCode)
			if err != nil {
				if exchangeErr, ok := err.(*AccessTokenExchangeError); ok {
					logger.WithFields(logrus.Fields{
						"error-code":        exchangeErr.Response.Error,
						"error-description": exchangeErr.Response.ErrorDescription,
						"error-uri":         exchangeErr.Response.ErrorURI,
					}).Warn("poll failed with expected error")
				} else {
					logger.WithError(err).Warn("poll failed with unexpected error")
				}
				continue
			}

			accessToken = accessTokenResponse.AccessToken
			break
		}

		if err := ioutil.WriteFile("access-token", []byte(accessToken), os.ModePerm); err != nil {
			logger.WithError(err).Warn("failed writing access token to disk, continuing")
		}

		logger.WithField("access-token", accessToken).Info("retrieved new token!")
	}

	// make API requests with our access token
	output, err := hitOurCoolAPI(client, apiURL, accessToken)
	if err != nil {
		logger.WithError(err).Fatal("failed to hit our API")
	}

	logger.WithField("output", output).Info("hit our API :)")
}

func generateDeviceCode(client *http.Client, clientID string, scope string) (*DeviceCodeResponse, error) {
	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("scope", scope)

	request, err := http.NewRequest("POST", "https://github.com/login/device/code", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	deviceCodeResponse := &DeviceCodeResponse{}

	if err = json.NewDecoder(resp.Body).Decode(deviceCodeResponse); err != nil {
		return nil, err
	}

	return deviceCodeResponse, nil
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func pollDeviceCode(client *http.Client, clientID string, deviceCode string) (*AccessTokenResponse, error) {
	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("device_code", deviceCode)
	form.Add("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	request, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	bufferedBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	accessTokenResponse := &AccessTokenResponse{}

	if err = json.NewDecoder(bytes.NewReader(bufferedBody)).Decode(accessTokenResponse); err != nil {
		return nil, err
	}

	// json.NewDecoder will decode any valid JSON.
	// if this isn't a successfull access token response, .AccessToken will be empty
	// decode it into an error response when this happens
	if accessTokenResponse.AccessToken == "" {
		errorResponse := AccessTokenErrorResponse{}

		if err = json.NewDecoder(bytes.NewReader(bufferedBody)).Decode(&errorResponse); err != nil {
			return nil, err
		}

		if errorResponse.Error != "" {
			return nil, &AccessTokenExchangeError{
				Response: errorResponse,
			}
		} else {
			return nil, errors.New("unknown access token response")
		}
	}

	return accessTokenResponse, nil
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type AccessTokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

type AccessTokenExchangeError struct {
	Response AccessTokenErrorResponse
}

func (err *AccessTokenExchangeError) Error() string {
	return err.Response.ErrorDescription
}

func hitOurCoolAPI(client *http.Client, url string, accessToken string) (string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("Accept", "text/plain")
	request.Header.Set("Authorization", "Bearer github "+accessToken)

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("non-200 status received, check Oathkeeper logs, something probably exploded")
	}

	bufferedBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bufferedBody), nil
}
