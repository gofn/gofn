## How to run

First, you need to provide `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` variables.
You can get them at the AWS dashboard, from the AWS menus, select Services > IAM, create a new user and generate the keys.

```
$ export AWS_ACCESS_KEY=value-of-your-access-key
$ export AWS_SECRET_KEY=value-of-your-secret-key
$ go run main.go
```
