package helpers

import (
	"context"
	"encoding/csv"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CsvTable struct {
	Headers map[string]int
	Rows    []map[string]string
}

var (
	mergeOutputFileName   = "mergedData.csv"
	migrateOutputFileName = "migratedDate.csv"
)

func Merge(clientFileName string, stripeFileName string) {
	clientFile := openCsv(clientFileName)
	stripeFile := openCsv(stripeFileName)

	mergedFiles := mergeFiles(clientFile, stripeFile, "c_", "s_", "CustomerId", "old id")

	writeCsv(mergedFiles, mergeOutputFileName)
}

func Migrate(location *time.Location) {
	mergedFile := openCsv(mergeOutputFileName)

	migrationFile := migrateFile(mergedFile, location)

	writeCsv(migrationFile, migrateOutputFileName)
}

func migrateFile(mf *CsvTable, location *time.Location) *CsvTable {
	requiredHeaders := []string{"customer", "start_date", "price", "quantity", "automatic_tax", "billing_cycle_anchor", "coupon", "trial_end", "proration_behaviour", "collection_method", "cancel_at_period_end"}

	headers := make(map[string]int)
	for i, header := range requiredHeaders {
		headers[header] = i
	}
	rows := make([]map[string]string, 0, len(mf.Rows))
	for _, row := range parallel(mf.Rows) {
		rows = append(rows, processRow(row, location))
	}

	return &CsvTable{
		Headers: headers,
		Rows:    rows,
	}
}

func parallel[E any](events []E) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(events))

		for idx, row := range events {
			go func() {
				defer waitGroup.Done()

				select {
				case <-ctx.Done():
					return
				default:
					if !yield(idx, row) {
						cancel()
					}
				}
			}()
		}

		waitGroup.Wait()
	}
}

func processRow(row map[string]string, location *time.Location) map[string]string {
	fmt.Printf("Processing row for customer %v\n", row["c_CustomerId"])
	var startDateString string
	startDateString = strconv.FormatInt(time.Now().In(location).Add(time.Hour).UTC().Unix(), 10)

	billingCycleAnchorDate := getBillingCycleAnchorDate(row["c_StartDateISO"], row["c_BillingInterval"], row["c_NextBillingDateISO"], location)

	var quantity string
	if val, ok := row["c_Quantity"]; ok {
		quantity = val
	}

	r := map[string]string{
		"customer":             row["s_new id"],
		"start_date":           startDateString,
		"price":                "",
		"quantity":             quantity,
		"automatic_tax":        "false",
		"billing_cycle_anchor": billingCycleAnchorDate,
		"coupon":               "",
		"trial_end":            "",
		"proration_behaviour":  "",
		"collection_method":    "charge automatically",
		"cancel_at_period_end": "false",
	}
	return r
}

func getBillingCycleAnchorDate(startDateString, billingCycleInterval, nextBillDateString string, location *time.Location) string {
	months := 0
	years := 0
	if strings.ToLower(billingCycleInterval) == "yearly" {
		years++
	} else if strings.ToLower(billingCycleInterval) == "monthly" {
		months++
	}

	now := time.Now().In(location)
	//sd, err := time.ParseInLocation("YYYY/MM/DD", sds, loc)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	nextBillDate, err := time.ParseInLocation("2006/01/02", nextBillDateString, location)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if now.After(nextBillDate) {
		nextBillDate = nextBillDate.AddDate(years, months, 0)
	}

	if now.AddDate(years, months, 0).Before(nextBillDate) {
		nextBillDate = now.AddDate(years, months, 0)
	}

	updatedNextBillDate := strconv.FormatInt(time.Date(nextBillDate.Year(), nextBillDate.Month(), nextBillDate.Day(), 0, 0, 0, 0, location).UTC().Unix(), 10)
	return updatedNextBillDate
}

func csvFileToSlice(c *CsvTable) [][]string {
	csvSlice := make([][]string, len(c.Rows)+1)

	headers := make([]string, 0, len(c.Headers))
	for header := range c.Headers {
		headers = append(headers, header)
	}
	csvSlice[0] = headers

	for i, row := range c.Rows {
		csvSlice[i+1] = make([]string, len(headers))
		for j, header := range headers {
			csvSlice[i+1][j] = row[header]
		}
	}

	return csvSlice
}

func writeCsv(c *CsvTable, outputFileName string) {
	path := filepath.Join("../../data", outputFileName)
	fileWriter, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cw := csv.NewWriter(fileWriter)
	defer cw.Flush()

	cs := csvFileToSlice(c)
	for _, row := range cs {
		err = cw.Write(row)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func newCsvTable(headers []string, rows [][]string) *CsvTable {
	mergedHeaders := make(map[string]int)
	for i, header := range headers {
		mergedHeaders[header] = i
	}

	tableRows := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		rowMap := make(map[string]string)
		for key, index := range mergedHeaders {
			rowMap[key] = row[index]
		}
		tableRows = append(tableRows, rowMap)
	}
	return &CsvTable{
		Headers: mergedHeaders,
		Rows:    tableRows,
	}
}

func openCsv(fileName string) *CsvTable {
	filePath := filepath.Join("../../data", fileName)
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	allRows, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return newCsvTable(headers, allRows)
}

func mergeFiles(clientData *CsvTable, stripeData *CsvTable, clientPrepend string, stripePrepend string, clientId string, stripeId string) *CsvTable {
	mergedHeaders := make(map[string]int)
	for header, i := range clientData.Headers {
		mergedHeaders[clientPrepend+header] = i
	}
	for header := range stripeData.Headers {
		mergedHeaders[stripePrepend+header] = len(mergedHeaders)
	}

	sdMap := make(map[string]map[string]string)
	for _, row := range stripeData.Rows {
		sdMap[row[stripeId]] = row
	}

	mergedRows := make([]map[string]string, 0)

	for _, cdRow := range clientData.Rows {
		if sdRow, ok := sdMap[cdRow[clientId]]; ok {
			newRow := make(map[string]string)
			for header, value := range cdRow {
				newRow[clientPrepend+header] = value
			}
			for header, value := range sdRow {
				newRow[stripePrepend+header] = value
			}
			mergedRows = append(mergedRows, newRow)
		}
	}

	return &CsvTable{
		Headers: mergedHeaders,
		Rows:    mergedRows,
	}
}
