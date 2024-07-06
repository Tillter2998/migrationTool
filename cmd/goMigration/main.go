package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tillter/goMigration/internal/helpers"
)

func main() {
	baseDir, err := os.Getwd() // This gets the current working directory
	if err != nil {
		panic(err)
	}

	merge := flag.Bool("merge", false, "Merge client and stripe data")
	migrate := flag.Bool("migrate", false, "Migrate data")
	postMigration := flag.Bool("post-migrate", false, "Migrate data")

	clientFile := flag.String("clientFile", "clientData.csv", "Client file to read from")
	stripeFile := flag.String("stripeFile", "stripeData.csv", "Stripe file to read from")
	outputFile := flag.String("outputFile", "mergedData.csv", "Output file to write to")
	migrationFile := flag.String("migrationFile", "mergedData.csv", "Migration file to read from")
	migrationOutputFile := flag.String("MigrateOutputFile", "migrationData.csv", "Migration Output file to write to")
	tz := flag.String("timezone", "America/Halifax", "Time zone")
	flag.Parse()

	loc, err := time.LoadLocation(*tz)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	h := helpers.Helper{
		ClientFileName:          *clientFile,
		StripeFileName:          *stripeFile,
		MergeOutputFileName:     *outputFile,
		MigrationInputFileName:  *migrationFile,
		MigrationOutputFileName: *migrationOutputFile,
		Location:                loc,
		BaseDir:                 baseDir,
	}
	if *merge {
		h.Merge()
	}

	if *migrate {
		h.Migrate()
	}

	if *postMigration {
	}
}
