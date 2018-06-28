package printing

import (
	"strings"
	"unicode/utf8"

	"github.com/01walid/goarabic"
	"github.com/abadojack/whatlanggo"
)

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
