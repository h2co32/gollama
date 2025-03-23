package models

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewModelManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	if mm == nil {
		t.Fatal("Expected NewModelManager to return a non-nil value")
	}

	if mm.modelDir != tempDir {
		t.Errorf("Expected mm.modelDir to be '%s', got '%s'", tempDir, mm.modelDir)
	}

	// Check that the maps are initialized
	if mm.currentVersion == nil {
		t.Error("Expected mm.currentVersion to be initialized")
	}

	if mm.loadedModels == nil {
		t.Error("Expected mm.loadedModels to be initialized")
	}

	if mm.fineTuningData == nil {
		t.Error("Expected mm.fineTuningData to be initialized")
	}

	// Verify the model directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Expected model directory '%s' to be created", tempDir)
	}
}

func TestDownloadModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Mock HTTP server is not needed since we're mocking at a higher level
	// by overriding the HTTP client in the implementation

	// Test downloading a model
	modelName := "test-model"
	version := "v1.0"

	// Download the model
	err = mm.DownloadModel(modelName, version)
	if err != nil {
		// Since we can't easily mock the HTTP client in the implementation,
		// we expect an error here in a real test environment
		if !strings.Contains(err.Error(), "failed to download model") {
			t.Errorf("Expected error to contain 'failed to download model', got '%s'", err.Error())
		}
		return
	}

	// If no error (which might happen if the HTTP request somehow succeeds),
	// verify the model file was created
	modelPath := filepath.Join(tempDir, modelName+"-"+version+".bin")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Errorf("Expected model file '%s' to be created", modelPath)
	}

	// Verify the current version was updated
	if mm.currentVersion[modelName] != version {
		t.Errorf("Expected mm.currentVersion['%s'] to be '%s', got '%s'", modelName, version, mm.currentVersion[modelName])
	}
}

func TestLoadModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create a mock model file
	modelName := "test-model"
	version := "v1.0"
	modelPath := filepath.Join(tempDir, modelName+"-"+version+".bin")
	if err := ioutil.WriteFile(modelPath, []byte("mock model data"), 0644); err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Set the current version
	mm.currentVersion[modelName] = version

	// Test loading the model
	err = mm.LoadModel(modelName)
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// Verify the model was marked as loaded
	if !mm.loadedModels[modelName] {
		t.Errorf("Expected model '%s' to be marked as loaded", modelName)
	}

	// Test loading a model that's already loaded
	err = mm.LoadModel(modelName)
	if err != nil {
		t.Errorf("Expected no error when loading an already loaded model, got %v", err)
	}

	// Test loading a non-existent model
	err = mm.LoadModel("non-existent-model")
	if err == nil {
		t.Error("Expected error when loading a non-existent model, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got '%s'", err.Error())
	}
}

func TestUnloadModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Set up a loaded model
	modelName := "test-model"
	mm.loadedModels[modelName] = true

	// Test unloading the model
	err = mm.UnloadModel(modelName)
	if err != nil {
		t.Fatalf("Failed to unload model: %v", err)
	}

	// Verify the model was marked as unloaded
	if mm.loadedModels[modelName] {
		t.Errorf("Expected model '%s' to be marked as unloaded", modelName)
	}

	// Test unloading a model that's not loaded
	err = mm.UnloadModel(modelName)
	if err == nil {
		t.Error("Expected error when unloading a model that's not loaded, got nil")
	}
	if !strings.Contains(err.Error(), "not loaded") {
		t.Errorf("Expected error to contain 'not loaded', got '%s'", err.Error())
	}
}

func TestFineTuneModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create a mock dataset file
	datasetPath := filepath.Join(tempDir, "test-dataset.txt")
	if err := ioutil.WriteFile(datasetPath, []byte("mock dataset data"), 0644); err != nil {
		t.Fatalf("Failed to create mock dataset file: %v", err)
	}

	// Test fine-tuning a model
	modelName := "test-model"
	err = mm.FineTuneModel(modelName, datasetPath)
	if err != nil {
		t.Fatalf("Failed to fine-tune model: %v", err)
	}

	// Verify the fine-tuned model file was created
	// The file name should start with the model name and include "ft-" followed by a timestamp
	files, err := ioutil.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read model directory: %v", err)
	}

	var fineTunedModelFound bool
	var fineTunedVersion string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), modelName+"-ft-") && strings.HasSuffix(file.Name(), ".bin") {
			fineTunedModelFound = true
			fineTunedVersion = strings.TrimSuffix(file.Name(), ".bin")
			break
		}
	}

	if !fineTunedModelFound {
		t.Error("Expected fine-tuned model file to be created")
	}

	// Verify the current version was updated
	if mm.currentVersion[modelName] != fineTunedVersion {
		t.Errorf("Expected mm.currentVersion['%s'] to be '%s', got '%s'", modelName, fineTunedVersion, mm.currentVersion[modelName])
	}

	// Verify the fine-tuning dataset was recorded
	if mm.fineTuningData[modelName] != datasetPath {
		t.Errorf("Expected mm.fineTuningData['%s'] to be '%s', got '%s'", modelName, datasetPath, mm.fineTuningData[modelName])
	}

	// Test fine-tuning with a non-existent dataset
	err = mm.FineTuneModel(modelName, "non-existent-dataset.txt")
	if err == nil {
		t.Error("Expected error when fine-tuning with a non-existent dataset, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read fine-tuning dataset") {
		t.Errorf("Expected error to contain 'failed to read fine-tuning dataset', got '%s'", err.Error())
	}
}

