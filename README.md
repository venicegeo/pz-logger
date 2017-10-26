# Running pz-logger locally

To run logger, golang 1.7 and a valid go environment are required. Installation instructions can be found here: https://golang.org/doc/install

In order for logger to successfully start it needs access to running, local ElasticSearch instance.
Additionally, the environment variable `LOGGER_INDEX` must be set; the value of this will be the name of the index in ElasticSearch containing logs.

Execute:
```
mkdir $GOPATH/src
mkdir $GOPATH/src/github.com/
mkdir $GOPATH/src/github.com/venicegeo
cd $GOPATH/src/github.com/venicegeo/
git clone https://github.com/venicegeo/pz-logger
cd pz-logger
go build
./pz-logger
```