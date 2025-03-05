package main

import (
	"os"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	soda "github.com/SebastiaanKlippert/go-soda"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

const buildingPermitEndpoint string = "https://data.seattle.gov/resource/76t5-zqzr"

type buildingPermit struct {
	PermitNum         string  `json:"permitnum"`
	PermitClass       string  `json:"permitclass"`
	PermitClassMapped string  `json:"permitclassmapped"`
	PermitTypeMapped  string  `json:"permittypemapped"`
	PermitTypeDesc    string  `json:"permittypedesc"`
	Description       string  `json:"description"`
	AppliedDate       string  `json:"applieddate"`
	IssuedDate        string  `json:"issueddate"`
	ExpiresDate       string  `json:"expiresdate"`
	Status            string  `json:"statuscurrent"`
	RelatedMup        string  `json:"relatedmup"`
	Address           string  `json:"originaladdress1"`
	//Link              string  `json:"link"`
	Latitude          string `json:"latitude"`
	Longitude         string `json:"longitude"`
}

type coordinates struct {
	Lat  float32
	Long float32
}

type byAddress []buildingPermit
func (r byAddress) Len() int	{ return len(r) }
func (r byAddress) Less(i, j int) bool { return r[i].Address < r[j].Address }
func (r byAddress) Swap(i, j int) { r[i], r[j] = r[j], r[i] }


func main() {
	appToken := os.Getenv("APP_TOKEN")
	var nwCorner = coordinates{Lat: 47.712288, Long: -122.327557}
	var seCorner = coordinates{Lat: 47.704218, Long: -122.301912}
	var limit uint = 1000
	queryStatusExclusions := [...]string{
		"Completed",
		"Withdrawn",
		"Expired",
		"Canceled",
		"Closed"}
	var sb strings.Builder
	for i, str := range queryStatusExclusions {
		sb.WriteString("'")
		sb.WriteString(str)
		sb.WriteString("'")
		if i<len(queryStatusExclusions)-1{
			sb.WriteString(", ")
		}
	}
	statusExclusions := sb.String()

	query := fmt.Sprintf("statuscurrent not in (%s) AND within_box(location1, %.6f, %.6f, %.6f, %.6f)", statusExclusions, nwCorner.Lat, nwCorner.Long, seCorner.Lat, seCorner.Long)
	buildingPermits, recordCount := GetSeaOpendataRecords(buildingPermitEndpoint, appToken, query, limit)
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
  	columnFmt := color.New(color.FgYellow).SprintfFunc()
	
	permitTable := table.New("PermitNum", "PermitClass", "PermitTypeDesc", "Description", "AppliedDate", "IssuedDate", "ExpiresDate", "Status", "RelatedMUP", "Address", "Description")
	permitTable.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, record := range buildingPermits {
		permitTable.AddRow(record.PermitNum, record.PermitClass, record.PermitTypeDesc, record.AppliedDate, record.IssuedDate, record.ExpiresDate, record.Status, record.RelatedMup, record.Address, record.Description)
	}
	permitTable.Print()
	fmt.Printf("Available Records: %d", recordCount)
}

func GetSeaOpendataRecords(sodaEndpoint string, appToken string, sodaQuery string, limit uint) ([]buildingPermit, uint) {
	var records []buildingPermit
	sodareq := soda.NewGetRequest(sodaEndpoint, appToken)
	sodareq.Format = "json"
	sodareq.Query.Where = sodaQuery
	sodareq.Query.Limit = limit
	count, err := sodareq.Count()
	if err != nil {
		log.Fatal(err)
	}
	resp, err := sodareq.Get()
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		log.Fatal(err)
	}
	sort.Sort(byAddress(records))
	return records, count
}
