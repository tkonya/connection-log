## Connection Logger

This is a lightweight container which periodically logs internet connection liveness (ping) and speed, and emails the results of these in csv format. The intended use-case is to track whether you are getting the speeds you pay for with your ISP.

This container should run as a daemon, aka running all the time. Memory consumption when idle is ~5MiB. A gmail address is required for sending the email. Can send to any valid email address.

### Required environmental variables:

EMAIL_USER - full gmail address, the emailing portion of this is hardcoded to work with gmail

EMAIL_PASS - password to above account

EMAIL_TO - email address to send the email to

### Optional environmental Variables

PING_INTERVAL - time between ping checks - default "30s"

SPEED_INTERVAL - time between speed checks - default "20m"

EMAIL_INTERVAL - time between emailing the report - default "168h" (1 week)

[See this link on how to specify duration](https://golang.org/pkg/time/#ParseDuration).

Uses [github.com/sparrc/go-ping](https://github.com/sparrc/go-ping) for pings.

Uses [Ookla Speedtest CLI](https://www.speedtest.net/apps/cli) for speed tests.

Uses [github.com/jordan-wright/email](https://github.com/jordan-wright/email) for email.

[Container Image on Docker Hub](https://hub.docker.com/repository/docker/tkonya/connection-logger/general)