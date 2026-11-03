package main

import (
	"context"
	"strings"

	"dagger/finch-dev/internal/dagger"
)

func base() *dagger.Container {
	return dag.Container().
		From("golang:1.27.1").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build"))
}

func virtualDisplay(container *dagger.Container) *dagger.Container {
	packages := []string{
		"xvfb", "xauth", "pkg-config", "libgl1-mesa-dev", "libgl1-mesa-dri",
		"libxcursor-dev", "libxi-dev", "libxinerama-dev", "libxrandr-dev", "libxxf86vm-dev", "libasound2-dev",
	}
	install := "apt-get update && apt-get install -y --no-install-recommends " + strings.Join(packages, " ")
	return container.WithExec([]string{"sh", "-c", install})
}

type FinchDev struct {
	// +private
	Source *dagger.Directory
}

func New(ws *dagger.Workspace) *FinchDev {
	return &FinchDev{
		Source: ws.Directory("/", dagger.WorkspaceDirectoryOpts{
			Exclude: []string{
				"**/.git",
				"**/.dagger",
				"**/.idea",
			},
		}),
	}
}

// Integration runs the integration tests on a virtual display.
//
// +check
func (f *FinchDev) Integration(ctx context.Context) error {
	_, err := base().With(virtualDisplay).
		WithDirectory("/src", f.Source).
		WithWorkdir("/src").
		WithExec([]string{"xvfb-run", "-a", "go", "test", "-tags", "integration", "-run", "^TestIntegration", "./..."}).
		Sync(ctx)
	return err
}