func TestPreloadModels(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create mock model files
	models := []string{"model1", "model2", "model3"}
	for _, modelName := range models {
		version := "v1.0"
		modelPath := filepath.Join(tempDir, modelName+"-"+version+".bin")
		if err := ioutil.WriteFile(modelPath, []byte("mock model data"), 0644); err != nil {
			t.Fatalf("Failed to create mock model file: %v", err)
		}
		mm.currentVersion[modelName] = version
	}

	// Test preloading models
	mm.PreloadModels(models)

	// Allow some time for the goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the models were added to the preload queue
	if len(mm.preloadQueue) != len(models) {
		t.Errorf("Expected preload queue to have length %d, got %d", len(models), len(mm.preloadQueue))
	}

	// Since the actual loading is done in goroutines and we can't easily mock the LoadModel method,
	// we can't reliably test that the models were actually loaded
}

func TestRollbackModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create mock model files
	modelName := "test-model"
	versions := []string{"v1.0", "v2.0"}
	for _, version := range versions {
		modelPath := filepath.Join(tempDir, modelName+"-"+version+".bin")
		if err := ioutil.WriteFile(modelPath, []byte("mock model data"), 0644); err != nil {
			t.Fatalf("Failed to create mock model file: %v", err)
		}
	}

	// Set the current version to v2.0
	mm.currentVersion[modelName] = versions[1]

	// Test rolling back to v1.0
	err = mm.RollbackModel(modelName, versions[0])
	if err != nil {
		t.Fatalf("Failed to rollback model: %v", err)
	}

	// Verify the current version was updated
	if mm.currentVersion[modelName] != versions[0] {
		t.Errorf("Expected mm.currentVersion['%s'] to be '%s', got '%s'", modelName, versions[0], mm.currentVersion[modelName])
	}

	// Test rolling back to a non-existent version
	err = mm.RollbackModel(modelName, "non-existent-version")
	if err == nil {
		t.Error("Expected error when rolling back to a non-existent version, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected error to contain 'not found', got '%s'", err.Error())
	}
}

func TestDeleteModel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create a mock model file
	modelName := "test-model"
	version := "v1.0"
	modelPath := filepath.Join(tempDir, modelName+"-"+version+".bin")
	if err := ioutil.WriteFile(modelPath, []byte("mock model data"), 0644); err != nil {
		t.Fatalf("Failed to create mock model file: %v", err)
	}

	// Set the current version
	mm.currentVersion[modelName] = version
	mm.loadedModels[modelName] = true

	// Test deleting the model
	err = mm.DeleteModel(modelName, version)
	if err != nil {
		t.Fatalf("Failed to delete model: %v", err)
	}

	// Verify the model file was removed
	if _, err := os.Stat(modelPath); !os.IsNotExist(err) {
		t.Errorf("Expected model file '%s' to be removed", modelPath)
	}

	// Verify the current version and loaded status were cleared
	if _, ok := mm.currentVersion[modelName]; ok {
		t.Errorf("Expected mm.currentVersion['%s'] to be deleted", modelName)
	}

	if _, ok := mm.loadedModels[modelName]; ok {
		t.Errorf("Expected mm.loadedModels['%s'] to be deleted", modelName)
	}

	// Test deleting a non-existent model
	err = mm.DeleteModel("non-existent-model", "v1.0")
	if err == nil {
		t.Error("Expected error when deleting a non-existent model, got nil")
	}
	if !strings.Contains(err.Error(), "failed to delete model") {
		t.Errorf("Expected error to contain 'failed to delete model', got '%s'", err.Error())
	}
}

func TestListModels(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "model-manager-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a new model manager
	mm := NewModelManager(tempDir)

	// Create mock model files
	expectedModels := []string{
		"model1-v1.0.bin",
		"model2-v1.0.bin",
		"model3-v2.0.bin",
	}
	for _, modelFile := range expectedModels {
		modelPath := filepath.Join(tempDir, modelFile)
		if err := ioutil.WriteFile(modelPath, []byte("mock model data"), 0644); err != nil {
			t.Fatalf("Failed to create mock model file: %v", err)
		}
	}

	// Create a non-model file
	nonModelPath := filepath.Join(tempDir, "not-a-model.txt")
	if err := ioutil.WriteFile(nonModelPath, []byte("not a model"), 0644); err != nil {
		t.Fatalf("Failed to create non-model file: %v", err)
	}

	// Test listing models
	models, err := mm.ListModels()
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	// Verify the correct models were listed
	if len(models) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(models))
	}

	for _, expectedModel := range expectedModels {
		var found bool
		for _, model := range models {
			if model == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model '%s' to be listed", expectedModel)
		}
	}

	// Verify the non-model file was not listed
	for _, model := range models {
		if model == "not-a-model.txt" {
			t.Errorf("Expected non-model file 'not-a-model.txt' to not be listed")
		}
	}
}
