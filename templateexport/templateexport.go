package templateexport

import (
	"html/template"
	"log"
)

var ExportedTemplates *template.Template

func ParseTemplates(templatesPath string) {
	var err error
	ExportedTemplates, err = template.ParseGlob(templatesPath)
	// templates, err = template.ParseGlob(filepath.Join(cwd, "templates/*"))
	if err != nil {
		log.Println("Failed to parse html templates", err.Error())
	}
}
