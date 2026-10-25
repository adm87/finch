package assets_test

import (
	"errors"
	"io"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

type fakeImporter struct {
	next      assets.Handle
	contents  map[assets.Handle]string
	imported  []assets.Asset
	fsys      fs.FS
	deleted   []assets.Handle
	importErr error
}

func newFakeImporter() *fakeImporter {
	return &fakeImporter{contents: make(map[assets.Handle]string)}
}

func (f *fakeImporter) Import(asset assets.Asset, fsys fs.FS, reader io.Reader) (assets.Handle, error) {
	f.imported = append(f.imported, asset)
	f.fsys = fsys
	if f.importErr != nil {
		return 0, f.importErr
	}
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	f.next++
	f.contents[f.next] = string(raw)
	return f.next, nil
}

func (f *fakeImporter) Delete(handle assets.Handle) error {
	f.deleted = append(f.deleted, handle)
	delete(f.contents, handle)
	return nil
}

var testFS = fstest.MapFS{
	"a.txt":     {Data: []byte("alpha")},
	"b.txt":     {Data: []byte("beta")},
	"other.bin": {Data: []byte{1, 2, 3}},
}

func newAssets(t *testing.T) (*assets.Assets, *fakeImporter) {
	t.Helper()
	as := assets.NewAssets()
	imp := newFakeImporter()
	require.NoError(t, as.RegisterImporter(imp, []string{"txt"}))
	return as, imp
}

func TestLoadAndHandle(t *testing.T) {
	as, imp := newAssets(t)

	require.NoError(t, as.Load(testFS, "a.txt", "b.txt"))

	ha, err := as.Handle("a.txt")
	require.NoError(t, err)
	hb, err := as.Handle("b.txt")
	require.NoError(t, err)
	require.Equal(t, "alpha", imp.contents[ha])
	require.Equal(t, "beta", imp.contents[hb])
}

func TestLoadDuplicatesInOneBatchLoadOnce(t *testing.T) {
	as, imp := newAssets(t)

	require.NoError(t, as.Load(testFS, "a.txt", "a.txt"))
	require.Len(t, imp.contents, 1)
}

func TestLoadTwiceIsAlreadyExists(t *testing.T) {
	as, imp := newAssets(t)
	require.NoError(t, as.Load(testFS, "a.txt"))

	err := as.Load(testFS, "a.txt")
	require.True(t, errs.IsAlreadyExists(err))
	require.Len(t, imp.contents, 1)
}

func TestLoadErrors(t *testing.T) {
	cases := []struct {
		name  string
		asset assets.Asset
		check func(error) bool
	}{
		{"missing file", "missing.txt", func(err error) bool { return errs.IsNotFound(err) && errors.Is(err, fs.ErrNotExist) }},
		{"no importer for type", "other.bin", errs.IsNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			as, _ := newAssets(t)

			err := as.Load(testFS, tc.asset)
			require.Error(t, err)
			require.True(t, tc.check(err), err.Error())

			_, err = as.Handle(tc.asset)
			require.True(t, errs.IsNotFound(err))
		})
	}
}

func TestLoadFailureDoesNotStopTheBatch(t *testing.T) {
	as, _ := newAssets(t)

	err := as.Load(testFS, "missing.txt", "a.txt")
	require.True(t, errs.IsNotFound(err))

	_, err = as.Handle("a.txt")
	require.NoError(t, err)
}

func TestImportFailureLeavesAssetUnloaded(t *testing.T) {
	as, imp := newAssets(t)
	imp.importErr = errors.New("bad data")

	err := as.Load(testFS, "a.txt")
	require.ErrorIs(t, err, imp.importErr)

	_, err = as.Handle("a.txt")
	require.True(t, errs.IsNotFound(err))
}

func TestLoadNilFileSystem(t *testing.T) {
	as, _ := newAssets(t)
	require.ErrorIs(t, as.Load(nil, "a.txt"), errs.ErrArgumentNil)
}

func TestUnload(t *testing.T) {
	as, imp := newAssets(t)
	require.NoError(t, as.Load(testFS, "a.txt"))
	h, err := as.Handle("a.txt")
	require.NoError(t, err)

	require.NoError(t, as.Unload("a.txt"))
	require.Equal(t, []assets.Handle{h}, imp.deleted)

	_, err = as.Handle("a.txt")
	require.True(t, errs.IsNotFound(err))

	require.True(t, errs.IsNotFound(as.Unload("a.txt")))
	require.NoError(t, as.Load(testFS, "a.txt"))
}

