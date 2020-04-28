package htd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)


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
	logOutgoingRequest(req, "")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logResponse(resp, "Temperature declaration failed.\n")
		return errors.New("failed to submit temperature")
	}
	log.Println("Successful.")
	return nil
}
