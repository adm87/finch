# finch

A small 2D game engine in Go, built on [Ebitengine](https://ebitengine.org) for the window, drawing and input. The repo also has a test game that runs on it.

## Running

```bash
go run ./cmd/finch-game
```

Press 1 to toggle the CRT effect and Esc to quit. The top-left corner shows FPS, CPU draw time and memory use.

## Layout

- `pkg/` is the engine core:
  - `engine` runs the game loop: the update phases (update, fixed update, later update), drawing and layout. A frame is drawn by a pipeline of draw passes.
  - `display` holds the screen buffer the game draws into, scaled to fill the window.
  - `ecs` holds the world. Under it, `transform` places entities, `render` has adapters turn entities into draw commands and draws them in layer and depth order, and `sprite` is the adapter for images.
  - `geom` has the geometry types.
  - `hexcolor` reads hex colours from config files.
  - `slotmap` stores values under keys that stop working once the value is deleted.
- `cmd/finch-game` is the test game. Its post-processing effects live in `internal/game/effects`.
- `content/embedded` holds the files built into the game, including the game config in `configs/game.yaml`.
- `.dagger/` holds the Dagger module that runs the checks in Linux containers.

## Configs

`game.yaml` has two sections. `engine` sets the window and the clock: `tick_rate` is updates per second and `fixed_rate` is fixed steps per second. `display` sets the screen buffer size, its resolution scale and `clear_color`, a hex colour written as `#RRGGBB` or `#RRGGBBAA`. Keys left out keep their defaults, and unknown keys are an error.

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
