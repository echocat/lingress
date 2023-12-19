package fallback

import (
	"github.com/echocat/lingress/file/providers"
	"github.com/echocat/lingress/i18n"
	"github.com/echocat/lingress/support"
	"html/template"
)

var (
	defaultFuncMap = func() template.FuncMap {
		tlc := &i18n.LocalizationContext{}
		return template.FuncMap{
			"join":          support.Join,
			"i18n":          tlc.Message,
			"langBy":        tlc.LangBy,
			"i18nOrDefault": tlc.MessageOrDefault,
		}
	}()
)

func newTemplate(fp providers.FileProvider, name string, funcMaps ...template.FuncMap) (*template.Template, error) {
	if plain, err := fp.ReadFile(name); err != nil {
		return nil, err
	} else {
		tmpl := template.New("resources/templates/" + name).Funcs(defaultFuncMap)
		for _, funcMap := range funcMaps {
			tmpl.Funcs(funcMap)
		}
		return tmpl.Parse(string(plain))
	}
}

func cloneAndLocalizeTemplate(in *template.Template, lc *i18n.LocalizationContext) (*template.Template, error) {
	if out, err := in.Clone(); err != nil {
		return nil, err
	} else {
		return out.Funcs(template.FuncMap{
			"i18n":          lc.Message,
			"langBy":        lc.LangBy,
			"i18nOrDefault": lc.MessageOrDefault,
		}), nil
	}
}
