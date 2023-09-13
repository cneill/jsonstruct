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
NAME:
   jsonstruct - generate Go structs for JSON values

USAGE:
   jsonstruct [global options] command [command options] [file]...

DESCRIPTION:
   You can either pass in files as args or JSON in STDIN. Results are printed to STDOUT.

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --name value, -n value  override the default name derived from filename
   --value-comments, -c    add a comment to struct fields with the example value(s) (default: false)
   --sort-fields, -s       sort the fields in alphabetical order; default behavior is to mirror input (default: false)
   --inline-structs, -i    use inline structs instead of creating different types for each object (default: false)
   --print-filenames, -f   print the filename above the structs defined within (default: false)
   --debug                 enable debug logs (default: false)
   --help, -h              show help
```

## Examples

### JSON object

**Input:**

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
    "nested_numbers": [
        [1, 2, 3],
        [2, 3, 4],
        [3, 4, 5]
    ],
    "nothing": null
}
```

**Output:**

```golang
type Stdin1 struct {
        Currency      string   `json:"currency"`
        Amount        float64  `json:"amount"`
        Map           *Map     `json:"map"`
        Array         []int64  `json:"array"`
        StringArray   []string `json:"string_array"`
        CamelKey      string
        BlahBlahBlah  string           `json:"blahBlahBlah"`
        Structs       []*Structs       `json:"structs"`
        NestedNumbers [][]int64        `json:"nested_numbers"`
        Nothing       *json.RawMessage `json:"nothing"`
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

### Value comments (`-c`)

**Input:**

```json
{
  "test_bool": true,
  "test_int64": 1234,
  "test_float64": 1234.0,
  "test_string": "test",
  "test_array_of_bool": [true, false, true],
  "test_array_of_int64": [1, 2, 3, 4],
  "test_array_of_float64": [1.0, 2.0, 3.0, 4.0],
  "test_array_of_string": ["test1", "test2", "test3"],
  "test_struct": {
    "test_string": "test",
    "test_array_of_string": ["test1", "test2", "test3"]
  },
  "test_array_of_struct": [
    {
      "test_string": "test1",
      "test_optional_string": "test1"
    },
    {
      "test_string": "test2",
      "test_optional_string": "test2"
    },
    {
      "test_string": "test3"
    },
    {
      "test_string": "test4"
    }
  ],
  "test_garbage_array": [1, "1", 1.0],
  "terrible-name.for_a.key": "test"
}
```

**Output:**

```golang
type Stdin1 struct {
        TestBool            bool                 `json:"test_bool"`             // Example: true
        TestInt64           int64                `json:"test_int64"`            // Example: 1234
        TestFloat64         float64              `json:"test_float64"`          // Example: 1234.000
        TestString          string               `json:"test_string"`           // Example: "test"
        TestArrayOfBool     []bool               `json:"test_array_of_bool"`    // Example: [true, false, true]
        TestArrayOfInt64    []int64              `json:"test_array_of_int64"`   // Example: [1, 2, 3, 4]
        TestArrayOfFloat64  []float64            `json:"test_array_of_float64"` // Example: [1.000, 2.000, 3.000, 4.000]
        TestArrayOfString   []string             `json:"test_array_of_string"`  // Example: ["test1", "test2", "test3"]
        TestStruct          *TestStruct          `json:"test_struct"`
        TestArrayOfStruct   []*TestArrayOfStruct `json:"test_array_of_struct"`
        TestGarbageArray    []*json.RawMessage   `json:"test_garbage_array"`
        TerribleNameForAKey string               `json:"terrible-name.for_a.key"` // Example: "test"
}

type TestStruct struct {
        TestString        string   `json:"test_string"`          // Example: "test"
        TestArrayOfString []string `json:"test_array_of_string"` // Example: ["test1", "test2", "test3"]
}

type TestArrayOfStruct struct {
        TestString         string `json:"test_string"`                    // Example: "test1"
        TestOptionalString string `json:"test_optional_string,omitempty"` // Example: "test1"
}
```

### Inline struct definitions (`-i`)

**Input:**

```json
{
  "test_struct": {
    "test_string": "test",
    "test_array_of_string": ["test1", "test2", "test3"]
  },
  "test_array_of_struct": [
    {
      "test_string": "test1",
      "test_optional_string": "test1"
    },
    {
      "test_string": "test2",
      "test_optional_string": "test2"
    },
    {
      "test_string": "test3"
    },
    {
      "test_string": "test4"
    }
  ]
}
```

**Output:**

```golang
type Stdin1 struct {
        TestStruct struct {
                TestString        string   `json:"test_string"`
                TestArrayOfString []string `json:"test_array_of_string"`
        } `json:"test_struct"`
        TestArrayOfStruct []struct {
                TestString         string `json:"test_string"`
                TestOptionalString string `json:"test_optional_string,omitempty"`
        } `json:"test_array_of_struct"`
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
