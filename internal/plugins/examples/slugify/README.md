# Example WASM plugin — `slugify`

A tiny, dependency-free WebAssembly plugin that showcases Fieldstone's WASM
plugin runtime (`internal/plugins/wasm.go`). It turns a string into a URL-safe
slug — exactly the kind of `beforeCreate` transform a BaaS needs:

```
"Shipping Realtime to Production!"  ->  "shipping-realtime-to-production"
```

Written in Rust, compiled to `wasm32-unknown-unknown` (~800 bytes), it implements
the host ABI verbatim: it exports `memory`, `malloc`, and `slugify`, reads the
input from linear memory, allocates an output buffer, and writes the result
pointer + length back into the slots the host provides.

## Files
- `src/lib.rs` — the plugin (`no_std`, bump allocator, slug logic).
- `build.sh` — `rustc … --target wasm32-unknown-unknown --crate-type cdylib`.
- `slugify.wasm` — prebuilt artifact (checked in so tests run without Rust).
- `harness/` — standalone Go runner (its own `go.mod`) that executes the wasm
  against mock data using the **same ABI** as the server. Runs on older Go too.

## Build
```bash
rustup target add wasm32-unknown-unknown   # once
./build.sh                                  # -> slugify.wasm
# or from repo root:
make plugin-build
```

## Test it (three convenient ways)

**1. Standalone harness (runs anywhere, no project build):**
```bash
cd harness && go run .
# or: make plugin-demo
```
Output:
```
"Shipping Realtime to Production!"  -> "shipping-realtime-to-production" [ok]
"  Hello, World  "                  -> "hello-world"                    [ok]
...
6/6 passed
```

**2. Through the real plugin runtime (Go 1.24 + `go mod tidy`):**
```bash
go mod tidy                 # pulls github.com/tetratelabs/wazero
make plugin-test            # go test ./internal/plugins -run TestSlugify -v
```

**3. From Go directly** — `plugin.Execute("slugify", []byte("My Title"))` returns
`"my-title"`. See `internal/plugins/slugify_test.go`.

## How it maps to a real hook
The runtime resolves collection hooks as exported functions named
`<collection>_<hook>` (e.g. `posts_beforeCreate`) via `Manager.ExecuteHook`.
To wire slugify into posts, export an additional function that reads the record
JSON, slugifies the `title`, and returns the updated record — the slug logic here
is the reusable core. For this showcase we export the simpler `slugify(string) ->
string` entry point and drive it with `plugin.Execute`.

## Notes
- The plugin is sandboxed: no filesystem, no network, fixed linear memory — the
  security model the runtime advertises.
- ASCII-only by design (non-ASCII bytes become separators); extend `lower_alnum`
  for Unicode transliteration if needed.
