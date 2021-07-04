# jsonstruct

`jsonstruct` is both a library and a command line tool to produce Go structs based on example JSON text.

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
    ]
}
```

The above JSON object Will produce this output:

```golang
type Test struct {
        Currency     string     `json:"currency"`
        Amount       float64    `json:"amount"`
        Map          *Map       `json:"map"`
        Array        []float64  `json:"array"`
        StringArray  []string   `json:"string_array"`
        CamelKey     string     `json:"CamelKey"`
        BlahBlahBlah string     `json:"blahBlahBlah"`
        Structs      []*Structs `json:"structs"`
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

## TODO

* De-duplicate structs that are created more than once by multiple instances in the example JSON file
* Handle plural names for slice types
