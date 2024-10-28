package main

import (
	"flag"
	"fmt"
	"gollama/internal/models"
)

func main() {
	model := flag.String("model", "default", "Specify the model to load")
	action := flag.String("action", "download", "Action to perform: download/preload/fine-tune")

	flag.Parse()

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
