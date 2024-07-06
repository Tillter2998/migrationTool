package main

import (
	"flag"
	"fmt"
	"github.com/tillter/goMigration/internal/helpers"
	"os"
	"time"
)

func main() {
	baseDir, err := os.Getwd() // This gets the current working directory
	if err != nil {
		panic(err)
	}

	m := flag.Bool("merge", false, "Merge client and stripe data")
	mig := flag.Bool("migrate", false, "Migrate data")
	pMig := flag.Bool("post-migrate", false, "Migrate data")

	cfn := flag.String("clientFile", "clientData.csv", "Client file to read from")
	sfn := flag.String("stripeFile", "stripeData.csv", "Stripe file to read from")
	ofn := flag.String("outputFile", "mergedData.csv", "Output file to write to")
	mfn := flag.String("migrationFile", "mergedData.csv", "Migration file to read from")
	mofn := flag.String("MigrateOutputFile", "migrationData.csv", "Migration Output file to write to")
	tz := flag.String("timezone", "America/Halifax", "Time zone")
	flag.Parse()

	loc, err := time.LoadLocation(*tz)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	h := helpers.Helper{
		ClientFileName:          *cfn,
		StripeFileName:          *sfn,
		MergeOutputFileName:     *ofn,
		MigrationInputFileName:  *mfn,
		MigrationOutputFileName: *mofn,
		Location:                loc,
		BaseDir:                 baseDir,
	}
	if *m {
		h.Merge()
	}

	if *mig {
		h.Migrate()
	}

	if *pMig {
	}
}
