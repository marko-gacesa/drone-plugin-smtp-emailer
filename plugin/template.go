package plugin

import (
	_ "embed"
	htemplate "html/template"
	"strings"
	ttemplate "text/template"

	"github.com/drone/funcmap"
)

//go:embed email_body.tmpl
var emailBody string

//go:embed email_subject.tmpl
var emailSubj string

const contentTypeHtml = "text/html"
const contentTypeText = "text/plain"

var (
	templateBody *htemplate.Template
	templateSubj *ttemplate.Template
)

func init() {
	var err error

	if emailBody == "" {
		panic("email body template is missing")
	}

	if emailSubj == "" {
		panic("email subject template is missing")
	}

	templateBody, err = htemplate.New("").Funcs(funcmap.Funcs).Parse(emailBody)
	if err != nil {
		panic(err)
	}

	templateSubj, err = ttemplate.New("").Funcs(funcmap.Funcs).Parse(emailSubj)
	if err != nil {
		panic(err)
	}
}

func makeSubject(baseDir, fileName string, data map[string]interface{}) (subject string, err error) {
	sb := &strings.Builder{}
	sb.Grow(64)

	// if file name is empty use the default subject template to generate the subject text
	if fileName == "" {
		err = templateSubj.Execute(sb, data)
		if err != nil {
			return
		}

		subject = sb.String()
		return
	}

	// subject is always plain text, so use text template engine
	temp, err := ttemplate.New(fileName).Funcs(funcmap.Funcs).ParseFiles(joinPath(baseDir, fileName))
	if err != nil {
		return
	}

	err = temp.Execute(sb, data)
	if err != nil {
		return
	}

	subject = sb.String()

	return
}

func makeBody(baseDir, fileName, contentType string, data map[string]interface{}) (body string, err error) {
	sb := &strings.Builder{}
	sb.Grow(512)

	// if file name is empty use the default body template to generate the body text
	// note: the default template is using HTML engine
	if fileName == "" {
		err = templateBody.Execute(sb, data)
		if err != nil {
			return
		}

		body = sb.String()
		return
	}

	if contentType == contentTypeText {
		var temp *ttemplate.Template

		temp, err = ttemplate.New(fileName).Funcs(funcmap.Funcs).ParseFiles(joinPath(baseDir, fileName))
		if err != nil {
			return
		}

		err = temp.Execute(sb, data)
		if err != nil {
			return
		}
	} else {
		var temp *htemplate.Template

		temp, err = htemplate.New(fileName).Funcs(funcmap.Funcs).ParseFiles(joinPath(baseDir, fileName))
		if err != nil {
			return
		}

		err = temp.Execute(sb, data)
		if err != nil {
			return
		}
	}

	body = sb.String()

	return
}
