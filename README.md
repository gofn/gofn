# gofn
Function process via container provider

## Premise

- Software makes a task
- After processing it dies
- Must put on stdout (print) a string formatted as JSON

## Run

    cd testDocker
    docker build . -t gofn-python
    cd ../
    go run main.go

