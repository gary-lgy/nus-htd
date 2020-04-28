package htd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type declaration struct {
	date string
	morningData string
	afternoonData string
}

const viewingUrl string = "https://myaces.nus.edu.sg/htd/htd?loadPage=viewtemperature"

func WriteDeclarations(writer io.Writer, client *http.Client, username, password string) error {
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
	logOutgoingRequest(req, "")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logResponse(resp, "Unexpected response from %s: \n")
		return nil, fmt.Errorf("failed to get past declarations")
	}

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
	date := strings.TrimSpace(cells.Eq(1).Text())
	morningData := parseCell(cells.Eq(2))
	afternoonData := parseCell(cells.Eq(3))
	return declaration{
		date:          date,
		morningData:   morningData,
		afternoonData: afternoonData,
	}
}

func parseCell(cell *goquery.Selection) string {
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
	for _, declaration := range declarations {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", declaration.date, declaration.morningData, declaration.afternoonData)
	}
}
