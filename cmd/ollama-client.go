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
		if err := client.DownloadModel(models.DownloadModelRequest{Model: *model}); err != nil {
			fmt.Printf("Error downloading model: %v\n", err)
			os.Exit(1)
		}
	case "preload":
		client.PreloadModels([]string{*model})
	case "fine-tune":
		if err := client.FineTuneModel(models.ModelFineTuningRequest{Dataset: "custom-dataset", ModelVersion: *model}); err != nil {
			fmt.Printf("Error fine-tuning model: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Invalid action provided")
	}
}
