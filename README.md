# Godradis
A full-featured library for accessing the [Dradis REST API](https://dradisframework.com/support/guides/rest_api/) from Go programs.

[![Documentation](https://godoc.org/github.com/njfox/godradis?status.svg)](http://godoc.org/github.com/njfox/godradis)

## Getting Started
```
$ go get -u github.com/njfox/godradis/...
```

Then import the library for use in other Go projects. E.g.:

```go
gd := godradis.Godradis{}
gd.Configure("https://example.com", "abcdefghijkl", false)
project, _ := gd.GetProjectByName("Example Network Penetration Test")
node, _ := gd.GetNodeByLabel(&project, "127.0.0.1")
for _, evidence := range node.Evidence {
	fmt.Printf("%v", evidence.GetField(Port))
}
```

## Limitations
The following API endpoints have not been implemented yet:

* Document Properties
* Content Blocks

Additionally, the Attachments endpoint has not been thoroughly tested.
