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

type declarationData struct {
	hasSymptoms          string
	householdHasSymptoms string
}

func (d *declarationData) String() string {
	// 36.7  No   No
	return fmt.Sprintf("%-3s  %-3s", d.hasSymptoms, d.householdHasSymptoms)
}

type dailyDeclaration struct {
	date          string
	morningData   *declarationData
	afternoonData *declarationData
}

func (d *dailyDeclaration) String() string {
	// 01/03/2020, Wednesday  am...  pm...
	return fmt.Sprintf("%-21s  %s  %s", d.date, d.morningData, d.afternoonData)
}

const viewingUrl string = "https://myaces.nus.edu.sg/htd/htd?loadPage=viewtemperature"

func WriteDeclarations(writer io.Writer, username, password string) error {
	client := makeNoRedirectHttpClient()
	table, err := getTable(client, username, password)
	if err != nil {
		return err
	}
	declarations := getDeclarations(table)
	for _, decl := range declarations {
		fmt.Fprintln(writer, decl)
	}
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

// TODO: handle errors
func getDeclarations(table *goquery.Selection) []*dailyDeclaration {
	declarations := make([]*dailyDeclaration, 0)
	tbody := table.Find("tbody")
	tbody.Find("tr").Each(func(i int, row *goquery.Selection) {
		declarations = append(declarations, parseDailyDeclaration(row))
	})
	return declarations
}

func parseDailyDeclaration(row *goquery.Selection) *dailyDeclaration {
	cells := row.Find("td")
	date := parseDate(cells.Eq(1))
	morningData := parseDeclarationData(cells.Slice(2, 4))
	afternoonData := parseDeclarationData(cells.Slice(4, 6))
	return &dailyDeclaration{
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

func parseDeclarationData(cells *goquery.Selection) *declarationData {
	symptomsText := strings.TrimSpace(cells.Eq(1).Text())
	householdSymptomsText := strings.TrimSpace(cells.Eq(2).Text())

	return &declarationData{
		hasSymptoms:          symptomsText,
		householdHasSymptoms: householdSymptomsText,
	}
}
