# KDLGo

[![GoDoc](https://godoc.org/github.com/frixuu/kdlgo?status.svg)](https://godoc.org/github.com/frixuu/kdlgo)

WIP Go parser for the [KDL Document Language](https://github.com/kdl-org/kdl), version 1.0.0.

## Current status

- [x] parsing to a kdl.Document model
- [x] serializing a kdl.Document model to a string
- [ ] marshalling from a struct
- [ ] unmarshalling to a struct
- [ ] improve performance?

## Usage

```go
import (
	kdl "github.com/frixuu/kdlgo"
)
```

### Parse (to a Document model)

```go
// or any of: ParseBytes, ParseFile, ParseReader
document, err := kdl.ParseString(`foo bar="baz"`)
```

### Modify the Document

```go
if document.Nodes[0].HasProp("bar") {
	n := kdl.NewNode("person")
	n.AddArg("known")
	// or: n.AddArgValue(kdl.NewStringValue("known", kdl.NoHint()))
	n.SetProp("name", "Joe")
	// or: n.SetPropValue("name", kdl.NewStringValue("Joe", kdl.NoHint()))
	document.AddChild(n)
}
```

### Serialize the Document

```go
// or Write() to an io.Writer
s, err := document.WriteString()
```
