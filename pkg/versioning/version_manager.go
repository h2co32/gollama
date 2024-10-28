package versioning

import "fmt"

type ModelVersionManager struct {
	currentVersion string
	versionHistory map[string]string
}

func NewModelVersionManager(initialVersion string) *ModelVersionManager {
	return &ModelVersionManager{
		currentVersion: initialVersion,
		versionHistory: make(map[string]string),
	}
}

func (mvm *ModelVersionManager) SetVersion(version string) {
	mvm.versionHistory[mvm.currentVersion] = version
	mvm.currentVersion = version
}

func (mvm *ModelVersionManager) Rollback() error {
	if prevVersion, exists := mvm.versionHistory[mvm.currentVersion]; exists {
		mvm.currentVersion = prevVersion
		return nil
	}
	return fmt.Errorf("no previous version to rollback to")
}
