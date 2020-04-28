package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gary-lgy/nus-htd/htd"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	const am string = "am"
	const pm string = "pm"

	app := kingpin.New("nus-htd",
		"A command-line tool to make your daily temperature declaration to NUS.")
	username := app.Flag("username",
		"Your NUSNET ID. (default: $HTD_USERNAME.)").Envar("HTD_USERNAME").Short('u').String()
	password := app.Flag("password",
		"Your NUSNET password. (default: $HTD_PASSWORD)").Envar("HTD_PASSWORD").Short('p').String()
	morningOrAfternoon := app.Arg("am or pm",
		"whether the declaration is for the morning or the afternoon").Required().Enum(am, pm)
	temperature := app.Arg("temperature",
		"Your temperature").Required().Float32()
	hasSymptoms := app.Flag("has-symptoms",
		"Whether you have cough, " +
		"a runny nose or sore throat that you have recently just acquired and is/are " +
		"not due to pre-existing conditions").Short('s').Bool()
	reportAnomaly := app.Flag("report-anomaly",
		"Continue to report even if your have a fever or cold symptoms.").Short('f').Bool()
	debug := app.Flag("debug",
		"print the received command line arguments and flag and immediately exit.").Bool()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		debugPrint(username, password, morningOrAfternoon, temperature, hasSymptoms, reportAnomaly)
		os.Exit(2)
	}

	if username == nil || *username == "" {
		app.FatalUsage("Please supply a username or set the $HTD_USERNAME environment variable.")
	}
	if password == nil || *password == "" {
		app.FatalUsage("Please supply a password or set the $HTD_PASSWORD environment variable.")
	}

	isMorning := true
	if *morningOrAfternoon != am {
		isMorning = false
	}

	if *temperature < 35.0 {
		printErrorMsgAndExit("Temperature too low. Check your thermometer.")
	}
	if *temperature >= 37.5 && !*reportAnomaly {
		printErrorMsgAndExit("Your have a fever; not reporting. Pass -f to override.")
	}

	if *hasSymptoms && !*reportAnomaly {
		printErrorMsgAndExit("Your have symptoms; not reporting. Pass -f to override.")
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	err := htd.ReportTemperature(client, *username, *password, time.Now(), isMorning, *temperature, *hasSymptoms)
	exitIfError(err)
}

func exitIfError(err error) {
	if err != nil {
		printErrorMsgAndExit(err.Error())
	}
}

func printErrorMsgAndExit(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func debugPrint(username, password, morningOrAfternoon *string, temperature *float32, hasSymptoms, reportAnomaly *bool) {
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Password: %s\n", *password)
	fmt.Printf("morningOrAfternoon: %s\n", *morningOrAfternoon)
	fmt.Printf("temperature: %.1f\n", *temperature)
	fmt.Printf("hasSymptoms: %v\n", *hasSymptoms)
	fmt.Printf("reportAnomaly: %v\n", *reportAnomaly)
}
