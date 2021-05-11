package plugin

import (
	_ "embed"
	"html/template"
)

//go:embed email.tmpl
var emailTemplateText string

var emailTemplate *template.Template

func init() {
	var err error

	if emailTemplateText == "" {
		panic("email template is missing")
	}

	emailTemplate, err = template.New("email").Parse(emailTemplateText)
	if err != nil {
		panic(err)
	}
}
