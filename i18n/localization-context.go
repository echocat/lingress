package i18n

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

func (this *LocalizationContext) LangBy(id string) language.Tag {
	_, tag := this.messageOrDefault("", id)
	return tag
}

func (this *LocalizationContext) Message(id string) string {
	return this.MessageOrDefault("", id)
}

func (this *LocalizationContext) MessageOrDefault(fallbackId string, id string) string {
	val, _ := this.messageOrDefault(fallbackId, id)
	return val
}
func (this *LocalizationContext) messageOrDefault(fallbackId string, id string) (string, language.Tag) {
	if val, tag, err := i18n.NewLocalizer(this.Bundle, this.AcceptLanguage).LocalizeWithTag(&i18n.LocalizeConfig{
		MessageID: id,
	}); err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			if valDefault, tagDefault, err := i18n.NewLocalizer(this.Bundle, "en-US").LocalizeWithTag(&i18n.LocalizeConfig{
				MessageID: id,
			}); err != nil {
				if _, ok := err.(*i18n.MessageNotFoundErr); ok && fallbackId != "" {
					return this.Message(fallbackId), language.Tag{}
				} else {
					this.Logger.
						WithError(err).
						With("accept", this.AcceptLanguage).
						With("id", id).
						Warn("There was a message id requested which does not exist; will respond with empty string.")
					return "", fallbackTag
				}
			} else {
				return valDefault, tagDefault
			}
		} else {
			this.Logger.
				WithError(err).
				With("accept", this.AcceptLanguage).
				With("id", id).
				Warn("There was a message id requested which does not exist; will respond with empty string.")
			return "", fallbackTag
		}
	} else {
		return val, tag
	}
}
