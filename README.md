A plugin to Drone Plug-in for sending status emails through a SMTP server.

# Usage

The following settings change this plugin's behavior.

* SmtpHost (mandatory)
* SmtpPort (mandatory)
* SmtpUsername
* SmtpPassword
* EmailSender
* EmailRecipient (mandatory)
* EmailContentType (Can be either text/plain or text/html. The latter is default.)
* EmailTemplateSubject (File name where a custom Go template for email subject can be found. The file could be in the repo.)
* EmailTemplateBody (File name where a custom Go template for email body can be found. The default template is uses HTML.)
* AttachFile (File name that will be attached to emails. Useful for example for code coverage reports.)

If used, the custom templates for subject and body must be valid Go templates. The templates have access to the following objects:
* `Build` (struct)
* `Commit` (struct)
* `Author` (struct, just a pointer to Commit.Author)
* `Repo` (struct)
* `Stage` (struct)
* `IsSuccess` (bool)

For struct definition, see: https://raw.githubusercontent.com/drone/boilr-plugin/master/template/plugin/pipeline.go

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
- name: run tests
  image: golang
  pull: if-not-exists
  commands:
    - go test -coverprofile=coverage.out
    - go tool cover -html=coverage.out -o coverage.html
- name: run markogacesa/drone-plugin-smtp-emailer plugin
  image: markogacesa/drone-plugin-smtp-emailer
  pull: if-not-exists
  settings:
    smtp_host: smtp.gmail.com
    smtp_port: 587
    smtp_username: emailer42
    smtp_password: p4$$w0rd
    email_sender: noreply@drone.io
    email_recipient: watcher@example.com
    attach_file: coverage.html
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t markogacesa/drone-plugin-smtp-emailer -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm -e PLUGIN_PARAM1=foo -e PLUGIN_PARAM2=bar \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  markogacesa/drone-plugin-smtp-emailer
```
