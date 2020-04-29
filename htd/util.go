package htd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const msisAuthCookieName string = "MSISAuth"
const jSessionIdCookieName string = "JSESSIONID"
const dateFormat string = "02/01/2006" // Golang: why can't you just be normal?

// Get the cookie named `name` from `cookies`
func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// Get the URL used for authentication
func getVafsAuthUrl() (*url.URL, error) {
	authUrl, err := url.Parse("https://vafs.nus.edu.sg/adfs/oauth2/authorize")
	if err != nil {
		return nil, err
	}
	queryStrings := url.Values{
		"response_type": {"code"},
		"client_id":     {"97F0D1CACA7D41DE87538F9362924CCB-184318"},
		"resource":      {"sg_edu_nus_oauth"},
		"redirect_uri":  {"https://myaces.nus.edu.sg:443/htd/htd"},
	}
	authUrl.RawQuery = queryStrings.Encode()
	return authUrl, nil
}

// Get the MSISAuthCookie set by the auth portal after authentication.
func getMsisAuthCookie(client *http.Client, authUrl *url.URL, username, password string) (*http.Cookie, error) {
	formBody := url.Values{
		"UserName":   {username},
		"Password":   {password},
		"AuthMethod": {"FormsAuthentication"},
	}

	req, err := http.NewRequest(http.MethodPost, authUrl.String(), strings.NewReader(formBody.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cookie := getCookie(resp.Cookies(), msisAuthCookieName)
	if cookie == nil {
		return nil, errors.New("failed to get auth cookie")
	}

	log.Printf("Obtained %s cookie from %s\n", msisAuthCookieName, authUrl.Host)
	return cookie, nil
}

// Get the URL used for daily temperature declaration.
// The auth portal will redirect to this URL after authentication. It contains a unique ID.
func getHtdUrl(client *http.Client, username, password string) (*url.URL, error) {
	// Get the auth cookie from the auth portal
	authUrl, err := getVafsAuthUrl()
	if err != nil {
		return nil, err
	}
	authCookie, err := getMsisAuthCookie(client, authUrl, username, password)
	if err != nil {
		return nil, err
	}

	// Use the cookie to get htd url
	req, err := http.NewRequest(http.MethodGet, authUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(authCookie)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	htdUrl, err := resp.Location()
	if err == nil {
		log.Printf("Obtained unique temperature declaration URL %s\n", htdUrl.Hostname())
	}
	return htdUrl, err
}

// This cookie is used on the daily temperature declaration site.
func getJSessionId(client *http.Client, htdUrl *url.URL) (*http.Cookie, error) {
	req, err := http.NewRequest(http.MethodGet, htdUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sessionCookie := getCookie(resp.Cookies(), jSessionIdCookieName)
	if sessionCookie == nil {
		return nil, fmt.Errorf("found no cookie with name %s", jSessionIdCookieName)
	}
	log.Printf("Obtained %s cookie from %s\n", jSessionIdCookieName, htdUrl.Hostname())
	return sessionCookie, nil
}
