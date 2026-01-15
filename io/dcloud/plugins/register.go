package plugins

import "github.com/microsoft/typescript-go/io/dcloud"

func init() {
	dcloud.InstallPluginCreator("test-plugin", NewTestPlugin)
}