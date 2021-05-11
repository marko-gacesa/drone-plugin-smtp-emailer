A plugin to Drone Plug-in for sending status emails through a SMTP server.

# Usage

The following settings changes this plugin's behavior.

* SmtpHost
* SmtpPort
* SmtpUsername
* SmtpPassword
* EmailSender (optional)
* EmailRecipient

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
- name: run markogacesa/drone-plugin-smtp-emailer plugin
  image: markogacesa/drone-plugin-smtp-emailer
  pull: if-not-exists
  settings:
    smtp_host: smtp.gmail.com
    smtp_port: 456
    smtp_username: emailer42
    smtp_password: p4$$w0rd
    email_sender: noreply@drone.io
    email_recipient: watcher@example.com
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
