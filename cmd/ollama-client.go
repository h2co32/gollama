package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/h2co32/gollama/internal/models"
	"github.com/h2co32/gollama/internal/utils"
)

func main() {
	model := flag.String("model", "default", "Specify the model to load")
	action := flag.String("action", "download", "Action to perform: download/preload/fine-tune")
	version := flag.Bool("version", false, "Display version information")

	flag.Parse()

	// Display version information if requested
	if *version {
		fmt.Printf("gollama version %s\n", utils.Version)
		os.Exit(0)
	}

	client := models.NewOllamaClient()

	switch *action {
	case "download":
		client.DownloadModel(models.DownloadModelRequest{Model: *model})
	case "preload":
		client.PreloadModels([]string{*model})
	case "fine-tune":
		client.FineTuneModel(models.ModelFineTuningRequest{Dataset: "custom-dataset", ModelVersion: *model})
	default:
		fmt.Println("Invalid action provided")
	}
}
