![🚧 Under construction 👷‍♂️](https://i.imgur.com/LEP2R3N.png)

# Fetch API for Go

📥 [The WHATWG Fetch API](https://fetch.spec.whatwg.org/) for Go

<table align=center><td>

```go
type Todo struct {
  UserID int
  ID int
  Title string
  Completed bool
}
response, err := fetch.Fetch("https://jsonplaceholder.typicode.com/todos/1")
if err != nil {
  log.Fatal(err)
}
var todo Todo
err = response.JSONInto(&todo)
if err != nil {
  log.Fatal(err)
}
log.Println("todo=%#+v", todo)
// Output: ...
```

</table>

## Installation

You can install this package using `go get` right from your command line:

```sh
go get github.com/jcbhmr/go-fetch
```

Or if you prefer you can import it in your Go code and use `go mod tidy` to automagically ✨ add it to your `go.mod`.

```go
import "github.com/jcbhmr/go-fetch"
```

## Usage

```go
package main

import (
  "log"

  "github.com/jcbhmr/go-fetch"
)

func main() {
  response, err := fetch.Fetch("https://example.com/")
  if err != nil {
    log.Fatal(err)
  }
  text, err := response.Text()
  if err != nil {
    log.Fatal(err)
  }
  log.Println("%s returned this %s:", response.URL(), response.Headers().Get("Content-Type"))
  log.Println(text)
}
```

## Development

WebIDL's primary language integration is between C++ and JavaScript but a lot of the constructs remain relatively portable. In this case, though, we are implementing & consuming the Fetch API using Go! That means there's some things to be aware of about how the WebIDL-defined API is translated into somewhat idiomatic Go code.

1. Names are transformed into Go-conformant PascalCase names to be properly exported.
2. `optional` parameters become `nil`-able `*T` types if the original type is un-`nil`-able (like `string`, `int`, etc.). If the original type **is `nil`-able** then it stays as-is (like a `*MyStruct` parameter).
3. Sum types that can't be shoehorned in using sealed `interface` hacks (like `MyStruct|string` since you can't do interfaces on primitives) are defined as `any` and annotated with a developer note.
4. `DOMString` and `USVString` are both represented using the Go `string` primitive. They are mostly considered interchangable. This might be changed later. 🤷‍♂️
5. `ByteString` is also just a Go `string`.
6. JavaScript-native `Promise<T>` has been replaced with a `<-chan Result[T]`
7. `BufferSource` where possible is a `[]byte|[]uint8|[]int8|[]uint16|...` union
8. `ArrayBuffer` is considered to be equivalent to `[]byte`
9. Go versions of WebIDL enums rely on the user to know the possible enum values. So `type RequestCredentials = string`.
10. All WebIDL-defined properties are exposed through getter and setter methods. There is no prefix for getters and a `Set` prefix for setters. For instance `myStruct.Value()` and `myStruct.SetValue()` would be the getter/setter for `MyStruct#value`.
11. This package **does expose** non-`[Exposed=...]` WebIDL types that are not available to JavaScript.
12. WebIDL `dictionary` structs always have all their fields made `nil`-able similar to `optional` parameters unless they are marked `required`.
13. `iterable<K,V>` means that it has an `.Iter()` method that returns a `Seq2[K,V]` function which can be used with https://go.dev/wiki/RangefuncExperiment
14. `iterable<V>` means that it has an `.Iter()` method that returns a `Seq[V]` function which can be used with https://go.dev/wiki/RangefuncExperiment
15. In places where a `TypeError` would be thrown it will `panic()`
16. In places where a `DOMException` would be thrown it will return an `error`
17. `sequence<T>` is taken to be equivalent to `[]T`
