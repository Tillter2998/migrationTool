package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tillter/goMigration/internal/helpers"
)

var cd = &helpers.CsvTable{
	Headers: map[string]int{
		"firstName": 0,
		"lastName":  1,
		"id":        2},
	Rows: []map[string]string{
		{
			"firstName": "Elrich",
			"lastName":  "De Villiers",
			"id":        "1",
		},
		{
			"firstName": "Haven",
			"lastName":  "Sloot Fralich",
			"id":        "2",
		},
		{
			"firstName": "Haven2",
			"lastName":  "Sloots Fralich",
			"id":        "3",
		},
	},
}

var sd = &helpers.CsvTable{
	Headers: map[string]int{
		"oldId": 0,
		"newId": 1,
	},
	Rows: []map[string]string{
		{
			"oldId": "1",
			"newId": "cus_1",
		},
		{
			"oldId": "2",
			"newId": "cus_2",
		},
	},
}

var ex = &helpers.CsvTable{
	Headers: map[string]int{
		"c_firstName": 0,
		"c_lastName":  1,
		"c_id":        2,
		"s_oldId":     3,
		"s_newId":     4,
	},
	Rows: []map[string]string{
		{
			"c_firstName": "Elrich",
			"c_lastName":  "De Villiers",
			"c_id":        "1",
			"s_oldId":     "1",
			"s_newId":     "cus_1",
		},
		{
			"c_firstName": "Haven",
			"c_lastName":  "Sloot Fralich",
			"c_id":        "2",
			"s_oldId":     "2",
			"s_newId":     "cus_2",
		},
	},
}

func TestMerge(t *testing.T) {
	baseDir, _ := os.Getwd()
	p := filepath.Join(baseDir, "data", "_TestMergeOutput.csv")
	os.Remove(p)

	loc, err := time.LoadLocation("America/Halifax")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	h := helpers.Helper{
		ClientFileName:          "clientData.csv",
		StripeFileName:          "stripeData.csv",
		MergeOutputFileName:     "_TestMergeOutput.csv",
		MigrationInputFileName:  "",
		MigrationOutputFileName: "",
		Location:                loc,
		BaseDir:                 baseDir,
		Stage:                   "",
	}

	h.Merge()

	assert.FileExists(t, p)
	os.Remove(filepath.Join(baseDir, "data", "_TestMergeOutput.csv"))
}
