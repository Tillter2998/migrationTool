package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/tillter/goMigration/internal/helpers"
)

func main() {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Merge Files", "merge"),
					huh.NewOption("Migrate Subscriptions", "migrate"),
				).
				Value(&action),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	switch action {
	case "merge":
		var clientFileName string
		var stripeFileName string

		err = huh.NewInput().
			Title("Client Data Input File Name:").
			Value(&clientFileName).
			Run()
		if err != nil {
			log.Fatal(err)
		}

		err = huh.NewInput().
			Title("Stripe Data Input File Name:").
			Value(&stripeFileName).
			Run()
		err := spinner.New().
			Title("Merging Files").
			Action(func() { helpers.Merge(clientFileName, stripeFileName) }).
			Run()
		if err != nil {
			log.Fatal(err)
		}
	case "migrate":
		var timezone string
		err = huh.NewSelect[string]().
			Title("Select Migration Timezone").
			Options(
				huh.NewOption("Atlantic Time", "America/Halifax"),
				huh.NewOption("Eastern Time", "America/Toronto"),
			).
			Value(&timezone).
			Run()

		h.Location, err = time.LoadLocation(timezone)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = spinner.New().
			Title("Merging Files").
			Action(h.Migrate).
			Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
