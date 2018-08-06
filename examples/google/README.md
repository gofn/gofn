## How to run

First, you need to provide `GOOGLE_PROJECT` and `GOOGLE_APPLICATION_CREDENTIALS` variables.
You can get them at the Google Cloud Platform dashboard (see: https://cloud.google.com/docs/authentication/getting-started).

```
$ export GOOGLE_PROJECT=value-of-your-project-id
$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials/file
$ go run main.go
```
