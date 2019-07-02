package fallback

import (
	"github.com/gobuffalo/packr"
	"github.com/echocat/lingress/support"
	"html/template"
)

var (
	plainTemplates = packr.NewBox("../resources/templates")

	defaultFuncMap = func() template.FuncMap {
		tlc := &support.LocalizationContext{}
		return template.FuncMap{
			"join":          support.Join,
			"i18n":          tlc.Message,
			"i18nOrDefault": tlc.MessageOrDefault,
		}
	}()
)

func newTemplate(name string, funcMaps ...template.FuncMap) (*template.Template, error) {
	if plain, err := plainTemplates.FindString(name); err != nil {
		return nil, err
	} else {
		tmpl := template.New("resources/templates/" + name).Funcs(defaultFuncMap)
		for _, funcMap := range funcMaps {
			tmpl.Funcs(funcMap)
		}
		return tmpl.Parse(plain)
	}
}

func cloneAndLocalizeTemplate(in *template.Template, lc *support.LocalizationContext) (*template.Template, error) {
	if out, err := in.Clone(); err != nil {
		return nil, err
	} else {
		return out.Funcs(template.FuncMap{
			"i18n":          lc.Message,
			"i18nOrDefault": lc.MessageOrDefault,
		}), nil
	}
}
