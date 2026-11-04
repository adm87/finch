# finch

A small 2D game engine in Go, built on [Ebitengine](https://ebitengine.org) for the window, drawing and input. The repo also has a test game that runs on it.

## Running

```bash
go run ./cmd/finch-game
```

Press 1 to toggle the vignette, 2 to toggle the CRT effect, and Esc to quit.

## Layout

- `pkg/` is the engine core:
  - `engine` runs the game loop: the update phases (update, fixed update, later update), drawing and layout. A frame is drawn by a pipeline of draw passes.
  - `display` holds the screen buffer the game draws into, scaled to fill the window.
  - `ecs` holds the world. Under it, `transform` places entities, `render` sorts them by layer and depth and hands each to an adapter, and `sprite` is the adapter for images.
  - `geom` has the geometry types.
- `cmd/finch-game` is the test game. Its post-processing effects live in `internal/game/effects`.
- `content/embedded` holds the files built into the game, including the engine and game configs in `configs/`.
- `.dagger/` holds the Dagger module that runs the checks in Linux containers.

## Configs

`engine.yaml` sets the window and the clock: `tick_rate` is updates per second and `fixed_rate` is fixed steps per second. `game.yaml` sets the screen buffer size and its resolution scale. Keys left out keep their defaults, and unknown keys are an error.

## Development

The Makefile wraps the common commands.

| Command | Does |
|---|---|
| `make test` | Runs the unit tests |
| `make lint` | Runs `betteralign -apply`, then `golangci-lint` |
| `make cover` | Writes a coverage report for `pkg/` and opens it in the browser |
| `make bench` | Runs the benchmarks with allocation counts |
| `make test-integration` | Runs the integration tests locally (they open a window, so use the Dagger check on macOS) |
| `make check` | Runs every Dagger check silently: tests, lint, generated-code checks and the integration tests |
| `make check-integration` | Runs only the integration check, with output |
| `make check-verbose` | Runs every Dagger check with plain-text output |

Integration tests use the `integration` build tag. The Dagger checks need [Dagger](https://dagger.io) and a running Docker engine, and run the integration tests on a virtual display (Xvfb).

## License

Apache 2.0. See [LICENSE](LICENSE).
