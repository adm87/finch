# finch

A small 2D game engine in Go, built on [Ebitengine](https://ebitengine.org) for the window, drawing and input. The repo also has a test game that runs on it.

## Running

```bash
go run ./cmd/finch-game
```

Press Esc to quit.

## Layout

- `pkg/` is the engine core:
  - `engine` runs the game loop: the update phases (update, fixed update, later update), drawing and layout. A frame is drawn by a pipeline of draw passes.
  - `display` holds the screen buffer the game draws into, scaled to fill the window.
  - `geom` has the geometry types.
- `cmd/finch-game` is the test game.
- `.dagger/` holds the Dagger module that runs the checks in Linux containers.

## Development

The Makefile wraps the common commands.

| Command | Does |
|---|---|
| `make test` | Runs the unit tests |
| `make lint` | Runs `betteralign -apply`, then `golangci-lint` |
| `make cover` | Writes a coverage report for `pkg/` and opens it in the browser |
| `make test-integration` | Runs the integration tests locally (they open a window) |
| `make check` | Runs every Dagger check silently: tests, lint, generated-code checks and the integration tests |
| `make check-integration` | Runs only the integration check, with output |
| `make check-verbose` | Runs every Dagger check with plain-text output |

Integration tests use the `integration` build tag. The Dagger checks need [Dagger](https://dagger.io) and a running Docker engine, and run the integration tests on a virtual display (Xvfb).

## License

Apache 2.0. See [LICENSE](LICENSE).
