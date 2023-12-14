package support

import (
	"github.com/echocat/slf4g"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	fallbackTag = language.Make("en-US")
)

type LocalizationContext struct {
	Bundle         *i18n.Bundle
	AcceptLanguage string
	Logger         log.Logger
}

func (instance *LocalizationContext) LangBy(id string) language.Tag {
	_, tag := instance.messageOrDefault("", id)
	return tag
}

func (instance *LocalizationContext) Message(id string) string {
	return instance.MessageOrDefault("", id)
}

func (instance *LocalizationContext) MessageOrDefault(fallbackId string, id string) string {
	val, _ := instance.messageOrDefault(fallbackId, id)
	return val
}
func (instance *LocalizationContext) messageOrDefault(fallbackId string, id string) (string, language.Tag) {
	if val, tag, err := i18n.NewLocalizer(instance.Bundle, instance.AcceptLanguage).LocalizeWithTag(&i18n.LocalizeConfig{
		MessageID: id,
	}); err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			if valDefault, tagDefault, err := i18n.NewLocalizer(instance.Bundle, "en-US").LocalizeWithTag(&i18n.LocalizeConfig{
				MessageID: id,
			}); err != nil {
				if _, ok := err.(*i18n.MessageNotFoundErr); ok && fallbackId != "" {
					return instance.Message(fallbackId), language.Tag{}
				} else {
					instance.Logger.
						WithError(err).
						With("accept", instance.AcceptLanguage).
						With("id", id).
						Warn("There was a message id requested which does not exist; will respond with empty string.")
					return "", fallbackTag
				}
			} else {
				return valDefault, tagDefault
			}
		} else {
			instance.Logger.
				WithError(err).
				With("accept", instance.AcceptLanguage).
				With("id", id).
				Warn("There was a message id requested which does not exist; will respond with empty string.")
			return "", fallbackTag
		}
	} else {
		return val, tag
	}
}
