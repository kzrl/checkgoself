# checkgoself 


## Installation

    go get github.com/kzrl/checkgoself...
    go install github.com/kzrl/checkgoself...


## Usage

```
checkgoself
  -config="config.json": Path to config.json - defaults to the working directory
  -email=true: Send email alerts
  -help=false: Show usage
  -version=false: Show version
```

Runs through the checks defined in `config.json`. No output to syslog or STDOUT unless a check is out of bounds.




    testserver

A simple webserver listening on port 4242. It prints the contents of the request to STDOUT and HTTP 200 status to clients.
Used to verify that checkgoself is making Alarm GET requests properly

