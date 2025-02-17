# Must and Panic

Tired of typing `if a, err := someFunction(); err != nil`?

```go
package main

import (
	"context"

	. "github.com/nguyengg/golambda/must"
)

func main() {
	// should be dot imported for convenience.
	a := Must(someFunction())

	// variants with more return values with error being the last.
	Must0(functionThatReturnsError())
	a, b := Must2(ab())
	a, b, c := Must3(abc())

	// if you need more, you should really rethink your function.
}
```
