# jsonstruct

`jsonstruct` is both a library and a command line tool to produce Go structs based on example JSON text.

## Example

The following example JSON file

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

Will produce this output

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
