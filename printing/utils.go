package printing

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func AddLine(lineType string, charPerLine int) string {
	if lineType == "doubledashed" {

		return strings.Repeat("=", charPerLine)
	} else if lineType == "dash" {
		return strings.Repeat("-", charPerLine)
	}
	return ""
}

func Send(api string, payload []byte) (int, []byte, error) {
	c := &http.Client{}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequest("POST", api, body)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/xml")
	resp, err := c.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, respBody, err
		}
		str := fmt.Sprintf("failed to make request, returned error of %d %s\n", resp.StatusCode, string(respBody))
		return resp.StatusCode, nil, errors.New(str)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, respBody, err
	}
	log.Println(respBody)
	return resp.StatusCode, respBody, nil
}
