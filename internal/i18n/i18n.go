// Package i18n provides bilingual (zh-CN / en-US) message support.
package i18n

// Lang represents a supported language.
type Lang string

const (
	EN Lang = "en-US"
	ZH Lang = "zh-CN"
)

// Locale holds the current language and provides translation.
type Locale struct {
	lang Lang
}

// NewLocale creates a Locale with the default language (zh-CN).
func NewLocale() *Locale {
	return &Locale{lang: ZH}
}

// Set changes the current language.
func (l *Locale) Set(lang Lang) {
	l.lang = lang
}

// Current returns the current language.
func (l *Locale) Current() Lang {
	return l.lang
}

// T returns the translation for a given key in the current language.
// Falls back to en-US then to the key itself.
func (l *Locale) T(key string) string {
	if m, ok := messages[l.lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	if s, ok := messages[EN][key]; ok {
		return s
	}
	return key
}

// messages is populated from zh_cn.go and en_us.go via init functions.
var messages = map[Lang]map[string]string{
	ZH: zhMessages,
	EN: enMessages,
}
