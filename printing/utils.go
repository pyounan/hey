package printing

import (
	"pos-proxy/config"
	"strings"
	"unicode/utf8"

	"github.com/01walid/goarabic"
	"github.com/abadojack/whatlanggo"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

var LANG string

//SetLang function to set language comes from configurations file
func SetLang(rcrs string) string {
	if rcrs != "" && config.Config.IsFDMEnabled == true {
		for _, fdm := range config.Config.FDMs {
			if fdm.RCRS == rcrs {
				LANG = fdm.Language
				return LANG
			}
		}
	} else if config.Config.Language != "" {
		LANG = config.Config.Language
		return LANG
	}
	return "en-us"
}

//Pad insert padding between text
func Pad(size int) string {
	if size > 0 {
		return strings.Join(make([]string, size+1), " ")

	}
	return ""
}

//Center to Center and Image
func Center(phrase string, paperWidth int) string {
	wordLen := utf8.RuneCountInString(phrase)
	if paperWidth == 800 {

		return Pad((40 - wordLen) / 2)
	}
	return Pad((32 - wordLen) / 2)
}

//CheckLang to check language of the text before printing it to do the necessary processing
func CheckLang(phrase string) string {
	info := whatlanggo.Detect(phrase)
	if whatlanggo.Scripts[info.Script] == "Arabic" {
		arabicText := goarabic.Reverse(goarabic.ToGlyph(phrase))
		return arabicText
	}
	return phrase
}

//Translate takes a text and translate it to a language comes fromthe configurations
func Translate(text string) string {
	I18n := i18n.New(
		yaml.New("../locales"),
	)
	translatedText := string(I18n.T(LANG, text))
	return translatedText
}
