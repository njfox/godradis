# Godradis
A full-featured library for accessing the [Dradis REST API](https://dradisframework.com/support/guides/rest_api/) from Go programs.

[![Documentation](https://godoc.org/github.com/njfox/godradis?status.svg)](http://godoc.org/github.com/njfox/godradis)

## Getting Started
```
$ go get -u github.com/njfox/godradis/...
```

Documentation via godoc will be added when this repository is officially released. In the meantime, you can build the
documentation locally using `godoc`:

```
$ cd ~/go/src/github.com/njfox/godradis && godoc -http=:6060
```

Then browse to localhost:6060 to view the documentation.

## Limitations
The following API endpoints have not been implemented yet:

* IssueLibrary
* Document Properties
* Content Blocks

Additionally, the Attachments endpoint has not been thoroughly tested.