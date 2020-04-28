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


// Declare your temperature and symptoms.
func Declare(
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

	formattedDate := date.Format(dateFormat)
	formData := url.Values{
		"actionName":    {"dlytemperature"},
		"tempDeclOn":    {formattedDate},
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
	log.Printf("Successfully made a new declaration for %s %s.\nTemperature: %.1f\nSymptoms: %s\n",
		declFrequency, formattedDate, temperature, symptomsFlag)
	return nil
}
