package printing

import (
	"strings"

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
func Center(wordLen int) string {
	return Pad((40 - wordLen) / 2)

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
