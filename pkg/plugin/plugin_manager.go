package plugin

import "fmt"

type ModelPlugin interface {
	Process(data interface{}) (interface{}, error)
	ModelName() string
}

type PluginManager struct {
	plugins map[string]ModelPlugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{plugins: make(map[string]ModelPlugin)}
}

func (pm *PluginManager) RegisterPlugin(plugin ModelPlugin) {
	pm.plugins[plugin.ModelName()] = plugin
}

func (pm *PluginManager) ExecutePlugin(modelName string, data interface{}) (interface{}, error) {
	plugin, exists := pm.plugins[modelName]
	if !exists {
		return nil, fmt.Errorf("no plugin registered for model: %s", modelName)
	}
	return plugin.Process(data)
}
