package dcloud

import (
	"context"
	"sync"

	"github.com/microsoft/typescript-go/internal/ls"
	"github.com/microsoft/typescript-go/internal/lsp/lsproto"
	dis "github.com/microsoft/typescript-go/io/dcloud/disposable"
)

type PluginLanguageService interface{
	LanguageService
	// IsEnable(fileName lsproto.DocumentUri)bool
}
type Plugin interface{
	dis.Disposable
	GetLanguageService(ctx context.Context, defaultLs *ls.LanguageService, documentURI lsproto.DocumentUri) PluginLanguageService
}

type PluginCreator func(project *Project) (Plugin, error)
var pluginHandle = sync.OnceValues(func()(func(string, PluginCreator), func() map[string]PluginCreator){
	creators := make(map[string]PluginCreator)
	var mu sync.Mutex

	register := func(id string, creator PluginCreator){
		mu.Lock()
		defer mu.Unlock()
		creators[id] = creator
	}

	get := func() map[string]PluginCreator{
		mu.Lock()
		defer mu.Unlock()
		return creators
	}

	return register, get
})
func InstallPluginCreator(pluginId string, creator PluginCreator) {
	register, _ := pluginHandle()
	register(pluginId, creator)
}
func GetPluginCreators() map[string]PluginCreator {
	_, get := pluginHandle()
	return get()
}

