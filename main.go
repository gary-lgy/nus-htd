package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gary-lgy/nus-htd/htd"
	"gopkg.in/alecthomas/kingpin.v2"
)

const am string = "am"
const pm string = "pm"


func main() {
	app := kingpin.New("nus-htd",
		"A command-line tool for making and viewing your daily temperature declarations at NUS.")
	username := app.Flag("username",
		"Your NUSNET ID. (default: $HTD_USERNAME.)").Envar("HTD_USERNAME").Short('u').String()
	password := app.Flag("password",
		"Your NUSNET password. (default: $HTD_PASSWORD)").Envar("HTD_PASSWORD").Short('p').String()
	debug := app.Flag("debug",
		"print the received command line arguments and flag and immediately exit.").Bool()

	report := app.Command("report", "Report your temperature.")
	morningOrAfternoon := report.Arg("am or pm",
		"whether the declaration is for the morning or the afternoon").Required().Enum(am, pm)
	temperature := report.Arg("temperature",
		"Your temperature").Required().Float32()
	hasSymptoms := report.Flag("has-symptoms",
		"Whether you have cough, "+
			"a runny nose or sore throat that you have recently just acquired and is/are "+
			"not due to pre-existing conditions").Short('s').Bool()
	reportAnomaly := report.Flag("report-anomaly",
		"Continue to report even if your have a fever or cold symptoms.").Short('f').Bool()

	view := app.Command("view", "View your past declarations.")

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		debugPrint(command, username, password, morningOrAfternoon, temperature, hasSymptoms, reportAnomaly)
		os.Exit(2)
	}

	if username == nil || *username == "" {
		app.FatalUsage("Please supply a username or set the $HTD_USERNAME environment variable.")
	}
	if password == nil || *password == "" {
		app.FatalUsage("Please supply a password or set the $HTD_PASSWORD environment variable.")
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	switch command {
	case report.FullCommand():
		makeReport(client, *username, *password, *morningOrAfternoon, *temperature, *hasSymptoms, *reportAnomaly)
	case view.FullCommand():
		printPastDeclarations(client, *username, *password)
	}

}

func makeReport(
	client *http.Client,
	username, password, amOrPm string,
	temperature float32,
	hasSymptoms,
	reportAnomaly bool,
) {
	var isMorning bool
	if amOrPm == am {
		isMorning = true
	} else if amOrPm == pm {
		isMorning = false
	} else {
		printErrorMsgAndExit("Unexpected error")
	}

	if temperature < 35.0 {
		printErrorMsgAndExit("Temperature too low. Check your thermometer.")
	}
	if temperature >= 37.5 && !reportAnomaly {
		printErrorMsgAndExit("Your have a fever; not reporting. Pass -f to override.")
	}

	if hasSymptoms && !reportAnomaly {
		printErrorMsgAndExit("Your have symptoms; not reporting. Pass -f to override.")
	}

	err := htd.ReportTemperature(client, username, password, time.Now(), isMorning, temperature, hasSymptoms)
	exitIfError(err)
}

func printPastDeclarations(client *http.Client, username, password string) {
	err := htd.WriteDeclarations(os.Stdout, client, username, password)
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

