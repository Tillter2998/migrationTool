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

type Helper struct {
	ClientFileName          string
	StripeFileName          string
	MergeOutputFileName     string
	MigrationInputFileName  string
	MigrationOutputFileName string
	Location                *time.Location
	BaseDir                 string
	Stage                   string
}

func (h Helper) Merge() {
	cd := openCsv(h.BaseDir, h.ClientFileName)
	sd := openCsv(h.BaseDir, h.StripeFileName)

	mf := mergeFiles(cd, sd, "c_", "s_", "CustomerId", "old id")

	writeCsv(h.BaseDir, mf, h.MergeOutputFileName)
}

func (h Helper) Migrate() {
	mf := openCsv(h.BaseDir, h.MigrationInputFileName)

	mif := h.migrateFile(mf)

	writeCsv(h.BaseDir, mif, h.MigrationOutputFileName)
}

func (h Helper) migrateFile(mf *CsvTable) *CsvTable {
	hs := []string{"customer", "start_date", "price", "quantity", "automatic_tax", "billing_cycle_anchor", "coupon", "trial_end", "proration_behaviour", "collection_method", "cancel_at_period_end"}

	he := make(map[string]int)
	for i, header := range hs {
		he[header] = i
	}
	rs := make([]map[string]string, 0, len(mf.Rows))
	for _, row := range parallel(mf.Rows) {
		rs = append(rs, h.processRow(row))
	}

	return &CsvTable{
		Headers: he,
		Rows:    rs,
	}
}

func parallel[E any](events []E) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(len(events))

		for idx, row := range events {
			go func() {
				defer wg.Done()

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

		wg.Wait()
	}
}

func (h Helper) processRow(row map[string]string) map[string]string {
	fmt.Printf("Processing row for customer %v\n", row["c_CustomerId"])
	var sds string
	if h.Stage == "production" {
		sds = strconv.FormatInt(time.Now().In(h.Location).Add(time.Hour).UTC().Unix(), 10)
	} else if h.Stage == "development" {
		sds = strconv.FormatInt(time.Now().In(h.Location).Add(time.Hour+time.Minute*30).UTC().Unix(), 10)
	}

	bsad := getBillingCycleAnchorDate(row["c_StartDateISO"], row["c_BillingInterval"], row["c_NextBillingDateISO"], h.Location)

	var quan string
	if val, ok := row["c_Quantity"]; ok {
		quan = val
	}

	r := map[string]string{
		"customer":             row["s_new id"],
		"start_date":           sds,
		"price":                "",
		"quantity":             quan,
		"automatic_tax":        "false",
		"billing_cycle_anchor": bsad,
		"coupon":               "",
		"trial_end":            "",
		"proration_behaviour":  "",
		"collection_method":    "charge automatically",
		"cancel_at_period_end": "false",
	}
	return r
}

func getBillingCycleAnchorDate(sds, bci, nbds string, loc *time.Location) string {
	mths := 0
	yrs := 0
	if strings.ToLower(bci) == "yearly" {
		yrs++
	} else if strings.ToLower(bci) == "monthly" {
		mths++
	}

	now := time.Now().In(loc)
	//sd, err := time.ParseInLocation("YYYY/MM/DD", sds, loc)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	nbd, err := time.ParseInLocation("2006/01/02", nbds, loc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if now.After(nbd) {
		nbd = nbd.AddDate(yrs, mths, 0)
	}

	if now.AddDate(yrs, mths, 0).Before(nbd) {
		nbd = now.AddDate(yrs, mths, 0)
	}

	nbdU := strconv.FormatInt(time.Date(nbd.Year(), nbd.Month(), nbd.Day(), 0, 0, 0, 0, loc).UTC().Unix(), 10)
	return nbdU
}

func csvFileToSlice(c *CsvTable) [][]string {
	s := make([][]string, len(c.Rows)+1)

	headers := make([]string, 0, len(c.Headers))
	for header := range c.Headers {
		headers = append(headers, header)
	}
	s[0] = headers

	for i, row := range c.Rows {
		s[i+1] = make([]string, len(headers))
		for j, header := range headers {
			s[i+1][j] = row[header]
		}
	}

	return s
}

func writeCsv(baseDir string, c *CsvTable, ofn string) {
	p := filepath.Join(baseDir, "data", ofn)
	fw, err := os.Create(p)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cw := csv.NewWriter(fw)
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

func newCsvTable(h []string, rows [][]string) *CsvTable {
	headerMap := make(map[string]int)
	for i, header := range h {
		headerMap[header] = i
	}

	tableRows := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		rowMap := make(map[string]string)
		for key, index := range headerMap {
			rowMap[key] = row[index]
		}
		tableRows = append(tableRows, rowMap)
	}
	return &CsvTable{
		Headers: headerMap,
		Rows:    tableRows,
	}
}

func openCsv(bd, f string) *CsvTable {
	f = filepath.Join(bd, "data", f)
	file, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	r := csv.NewReader(file)
	headers, err := r.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	allRows, err := r.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return newCsvTable(headers, allRows)
}

func mergeFiles(cd *CsvTable, sd *CsvTable, cdPre string, sdPre string, cdId string, sdId string) *CsvTable {
	headerMap := make(map[string]int)
	for header, i := range cd.Headers {
		headerMap[cdPre+header] = i
	}
	for header := range sd.Headers {
		headerMap[sdPre+header] = len(headerMap)
	}

	sdMap := make(map[string]map[string]string)
	for _, row := range sd.Rows {
		sdMap[row[sdId]] = row
	}

	mr := make([]map[string]string, 0)

	for _, cdRow := range cd.Rows {
		if sdRow, ok := sdMap[cdRow[cdId]]; ok {
			nr := make(map[string]string)
			for header, value := range cdRow {
				nr[cdPre+header] = value
			}
			for header, value := range sdRow {
				nr[sdPre+header] = value
			}
			mr = append(mr, nr)
		}
	}

	return &CsvTable{
		Headers: headerMap,
		Rows:    mr,
	}
}
