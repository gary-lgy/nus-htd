package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gary-lgy/nus-htd/htd"
	"gopkg.in/alecthomas/kingpin.v2"
)

const am string = "am"
const pm string = "pm"

func main() {
	app := kingpin.New("nus-htd",
		"A command-line tool for making and viewing your daily health declarations at NUS.")
	username := app.Flag("username",
		"Your NUSNET ID. (default: $HTD_USERNAME.)").Envar("HTD_USERNAME").Short('u').String()
	password := app.Flag("password",
		"Your NUSNET password. (default: $HTD_PASSWORD)").Envar("HTD_PASSWORD").Short('p').String()

	declare := app.Command("declare", "Declare your health status.").Alias("d")
	morningOrAfternoon := declare.Arg("am or pm",
		"whether the declaration is for the morning or the afternoon").Required().Enum(am, pm)
	hasSymptoms := declare.Flag("has-symptoms",
		"Whether you have COVID-19 symptoms").Bool()
	familyHasSymptoms := declare.Flag("family-has-symptoms",
		"Whether anyone in your family has COVID-19 symptoms").Bool()
	declareAnomaly := declare.Flag("declare-anomaly",
		"Continue to declare even if you have a fever, or you or your family members have symptoms").Short('f').Bool()
	viewAfterDeclare := declare.Flag("view-after-declare",
		"View past declarations after making a new declaration. (default: true)").Default("true").Bool()

	view := app.Command("view", "View your past declarations.").Alias("v")

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	if username == nil || *username == "" {
		app.FatalUsage("Please supply a username or set the $HTD_USERNAME environment variable.")
	}
	if password == nil || *password == "" {
		app.FatalUsage("Please supply a password or set the $HTD_PASSWORD environment variable.")
	}

	switch command {
	case declare.FullCommand():
		makeDeclaration(*username, *password, *morningOrAfternoon, *hasSymptoms, *familyHasSymptoms, *declareAnomaly)
		if *viewAfterDeclare {
			printPastDeclarations(*username, *password)
		}
	case view.FullCommand():
		printPastDeclarations(*username, *password)
	}
}

func makeDeclaration(
	username, password, amOrPm string,
	hasSymptoms,
	familyHasSymptoms,
	declareAnomaly bool,
) {
	var isMorning bool
	if amOrPm == am {
		isMorning = true
	} else if amOrPm == pm {
		isMorning = false
	} else {
		printErrorMsgAndExit("Unexpected error")
	}

	if hasSymptoms && !declareAnomaly {
		printErrorMsgAndExit("Your have symptoms; not declaring. Pass -f to override.")
	}

	if familyHasSymptoms && !declareAnomaly {
		printErrorMsgAndExit("Someone in your family has symptoms; not declaring. Pass -f to override.")
	}

	err := htd.Declare(username, password, time.Now(), isMorning, hasSymptoms, familyHasSymptoms)
	exitIfError(err)
}

func printPastDeclarations(username, password string) {
	err := htd.WriteDeclarations(os.Stdout, username, password)
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
