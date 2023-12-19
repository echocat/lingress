package def

import (
	"embed"
	"github.com/echocat/lingress/file/providers"
)

var (
	//go:embed localization
	localizations embed.FS
	//go:embed templates
	templates embed.FS

	def = providers.DefaultFileProviders{
		Localization: providers.FileProviderStrippingPrefix(localizations, "localization"),
		Static:       providers.NoopFileProvider(),
		Templates:    providers.FileProviderStrippingPrefix(templates, "templates"),
	}
)

func Get() providers.FileProviders {
	return def
}
