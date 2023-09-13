# jsonstruct

[![test](https://github.com/cneill/jsonstruct/actions/workflows/test.yaml/badge.svg)](https://github.com/cneill/jsonstruct/actions/workflows/test.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cneill/jsonstruct.svg)](https://pkg.go.dev/github.com/cneill/jsonstruct)
[![Go Report Card](https://goreportcard.com/badge/github.com/cneill/jsonstruct)](https://goreportcard.com/report/github.com/cneill/jsonstruct)

`jsonstruct` is both a library and a command line tool to produce Go structs based on example JSON text. It's in the middle of
a refactor, so you probably don't want to rely on it too heavily right now.

## Installation

```bash
go install github.com/cneill/jsonstruct/cmd/jsonstruct@latest
```

## Usage

```
Usage of ./jsonstruct:
./jsonstruct [flags] [file name...]

Flags:
  -value-comments
        add a comment to struct fields with the example value(s)
```

## Examples

### JSON object

```json
{
    "currency": "value",
    "amount": 4.267,
    "map": {
        "something": "nothing",
        "this": true
    },
    "array": [
        1,
        2,
        3,
        4
    ],
    "string_array": [
        "string",
        "string2",
        "string3"
    ],
    "CamelKey": "blah",
    "blahBlahBlah": "blah",
    "structs": [
        {
            "stuff": "stuff"
        },
        {
            "stuff": "stuff2"
        },
        {
            "stuff": "stuff3",
            "differentStuff": "differentStuff"
        }
    ],
    "nothing": null
}
```

If passed in through stdin, the above JSON object will produce this output:

```golang
type Stdin1 struct {
        Currency     string   `json:"currency"`
        Amount       float64  `json:"amount"`
        Map          *Map     `json:"map"`
        Array        []int64  `json:"array"`
        StringArray  []string `json:"string_array"`
        CamelKey     string
        BlahBlahBlah string           `json:"blahBlahBlah"`
        Structs      []*Structs       `json:"structs"`
        Nothing      *json.RawMessage `json:"nothing"`
}

type Map struct {
        Something string `json:"something"`
        This      bool   `json:"this"`
}

type Structs struct {
        Stuff          string `json:"stuff"`
        DifferentStuff string `json:"differentStuff,omitempty"`
}
```

### JSON array of objects

```json
[
    {
        "stuff": "stuff"
    },
    {
        "stuff": "stuff2"
    },
    {
        "stuff": "stuff3",
        "differentStuff": "differentStuff"
    },
    {
        "stuff": 1,
        "differentStuff": "blah"
    }
]
```

```golang
type Stdin1 struct {
        Stuff          *json.RawMessage `json:"stuff"`
        DifferentStuff string           `json:"differentStuff,omitempty"`
}
```

## Notes

* When an array of JSON objects is detected, any keys that are provided in some objects but not others
  will get the `,omitempty` flag
* When the same field is detected in multiple objects in a JSON array with different value types, the
  Go type will be `*json.RawMessage`, which will contain the raw bytes of the field to allow for
  different types
* Defaults to a `*json.RawMessage` type when:
    * JSON `null` is provided
    * There are multiple types in e.g. an array
    * There is an empty array
* Can take input from either files passed in as CLI args or STDIN. Can take a stream of objects / arrays of objects.

## TODO

* De-duplicate structs that are created more than once by multiple instances in the example JSON file
* Handle plural names for slice types
