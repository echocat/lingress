package support

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
)

type LocalizationContext struct {
	Bundle         *i18n.Bundle
	AcceptLanguage string
}

func (instance *LocalizationContext) Message(id string) string {
	return instance.MessageOrDefault("", id)
}

func (instance *LocalizationContext) MessageOrDefault(fallbackId string, id string) string {
	if val, err := i18n.NewLocalizer(instance.Bundle, instance.AcceptLanguage).Localize(&i18n.LocalizeConfig{
		MessageID: id,
	}); err != nil {
		if _, ok := err.(*i18n.MessageNotFoundErr); ok {
			if valDefault, err := i18n.NewLocalizer(instance.Bundle, "en-US").Localize(&i18n.LocalizeConfig{
				MessageID: id,
			}); err != nil {
				if _, ok := err.(*i18n.MessageNotFoundErr); ok && fallbackId != "" {
					return instance.Message(fallbackId)
				} else {
					log.WithError(err).
						WithField("accept", instance.AcceptLanguage).
						WithField("id", id).
						Warn("There was a message id requested which does not exist; will respond with empty string.")
					return ""
				}
			} else {
				return valDefault
			}
		} else {
			log.WithError(err).
				WithField("accept", instance.AcceptLanguage).
				WithField("id", id).
				Warn("There was a message id requested which does not exist; will respond with empty string.")
			return ""
		}
	} else {
		return val
	}
}
