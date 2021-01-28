# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
**BREAKING** The `dgraphql.NewPaginator` factory method's signature changed, it now expects a `dgraphql.CursorDecoder` interface to decode `after` and `before` keys.

The old behavior can be obtain back by using the `dgraphql.NewOpaqueProtoCursrorDecoder` like this:

```go
func QueryElements() ([]*Element, error) {
    ...
    pagignator := dgraphql.NewPaginator(nil, nil, afterCursor, beforeCursor, 0, func () proto.Message { return &cursorProto{} })
    ...
}
```

To

```go
var elementCursorDecoder = dgraphql.NewOpaqueProtoCursorDecoder(func() proto.Message {
    return &pbpkg.ElementCursorProto{}
})

func QueryElements() ([]*Element, error) {
    ...
    pagignator := dgraphql.NewPaginator(nil, nil, afterCursor, beforeCursor, 0, elementCursorDecoder)
    ...
}
```

## [v0.0.1] - 2020-06-23

### Fixed
* Fixed client-js not reconnecting on certain circumstances

### Changed
* Now uses new dauth `authenticator` module
* Now uses a custom ErrorLog on HTTP servers

## [pre-open-sourcing]

### Changed
* Default parameters for `--auth-plugin` and `--dmetering-plugin` changed to `null://`.
* License changed to Apache 2.0
