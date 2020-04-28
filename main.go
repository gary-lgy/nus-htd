package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

type FeverError struct {
	temperature float32
}

func (err *FeverError) Error() string {
	return fmt.Sprintf("You have a fever of %.1f.", err.temperature)
}

const MsisAuthCookieName string = "MSISAuth"
const JSessionIdCookieName string = "JSESSIONID"
const DateFormat string = "02/01/2006"

func main() {
	const am string = "am"
	const pm string = "pm"
	app := kingpin.New("nus-htd", "A command-line tool to make your daily temperature declaration to NUS.")
	username := app.Flag("username", "Your NUSNET ID. (default: $HTD_USERNAME.)").Envar("HTD_USERNAME").Short('u').String()
	password := app.Flag("password", "Your NUSNET password. (default: $HTD_PASSWORD)").Envar("HTD_PASSWORD").Short('p').String()
	morningOrAfternoon := app.Arg("am or pm",
		"whether the declaration is for the morning or the afternoon").Required().Enum(am, pm)
	temperature := app.Arg("temperature", "Your temperature").Required().Float32()
	reportFever := app.Flag("report-fever", "Report the temperature even if your have a fever").Short('f').Bool()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if username == nil || *username == "" {
		app.FatalUsage("Please suppler a username or set the $HTD_USERNAME environment variable.")
	}
	if password == nil || *password == "" {
		app.FatalUsage("Please suppler a password or set the $HTD_PASSWORD environment variable.")
	}

	isMorning := true
	if *morningOrAfternoon != am {
		isMorning = false
	}

	// TODO :Separate into different files, including FeverError
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	htdUrl, err := getHtdUrl(client, *username, *password)
	exitIfError(err)
	err = reportTemperature(client, htdUrl, *temperature, isMorning, *reportFever)
	exitIfError(err)
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func getCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func getVafsAuthUrl() (string, error) {
	authUrl, err := url.Parse("https://vafs.nus.edu.sg/adfs/oauth2/authorize")
	if err != nil {
		return "", err
	}
	queryStrings := url.Values{
		"response_type": {"code"},
		"client_id":     {"97F0D1CACA7D41DE87538F9362924CCB-184318"},
		"resource":      {"sg_edu_nus_oauth"},
		"redirect_uri":  {"https://myaces.nus.edu.sg:443/htd/htd"},
	}
	authUrl.RawQuery = queryStrings.Encode()
	return authUrl.String(), nil
}

func getMsisAuthCookie(client *http.Client, authUrl, username, password string) (*http.Cookie, error) {
	formBody := url.Values{
		"UserName":   {username},
		"Password":   {password},
		"AuthMethod": {"FormsAuthentication"},
	}
	authUrl, err := getVafsAuthUrl()
	if err != nil {
		return nil, err
	}
	resp, err := client.PostForm(authUrl, formBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cookie := getCookie(resp.Cookies(), MsisAuthCookieName)
	if cookie == nil {
		return nil, errors.New("failed to get auth cookie")
	}
	return cookie, nil
}

func getHtdUrl(client *http.Client, username, password string) (*url.URL, error) {
	// Get the auth cookie from vafs.nus.edu.sg
	authUrl, err := getVafsAuthUrl()
	if err != nil {
		return nil, err
	}
	authCookie, err := getMsisAuthCookie(client, authUrl, username, password)

	// Use the cookie to get htd url
	req, err := http.NewRequest(http.MethodGet, authUrl, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(authCookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Location()
}

func getJSessionId(client *http.Client, htdUrl *url.URL) (*http.Cookie, error) {
	resp, err := client.Get(htdUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sessionCookie := getCookie(resp.Cookies(), JSessionIdCookieName)
	if sessionCookie == nil {
		return nil, fmt.Errorf("found no cookie with name %s", JSessionIdCookieName)
	}
	return sessionCookie, nil
}

func reportTemperature(
	client *http.Client,
	htdUrl *url.URL,
	temperature float32,
	isMorning bool,
	reportFever bool,
) error {
	if temperature < 35.0 {
		return errors.New("temperature too low; check your thermometer")
	}
	if temperature >= 37.5 && !reportFever {
		return &FeverError{temperature: temperature}
	}

	sessionCookie, err := getJSessionId(client, htdUrl)
	if err != nil {
		return err
	}

	date := time.Now().Format(DateFormat)
	declFrequency := "A"
	if !isMorning {
		declFrequency = "P"
	}
	formData := url.Values{
		"actionName":    {"dlytemperature"},
		"tempDeclOn":    {date},
		"declFrequency": {declFrequency},
		"temperature":   {fmt.Sprintf("%.1f", temperature)},
		"symptomsFlag":  {"N"},
	}
	req, err := http.NewRequest(http.MethodPost, htdUrl.String(),
		strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.AddCookie(sessionCookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		dump, _ := httputil.DumpResponse(resp, true)
		_, _ = fmt.Fprintf(os.Stderr, "Temperature submission failed.\nReceived %q", dump)
		return errors.New("failed to submit temperature")
	}
	return nil
}
