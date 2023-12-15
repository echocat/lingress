package providers

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

func (this *stripPrefixFileProvider) Open(name string) (fs.File, error) {
	return this.delegate.Open(path.Join(this.prefix, name))
}

func (this *stripPrefixFileProvider) ReadFile(name string) ([]byte, error) {
	return this.delegate.ReadFile(path.Join(this.prefix, name))
}

func (this *stripPrefixFileProvider) ReadDir(name string) ([]fs.DirEntry, error) {
	return this.delegate.ReadDir(path.Join(this.prefix, name))
}

func NoopFileProvider() FileProvider {
	return noopFileProviderV
}

var noopFileProviderV = &noopFileProvider{}

type noopFileProvider struct{}

func (this *noopFileProvider) Open(string) (fs.File, error) {
	return nil, os.ErrNotExist
}

func (this *noopFileProvider) ReadFile(string) ([]byte, error) {
	return nil, os.ErrNotExist
}

func (this *noopFileProvider) ReadDir(string) ([]fs.DirEntry, error) {
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

func (this DefaultFileProviders) GetLocalization() FileProvider {
	return this.Localization
}

func (this DefaultFileProviders) GetStatic() FileProvider {
	return this.Static
}

func (this DefaultFileProviders) GetTemplates() FileProvider {
	return this.Templates
}
