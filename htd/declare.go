package htd

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Declare your temperature and symptoms.
func Declare(
	username string,
	password string,
	date time.Time,
	isMorning bool,
	hasSymptoms bool,
	familyHasSymptoms bool,
) error {
	client := makeNoRedirectHttpClient()
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

	familySymptomsFlag := "N"
	if familyHasSymptoms {
		familySymptomsFlag = "Y"
	}

	formattedDate := date.Format(dateFormat)
	formData := url.Values{
		"actionName":         {"dlytemperature"},
		"webdriverFlag":      {""},
		"tempDeclOn":         {formattedDate},
		"declFrequency":      {declFrequency},
		"symptomsFlag":       {symptomsFlag},
		"familySymptomsFlag": {familySymptomsFlag},
	}
	req, err := http.NewRequest(http.MethodPost, htdUrl.String(), strings.NewReader(formData.Encode()))
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
		dump, _ := httputil.DumpResponse(resp, false)
		log.Printf("Temperature declaration failed.\nReceived %q\n", dump)
		return errors.New("failed to submit temperature")
	}
	log.Printf("Successfully made a new declaration for %s %s.\nSymptoms: %s\nHousehold symptoms: %s\n",
		declFrequency, formattedDate, symptomsFlag, familySymptomsFlag)
	return nil
}
