package htd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const msisAuthCookieName string = "MSISAuth"
const jSessionIdCookieName string = "JSESSIONID"
const dateFormat string = "02/01/2006" // Golang: why can't you just be normal?

func logOutgoingRequest(req *http.Request) {
	dump, _ := httputil.DumpRequestOut(req, true)
	log.Printf("Making request:\n%q\n", dump)
}

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
func getMsisAuthCookie(client *http.Client, authUrl, username, password string) (*http.Cookie, error) {
	formBody := url.Values{
		"UserName":   {username},
		"Password":   {password},
		"AuthMethod": {"FormsAuthentication"},
	}

	req, err := http.NewRequest(http.MethodPost, authUrl, strings.NewReader(formBody.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logOutgoingRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cookie := getCookie(resp.Cookies(), msisAuthCookieName)
	if cookie == nil {
		return nil, errors.New("failed to get auth cookie")
	}
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
	authCookie, err := getMsisAuthCookie(client, authUrl.String(), username, password)
	if err != nil {
		return nil, err
	}

	// Use the cookie to get htd url
	req, err := http.NewRequest(http.MethodGet, authUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(authCookie)
	logOutgoingRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Location()
}

// This cookie is used on the daily temperature declaration site.
func getJSessionId(client *http.Client, htdUrl *url.URL) (*http.Cookie, error) {
	req, err := http.NewRequest(http.MethodGet, htdUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	logOutgoingRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sessionCookie := getCookie(resp.Cookies(), jSessionIdCookieName)
	if sessionCookie == nil {
		return nil, fmt.Errorf("found no cookie with name %s", jSessionIdCookieName)
	}
	return sessionCookie, nil
}

// Report your temperature and symptoms.
func ReportTemperature(
	client *http.Client,
	username string,
	password string,
	date time.Time,
	isMorning bool,
	temperature float32,
	hasSymptoms bool,
) error {
	htdUrl, err := getHtdUrl(client, username, password)
	if err != nil {
		return err
	}
	sessionCookie, err := getJSessionId(client, htdUrl)
	if err != nil {
		return err
	}

	declFrequency := "A"
	if !isMorning {
		declFrequency = "P"
	}

	symptomsFlag := "N"
	if hasSymptoms {
		symptomsFlag = "Y" // I've no idea what the real flag is
	}

	formData := url.Values{
		"actionName":    {"dlytemperature"},
		"tempDeclOn":    {date.Format(dateFormat)},
		"declFrequency": {declFrequency},
		"temperature":   {fmt.Sprintf("%.1f", temperature)},
		"symptomsFlag":  {symptomsFlag},
	}
	req, err := http.NewRequest(http.MethodPost, htdUrl.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.AddCookie(sessionCookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	logOutgoingRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		dump, _ := httputil.DumpResponse(resp, true)
		log.Printf("Temperature submission failed.\nReceived %q", dump)
		return errors.New("failed to submit temperature")
	}
	log.Println("Successful.")
	return nil
}
