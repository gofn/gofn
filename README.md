# gofn
Function process via container provider

## Premise

- Software makes a task
- After processing it dies
- Must put on stdout (print) a string formatted as JSON

## Run

```bash
    go run main.go -contextDir=testDocker -imageName=python -dockerFile=Dockerfile 
```

You can also compile with _go build_ or build and install with _go install_ command then run it as a native executable.

#### Parameters

- -contextDir is the root directory where the Dockerfile, scripts, and other container dependencies are, by default current directory "./".

- -imageName is the name of the image you want to start, if it does not exist it will be automatically generated and if it exists the system will just start the container.

- -dockerFile is the name of the file containing the container settings, by default Dockerfile

- -h Shows the list of parameters

gofn generates the images with "gofn/" as a prefix.

