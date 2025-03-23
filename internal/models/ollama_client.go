package models

import (
	"fmt"
)

// OllamaClient provides a client for interacting with Ollama models
type OllamaClient struct {
	modelManager *ModelManager
}

// DownloadModelRequest represents a request to download a model
type DownloadModelRequest struct {
	Model   string
	Version string
}

// ModelFineTuningRequest represents a request to fine-tune a model
type ModelFineTuningRequest struct {
	ModelVersion string
	Dataset      string
}

// NewOllamaClient creates a new client for interacting with Ollama models
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		modelManager: NewModelManager("./models"),
	}
}

// DownloadModel downloads a model based on the provided request
func (c *OllamaClient) DownloadModel(req DownloadModelRequest) error {
	version := "latest"
	if req.Version != "" {
		version = req.Version
	}
	
	fmt.Printf("Downloading model %s (version %s)\n", req.Model, version)
	return c.modelManager.DownloadModel(req.Model, version)
}

// PreloadModels preloads multiple models for faster inference
func (c *OllamaClient) PreloadModels(models []string) {
	fmt.Printf("Preloading models: %v\n", models)
	c.modelManager.PreloadModels(models)
}

// FineTuneModel fine-tunes a model with a specific dataset
func (c *OllamaClient) FineTuneModel(req ModelFineTuningRequest) error {
	fmt.Printf("Fine-tuning model %s with dataset %s\n", req.ModelVersion, req.Dataset)
	return c.modelManager.FineTuneModel(req.ModelVersion, req.Dataset)
}
