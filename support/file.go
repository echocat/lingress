package support

import "net/http"

type FileProvider interface {
	List() []string
	Open(name string) (http.File, error)
	Find(filename string) ([]byte, error)
}

type FileProviders interface {
	GetLocalization() FileProvider
	GetStatic() FileProvider
	GetTemplates() FileProvider
}

type DefaultFileProviders struct {
	Localization FileProvider
	Static       FileProvider
	Templates    FileProvider
}

func (instance DefaultFileProviders) GetLocalization() FileProvider {
	return instance.Localization
}

func (instance DefaultFileProviders) GetStatic() FileProvider {
	return instance.Static
}

func (instance DefaultFileProviders) GetTemplates() FileProvider {
	return instance.Templates
}
