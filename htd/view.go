package htd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type declaration struct {
	date          string
	morningData   string
	afternoonData string
}

const viewingUrl string = "https://myaces.nus.edu.sg/htd/htd?loadPage=viewtemperature"

func WriteDeclarations(writer io.Writer, username, password string) error {
	client := makeNoRedirectHttpClient()
	table, err := getTable(client, username, password)
	if err != nil {
		return err
	}
	declarations := getDeclarations(table)
	printDeclarations(writer, declarations)
	return nil
}

func getTable(client *http.Client, username, password string) (*goquery.Selection, error) {
	htdUrl, err := getHtdUrl(client, username, password)
	if err != nil {
		return nil, err
	}
	sessionCookie, err := getJSessionId(client, htdUrl)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, viewingUrl, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(sessionCookie)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dump, _ := httputil.DumpResponse(resp, false)
		log.Printf("Failed to get past declarations.\nReceived unexpected response: %q\n", dump)
		return nil, fmt.Errorf("failed to get past declarations")
	}
	log.Println("Successfully retrieved past declarations.")

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse the received html")
	}

	return doc.Find("#myTable").First(), nil
}

func getDeclarations(table *goquery.Selection) []declaration {
	declarations := make([]declaration, 0)
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		declarations = append(declarations, parseRow(row))
	})
	return declarations
}

func parseRow(row *goquery.Selection) declaration {
	cells := row.Find("td")
	date := parseDate(cells.Eq(1))
	morningData := parseData(cells.Eq(2))
	afternoonData := parseData(cells.Eq(3))
	return declaration{
		date:          date,
		morningData:   morningData,
		afternoonData: afternoonData,
	}
}

func parseDate(cell *goquery.Selection) string {
	raw := cell.Text()
	components := strings.SplitN(raw, ",", 2)
	if len(components) < 2 {
		return raw
	}
	date := strings.TrimSpace(components[0])
	dayOfWeek := strings.TrimSpace(components[1])
	return date + ", " + dayOfWeek
}

func parseData(cell *goquery.Selection) string {
	rawData := cell.Text()
	components := strings.SplitN(rawData, ",", 2)
	if len(components) <= 1 {
		return ""
	}
	temperature := strings.TrimSpace(components[0])
	symptoms := strings.TrimSpace(components[1])
	return temperature + " " + symptoms
}

func printDeclarations(writer io.Writer, declarations []declaration) {
	maxDateLength := 22 // 01/04/2020 , Wednesday
	desiredSpacesAfterDate := 4
	maxAmLength := 7 // 35.9 No
	desiredSpacesAfterAm := 4
	for _, declaration := range declarations {
		spacesAfterDate := strings.Repeat(" ", maxDateLength+desiredSpacesAfterDate-len(declaration.date))
		spacesAfterAm := strings.Repeat(" ", maxAmLength+desiredSpacesAfterAm-len(declaration.morningData))
		fmt.Fprintf(writer, "%s%s%s%s%s\n", declaration.date, spacesAfterDate, declaration.morningData, spacesAfterAm, declaration.afternoonData)

	}
}
