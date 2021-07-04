# jsonstruct

`jsonstruct` is both a library and a command line tool to produce Go structs based on example JSON text.

## Installation

`go install github.com/cneill/jsonstruct/cmd/...`

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

The above JSON object Will produce this output:

```golang
type Test struct {
        Amount       float64    `json:"amount"`
        Array        []float64  `json:"array"`
        CamelKey     string     `json:"CamelKey"`
        BlahBlahBlah string     `json:"blahBlahBlah"`
        Structs      []*Structs `json:"structs"`
        Currency     string     `json:"currency"`
        Map          *Map       `json:"map"`
        StringArray  []string   `json:"string_array"`
        Nothing      string     `json:"nothing"`
}

type Structs struct {
        DifferentStuff string `json:"differentStuff,omitempty"`
        Stuff          string `json:"stuff"`
}

type Map struct {
        Something string `json:"something"`
        This      bool   `json:"this"`
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
type ArrayTest struct {
        Stuff          *json.RawMessage `json:"stuff"`
        DifferentStuff string           `json:"differentStuff,omitempty"`
}
```

## Notes

* When an array of JSON objects is detected, any keys that are provided in some objects but not others
  will get the `,omitempty` flag
* All numbers will be treated as `float64` - this is how Go interprets all JSON numbers
* When the same field is detected in multiple objects in a JSON array with different value types, the
  Go type will be `*json.RawMessage`, which will contain the raw bytes of the field to allow for
  different types
* Defaults to a string type when JSON `null` is provided

## TODO

* De-duplicate structs that are created more than once by multiple instances in the example JSON file
* Handle plural names for slice types
