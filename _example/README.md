# Example of khanerr.errors

```
$ go run -trimpath main.go
(1) Fields: [Kind:internal error,Message:root,bar:true,kind:internal error], Cause: internal error: Fields: [Kind:internal error,Message:root,bar:true], Cause: internal error: Fields: [Kind:internal error,Message:root], Cause: internal error: Something went wrong
  -- Stack trace:main.baz
  | 	./main.go:32
  | main.main
  | 	./main.go:36
Wraps: (2) internal error: Fields: [Kind:internal error,Message:root,bar:true], Cause: internal error: Fields: [Kind:internal error,Message:root], Cause: internal error: Something went wrong
Wraps: (3) Fields: [Kind:internal error,Message:root,bar:true], Cause: internal error: Fields: [Kind:internal error,Message:root], Cause: internal error: Something went wrong
  -- Stack trace:
  | [...repeated from below...]
Wraps: (4) internal error: Fields: [Kind:internal error,Message:root], Cause: internal error: Something went wrong
Wraps: (5) Fields: [Kind:internal error,Message:root], Cause: internal error: Something went wrong
  -- Stack trace:main.foo
  | 	./main.go:20
  | main.bar
  | 	./main.go:26
  | main.baz
  | 	./main.go:31
  | main.main
  | 	./main.go:36
  | runtime.main
  | 	runtime/proc.go:250
Wraps: (6) internal error: Something went wrong
Wraps: (7) Something went wrong
Error types: (1) *errors.withFields (2) *errors.wrapper (3) *errors.withFields (4) *errors.wrapper (5) *errors.withFields (6) *errors.wrapper (7) main.ErrMyError
  -- Stack trace:main.baz
  | 	./main.go:32
  | main.main
  | 	./main.go:36
```