func TestRegisterImporter(t *testing.T) {
	t.Run("normalizes types", func(t *testing.T) {
		as := assets.NewAssets()
		imp := newFakeImporter()
		require.NoError(t, as.RegisterImporter(imp, []string{".TXT"}))
		require.NoError(t, as.Load(testFS, "a.txt"))
	})

	t.Run("rejects a type that is already registered", func(t *testing.T) {
		as, _ := newAssets(t)
		err := as.RegisterImporter(newFakeImporter(), []string{"bin", "txt"})
		require.True(t, errs.IsAlreadyExists(err))
		require.True(t, errs.IsNotFound(as.Load(testFS, "other.bin")))
	})

	t.Run("rejects a type listed twice", func(t *testing.T) {
		as := assets.NewAssets()
		err := as.RegisterImporter(newFakeImporter(), []string{"txt", ".txt"})
		require.True(t, errs.IsAlreadyExists(err))
	})

	t.Run("rejects nil", func(t *testing.T) {
		as := assets.NewAssets()
		require.ErrorIs(t, as.RegisterImporter(nil, []string{"txt"}), errs.ErrArgumentNil)
	})
}

func TestAssetType(t *testing.T) {
	require.Equal(t, "png", assets.Asset("sprites/Hero.PNG").Type())
	require.Equal(t, "json", assets.Asset("a.b.json").Type())
	require.Empty(t, assets.Asset("README").Type())
}

func TestAssetBundle(t *testing.T) {
	as, _ := newAssets(t)
	require.NoError(t, as.RegisterFileSystem("test", testFS))

	bundle := assets.AssetBundle{FileSystem: "test", Assets: []assets.Asset{"a.txt", "b.txt"}}
	require.NoError(t, bundle.Load(as))
	_, err := as.Handle("b.txt")
	require.NoError(t, err)

	require.NoError(t, bundle.Unload(as))
	_, err = as.Handle("b.txt")
	require.True(t, errs.IsNotFound(err))

	missing := assets.AssetBundle{FileSystem: "nope", Assets: []assets.Asset{"a.txt"}}
	require.True(t, errs.IsNotFound(missing.Load(as)))
}

func TestRegisterFileSystem(t *testing.T) {
	as := assets.NewAssets()
	require.NoError(t, as.RegisterFileSystem("test", testFS))
	require.True(t, errs.IsAlreadyExists(as.RegisterFileSystem("test", testFS)))
	require.ErrorIs(t, as.RegisterFileSystem("other", nil), errs.ErrArgumentNil)

	fsys, err := as.FileSystem("test")
	require.NoError(t, err)
	require.NotNil(t, fsys)
}

func TestImportReceivesTheAssetAndFileSystem(t *testing.T) {
	as, imp := newAssets(t)

	require.NoError(t, as.Load(testFS, "a.txt", "b.txt"))
	require.ElementsMatch(t, []assets.Asset{"a.txt", "b.txt"}, imp.imported)
	require.Equal(t, testFS, imp.fsys)
}

// refImporter loads the asset its file names, relative to the file itself, the
// way a Tiled map loads its tilesets.
type refImporter struct {
	as   *assets.Assets
	next assets.Handle
}

func (r *refImporter) Import(asset assets.Asset, fsys fs.FS, reader io.Reader) (assets.Handle, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	ref := assets.Asset(path.Join(path.Dir(asset.String()), string(raw)))
	if err := r.as.Load(fsys, ref); err != nil {
		return 0, err
	}
	r.next++
	return r.next, nil
}

func (r *refImporter) Delete(assets.Handle) error {
	return nil
}

func TestImporterCanLoadAssetsItRefersTo(t *testing.T) {
	as, imp := newAssets(t)
	require.NoError(t, as.RegisterImporter(&refImporter{as: as}, []string{"map"}))

	fsys := fstest.MapFS{
		"maps/level.map":  {Data: []byte("../tiles/grass.txt")},
		"tiles/grass.txt": {Data: []byte("grass")},
	}
	require.NoError(t, as.Load(fsys, "maps/level.map"))

	_, err := as.Handle("maps/level.map")
	require.NoError(t, err)
	tile, err := as.Handle("tiles/grass.txt")
	require.NoError(t, err)
	require.Equal(t, "grass", imp.contents[tile])
}
