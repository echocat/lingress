package support

import (
	"io/fs"
	"os"
	"path"
)

type FileProvider interface {
	fs.ReadFileFS
	fs.ReadDirFS
}

func FileProviderStrippingPrefix(fp FileProvider, prefix string) FileProvider {
	return &stripPrefixFileProvider{prefix, fp}
}

type stripPrefixFileProvider struct {
	prefix   string
	delegate FileProvider
}

func (instance *stripPrefixFileProvider) Open(name string) (fs.File, error) {
	return instance.delegate.Open(path.Join(instance.prefix, name))
}

func (instance *stripPrefixFileProvider) ReadFile(name string) ([]byte, error) {
	return instance.delegate.ReadFile(path.Join(instance.prefix, name))
}

func (instance *stripPrefixFileProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	return instance.delegate.ReadDir(path.Join(instance.prefix, name))
}

func NoopFileProvider() FileProvider {
	return noopFileProviderV
}

var noopFileProviderV = &noopFileProvider{}

type noopFileProvider struct{}

func (instance *noopFileProvider) Open(name string) (fs.File, error) {
	return nil, os.ErrNotExist
}

func (instance *noopFileProvider) ReadFile(name string) ([]byte, error) {
	return nil, os.ErrNotExist
}

func (instance *noopFileProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, os.ErrNotExist
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
