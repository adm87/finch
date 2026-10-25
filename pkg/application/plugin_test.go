package application_test

import (
	"errors"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
)

type fakePlugin struct {
	name  string
	err   error
	inits int
}

func (p *fakePlugin) Name() string { return p.name }

func (p *fakePlugin) Init(*application.Application) error {
	p.inits++
	return p.err
}

type otherPlugin struct{ fakePlugin }

func newPluginApp(t *testing.T) *application.Application {
	t.Helper()
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)
	return app
}

func TestRegisterPluginInitializesItOnce(t *testing.T) {
	app := newPluginApp(t)
	plugin := &fakePlugin{name: "a"}

	require.NoError(t, app.RegisterPlugin(plugin))
	require.Equal(t, 1, plugin.inits)

	got, err := application.GetPlugin[*fakePlugin](app, "a")
	require.NoError(t, err)
	require.Same(t, plugin, got)
}

func TestRegisterPluginRejectsDuplicateNames(t *testing.T) {
	app := newPluginApp(t)
	require.NoError(t, app.RegisterPlugin(&fakePlugin{name: "a"}))

	second := &fakePlugin{name: "a"}
	require.True(t, errs.IsAlreadyExists(app.RegisterPlugin(second)))
	require.Zero(t, second.inits)
}

func TestRegisterPluginLeavesFailedPluginsOut(t *testing.T) {
	app := newPluginApp(t)
	broken := errors.New("broken")

	require.ErrorIs(t, app.RegisterPlugin(&fakePlugin{name: "a", err: broken}), broken)
	_, err := application.GetPlugin[*fakePlugin](app, "a")
	require.True(t, errs.IsNotFound(err))

	require.NoError(t, app.RegisterPlugin(&fakePlugin{name: "a"}))
}

func TestGetPluginErrors(t *testing.T) {
	app := newPluginApp(t)
	require.NoError(t, app.RegisterPlugin(&otherPlugin{fakePlugin{name: "other"}}))

	_, err := application.GetPlugin[*fakePlugin](app, "missing")
	require.True(t, errs.IsNotFound(err))

	_, err = application.GetPlugin[*fakePlugin](app, "other")
	require.ErrorIs(t, err, errs.ErrInvalidOperation)
}
