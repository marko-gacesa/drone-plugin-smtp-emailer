// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	SmtpHost             string `envconfig:"PLUGIN_SMTP_HOST"`
	SmtpPort             string `envconfig:"PLUGIN_SMTP_PORT"`
	SmtpUsername         string `envconfig:"PLUGIN_SMTP_USERNAME"`
	SmtpPassword         string `envconfig:"PLUGIN_SMTP_PASSWORD"`
	EmailSender          string `envconfig:"PLUGIN_EMAIL_SENDER"`
	EmailRecipient       string `envconfig:"PLUGIN_EMAIL_RECIPIENT"`
	EmailContentType     string `envconfig:"PLUGIN_EMAIL_CONTENT_TYPE"`
	EmailTemplateSubject string `envconfig:"PLUGIN_EMAIL_TEMPLATE_SUBJECT"`
	EmailTemplateBody    string `envconfig:"PLUGIN_EMAIL_TEMPLATE_BODY"`
	AttachFile           string `envconfig:"PLUGIN_ATTACH_FILE"`
}

// Validate checks and sets defaults values to some fields in an Args structure.
func (args *Args) Validate() (err error) {
	args.SmtpHost = strings.TrimSpace(args.SmtpHost)
	args.SmtpPort = strings.TrimSpace(args.SmtpPort)

	if args.SmtpHost == "" {
		err = errors.New("SMTP hostname is missing")
		return
	}

	if args.SmtpPort == "" {
		err = errors.New("SMTP port is missing")
	}

	if args.EmailRecipient == "" {
		err = errors.New("email recipient can't be empty")
		return
	}

	if args.EmailContentType == "" {
		// default to content type text/html
		args.EmailContentType = contentTypeHtml
	} else if args.EmailContentType != contentTypeHtml && args.EmailContentType != contentTypeText {
		// prevent any other type except plain text and html
		err = errors.New("email content type must be either text/html or text/plain")
		return
	}

	return
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) (err error) {
	logrus.Infof("Processing %q at %q\n", args.Build.Event, args.Repo.Link)

	err = args.Validate()
	if err != nil {
		return
	}

	smtpHost := args.SmtpHost
	smtpPort, err := strconv.Atoi(args.SmtpPort)
	if err != nil {
		err = errors.New("SMTP port must hold an integer: " + args.SmtpPort)
		return
	}

	if smtpPort <= 0 || smtpPort >= 65536 {
		err = errors.New("SMTP port is out of range: " + args.SmtpPort)
		return
	}

	workDir, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to get work directory: %w", err)
		return
	}

	isSuccess := args.Stage.Status != "failure" && len(args.Failed.Steps) == 0 && len(args.Failed.Stages) == 0
	data := map[string]interface{}{
		"Commit":    &args.Commit,
		"Author":    &args.Commit.Author,
		"Repo":      &args.Repo,
		"Stage":     &args.Stage,
		"Build":     &args.Build,
		"IsSuccess": isSuccess,
	}

	body, err := makeBody(workDir, args.EmailTemplateBody, args.EmailContentType, data)
	if err != nil {
		err = fmt.Errorf("failed to create email body: %w", err)
		return
	}

	subject, err := makeSubject(workDir, args.EmailTemplateSubject, data)
	if err != nil {
		err = fmt.Errorf("failed to create email subject: %w", err)
		return
	}

	m, err := prepareMessage(args.EmailSender, args.EmailRecipient, subject, body, args.EmailContentType, workDir, args.AttachFile)
	if err != nil {
		err = fmt.Errorf("failed to prepate email: %w", err)
		return
	}

	err = gomail.NewDialer(smtpHost, smtpPort, args.SmtpUsername, args.SmtpPassword).DialAndSend(m)
	if err != nil {
		err = fmt.Errorf("failed to send email: %w", err)
		return
	}

	logrus.Infof("Sent email to %s about event %q at repo %q\n", args.EmailRecipient, args.Build.Event, args.Repo.Link)

	return
}

func prepareMessage(sender, recipient, subject, body, contentType, baseDir, attachFileName string) (m *gomail.Message, err error) {
	m = gomail.NewMessage()

	if sender != "" {
		m.SetHeader("From", sender)
	}
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody(contentType, body)

	err = attachFile(m, baseDir, attachFileName)
	if err != nil {
		err = fmt.Errorf("failed to attach file: %w", err)
		return
	}

	return
}

func attachFile(m *gomail.Message, baseDir, fileName string) (err error) {
	if fileName == "" {
		return
	}

	absPath := joinPath(baseDir, fileName)

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		err = nil
		logrus.Warnf("Attachment file %q is missing\n", absPath)
		return
	} else if err != nil {
		return
	}

	m.Attach(absPath)

	return
}
