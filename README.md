# jsonedit

---

Edit JSON documents by path without reformatting them.

```go
src := []byte(`{"name": "alice", "tags": ["a", "b"]}`)

out, err := jsonedit.Set(src, "$.name", `"bob"`, "", false)
// {"name": "bob", "tags": ["a", "b"]}

out, err = jsonedit.Append(out, "$.tags[]", `"c"`)
// {"name": "bob", "tags": ["a", "b", "c"]}

out, err = jsonedit.Delete(out, "$.name", "", false)
// {"tags": ["a", "b", "c"]}
```

Every function takes bytes in, returns bytes out. The document is never round-tripped through `json.Unmarshal`
and `json.Marshal`: the range of the value at the path is spliced, and everything outside it is preserved byte-for-byte.
Formatting survives.

## API

| Function | Purpose |
|---|---|
| `Set(src, path, rawValue, expect, hasExpect)` | replace the value at path |
| `Delete(src, path, expect, hasExpect)` | remove the value at path |
| `Insert(src, path, rawValue)` | add a value; a key that exists is an error |
| `Append(src, path, rawValue)` | add to the end of an array |
| `Rename(src, path, newKey, expect, hasExpect)` | change a key |
| `Locate(src, path)` | byte range of the value at path |
| `Validate(src)` | is this valid JSON? |

`path` is a JSONPath subset: `$`, `.key`, `[N]`, `[]`. No wildcards, no filters, no recursion.

`rawValue` is a JSON literal as a string: `"x"`, `42`, `{"a": 1}`, `[1, 2]`. The empty string is not a value.

When `hasExpect` is true, the value at `path` must equal `expect` or the edit is refused and the source is unchanged.

## Install

```
go get github.com/avalak/jsonedit
```

Requires Go 1.18 or later.

## Relationship to other tools
- `sjson`, `gjson` — set and get values by path, fast. SJSON does not
  validate the input; jsonedit requires valid JSON and adapts the
  file's own indent and colon style for new members.
- `jsonpatch` — RFC 6902, a wire format for diffs. `jsonedit` is a library, not a format.
- `encoding/json` — parses and re-emits, reformatting the whole document. `jsonedit` touches the target range only.

## License

See `LICENSE`.
