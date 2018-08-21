# pz-logger
The Logger provides a system-wide, common way to record log messages. The log messages are stored in Elasticsearch. This is done though an HTTP API.

## Requirements
Before building and running the pz-logger project, please ensure that the following components are available and/or installed, as necessary:
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (for checking out repository source)
- [Go](https://golang.org/doc/install) v1.7 or later
- [Glide](https://glide.sh)
- [ElasticSearch](https://www.elastic.co/)

***
## Setup
 
 Create the directory the repository must live in:

    $ mkdir -p $GOPATH/src/github.com/venicegeo
    $ cd $GOPATH/src/github.com/venicegeo
    $ git clone git@github.com:venicegeo/pz-logger.git
    $ cd pz-logger

Set up Go environment variables

To function right, Go must have some environment variables set. Run the `go env`
command to list all relevant environment variables. The two most important 
variables to look for are `GOROOT` and `GOPATH`.

- `GOROOT` must point to the base directory at which Go is installed
- `GOPATH` must point to a directory that is to serve as your development
  environment. This is where this code and dependencies will live.

To quickly verify these variables are set, run the command from terminal:

	$ go env | egrep "GOPATH|GOROOT"


### Configuring

In order for pz-logger to successfully start it needs access to running, local [ElasticSearch](https://www.elastic.co/) instance. If not currently available,  it can be downloaded and documentation can be found [here](https://www.elastic.co/downloads/elasticsearch).
Additionally, the environment variable `LOGGER_INDEX` must be set; the value of this will be the name of the index in ElasticSearch containing logs.

## Installing, Building, Running & Unit Tests

### Install dependencies

This project manages dependencies by populating a `vendor/` directory using the
glide tool. If the tool is already installed, in the code repository, run:

    $ glide install -v

This will retrieve all the relevant dependencies at their appropriate versions
and place them in `vendor/`, which enables Go to use those versions in building
rather than the default (which is the newest revision in Github).

> **Adding new dependencies.** When adding new dependencies, simply installing
  them with `go get <package>` will fetch their latest version and place it in
  `$GOPATH/src`. This is undesirable, since it is not repeatable for others.
  Instead, to add a dependency, use `glide get <package>`, which will place it
  in `vendor/` and update `glide.yaml` and `glide.lock` to remember its version.

### Build the project
To build `pz-logger`, run `go install` from the project directory. To build it from elsewhere, run:

	$ go install github.com/venicegeo/pz-logger

This will build and place a statically-linked Go executable at `$GOPATH/bin/pz-logger`.


### Run the project

	$ go build
	$ ./pz-logger
	
### Run unit tests with coverage collection

To run `pz-logger`, unit tests, run the command shown below. This
will run the unit tests for `pz-logger` and all its subpackages and print
coverage summaries.

	$ go test -cover github.com/venicegeo/pz-logger/logger

An example response should look similar to

	ok  	github.com/venicegeo/pz-logger/logger	4.064s	coverage: 76.8% of statements
