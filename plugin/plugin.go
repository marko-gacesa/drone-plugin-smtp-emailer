// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"errors"
	"fmt"
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

	SmtpHost       string `envconfig:"PLUGIN_SMTP_HOST"`
	SmtpPort       string `envconfig:"PLUGIN_SMTP_PORT"`
	SmtpUsername   string `envconfig:"PLUGIN_SMTP_USERNAME"`
	SmtpPassword   string `envconfig:"PLUGIN_SMTP_PASSWORD"`
	EmailSender    string `envconfig:"PLUGIN_EMAIL_SENDER"`
	EmailRecipient string `envconfig:"PLUGIN_EMAIL_RECIPIENT"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) (err error) {
	logrus.Infof("Processing %q at %q\n", args.Build.Event, args.Repo.Link)

	var (
		smtpHost string
		smtpPort int
	)

	args.SmtpHost = strings.TrimSpace(args.SmtpHost)
	args.SmtpPort = strings.TrimSpace(args.SmtpPort)

	if args.SmtpHost == "" {
		err = errors.New("SMTP hostname is missing")
		return
	}

	smtpHost = args.SmtpHost

	if args.SmtpPort == "" {
		smtpPort = 25
	} else {
		smtpPort, err = strconv.Atoi(args.SmtpPort)
		if err != nil {
			err = errors.New("SMTP port must hold an integer: " + args.SmtpPort)
			return
		}

		if smtpPort <= 0 || smtpPort >= 65536 {
			err = errors.New("SMTP port is out of range: " + args.SmtpPort)
			return
		}
	}

	if args.EmailRecipient == "" {
		err = errors.New("email recipient can't be empty")
		return
	}

	isSuccess := args.Stage.Status != "failure" && len(args.Failed.Steps) == 0 && len(args.Failed.Stages) == 0
	var statusText string
	if isSuccess {
		statusText = "SUCCESS"
	} else {
		statusText = "FAILURE"
	}

	subject := fmt.Sprintf("Drone [%s] %s at %s", statusText, args.Build.Event, args.Repo.Link)
	data := map[string]interface{}{
		"Commit":    &args.Commit,
		"Author":    &args.Commit.Author,
		"Repo":      &args.Repo,
		"Build":     &args.Build,
		"IsSuccess": isSuccess,
		"Status":    statusText,
	}

	bodyBuilder := &strings.Builder{}
	bodyBuilder.Grow(512)
	err = emailTemplate.Execute(bodyBuilder, data)
	if err != nil {
		err = fmt.Errorf("failed to create email body: %w", err)
		return
	}

	body := bodyBuilder.String()

	m := gomail.NewMessage()

	if args.EmailSender != "" {
		m.SetHeader("From", args.EmailSender)
	}
	m.SetHeader("To", args.EmailRecipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	err = gomail.NewDialer(smtpHost, smtpPort, args.SmtpUsername, args.SmtpPassword).DialAndSend(m)
	if err != nil {
		err = fmt.Errorf("failed to send email: %w", err)
		return
	}

	logrus.Infof("Sent email to %s about event %q at repo %q\n", args.EmailRecipient, args.Build.Event, args.Repo.Link)

	return
}
