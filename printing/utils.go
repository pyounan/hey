package printing

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"

	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"pos-proxy/config"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/01walid/goarabic"
	"github.com/abadojack/whatlanggo"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

//SetLang function to set language comes from configurations file
func SetLang(rcrs string) string {
	lang := ""
	if rcrs != "" && config.Config.IsFDMEnabled == true {
		for _, fdm := range config.Config.FDMs {
			if fdm.RCRS == rcrs {
				lang = fdm.Language
				return lang
			}
		}
	} else if config.Config.Language != "" {
		lang = config.Config.Language
		return lang

	}
	lang = "en-us"
	return lang
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
func Translate(text string, lang string) string {
	I18n := i18n.New(
		yaml.New("../../locales"),
	)
	translatedText := string(I18n.T(lang, text))
	return translatedText
}

//AddLine return lines for receipt
func AddLine(lineType string, charPerLine int) string {
	if lineType == "doubledashed" {

		return strings.Repeat("=", charPerLine)
	} else if lineType == "dash" {
		return strings.Repeat("-", charPerLine)
	}
	return ""
}

//GetImage downloads and save images from a url
func GetImage(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		str := fmt.Sprintf("failed to make request, returned error of %d \n", response.StatusCode)
		return "", errors.New(str)
	}

	file, err := os.Create("/tmp/logo.png")
	if err != nil {
		return "", err
	}
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	file.Close()

	return file.Name(), nil
}

//GetImageDimension returns the width and height for an image
func GetImageDimension(imagePath string) (string, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", "", err
	}
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", "", err
	}
	return strconv.Itoa(image.Width), strconv.Itoa(image.Height), nil
}

//Send sends a http request
func Send(api string, payload []byte) (int, []byte, error) {
	c := &http.Client{}
	body := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", api, body)
	if err != nil {
		log.Println(err)
		return 0, nil, err
	}
	// req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
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
	return resp.StatusCode, respBody, nil
}
