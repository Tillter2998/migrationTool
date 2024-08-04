package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/tillter/goMigration/internal/helpers"
)

func main() {
	var action string
	var provider string
	var clientFile string
	var stripeFile string
	var timezone string
	var loc *time.Location

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

	var merge bool
	var migrate bool

	switch action {
	case "merge":
		err = huh.NewInput().
			Title("Client Data Input File Name:").
			Value(&clientFile).
			Run()
		if err != nil {
			log.Fatal(err)
		}

		err = huh.NewInput().
			Title("Stripe Data Input File Name:").
			Value(&stripeFile).
			Run()
		merge = true
	case "migrate":
		err = huh.NewSelect[string]().
			Title("Select Migration Timezone").
			Options(
				huh.NewOption("Atlantic Time", "America/Halifax"),
				huh.NewOption("Eastern Time", "America/Toronto"),
			).
			Value(&timezone).
			Run()

		loc, err = time.LoadLocation(timezone)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		migrate = true
	}

	h := helpers.Helper{
		ClientFileName: clientFile,
		StripeFileName: stripeFile,
		Location:       loc,
		Provider:       provider,
		Action:         action,
	}

	if merge {
		h.Merge()
	}

	if migrate {
		h.Migrate()
	}
}
