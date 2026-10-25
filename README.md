# finch

A small 2D game engine in Go, built on [Ebiten](https://ebitengine.org) for the window, drawing and input, and [Donburi](https://github.com/yohamta/donburi) for entities and components. The repo also has a test game that runs on it.

## Running

Run from the repo root. The game looks for `content/assets` relative to where it's started.

```bash
go run ./cmd/game
```

For debug logging with file and line numbers, use the dev build:

```bash
go run -tags finch_dev ./cmd/game
```

## Controls

| Key | Does |
|---|---|
| W A S D | Move |
| Up / Down | Zoom |
| 1 | Show entity bounds |
| 2 | Show FPS |
| 3 | Show partition cells |
| 4 | Show the world origin |
| Esc | Quit |

## Layout

- `pkg/` is the engine core:
  - `application` holds the assets, caches, screen, view, world, render pipeline and plugins.
  - `engine` runs the game loop, the update phases and drawing.
  - `world` is the entity world with a spatial index.
  - The rest are smaller building blocks: `assets`, `images`, `data`, `rendering`, `transform`, `display`, `hashgrid`, `slotmap` and `errs`.
- `plugins/` connects third-party tools:
  - `aseprite` handles sprite sheets and animation.
  - `tiled` is for maps, and is still in progress.

  Plugins can use `pkg/`, but `pkg/` never imports `plugins/`. A test checks that.
- `cmd/game` is the test game.
- `content/` holds the game's assets and its embedded config files.

## Development

```bash
go test ./...
```

```bash
golangci-lint run ./...
```

```bash
betteralign -apply ./...
```

## License

Apache 2.0. See [LICENSE](LICENSE).
