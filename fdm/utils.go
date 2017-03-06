package fdm

import (
	"crypto/sha1"
	"fmt"
)

// ApplySHA1 convert text to SHA1
func ApplySHA1(text string) string {
	msg := sha1.New()
	msg.Write([]byte(text))
	bs := msg.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func GeneratePLUHash(items []POSLineItem) string {
	text := ""
	for _, i := range items {
		text += fmt.Sprintf("%s", i.String())
	}
	fmt.Printf("plu before hasshing: %s\n", text)
	return ApplySHA1(text)
}
