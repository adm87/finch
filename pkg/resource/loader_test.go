package resource_test

import (
	"errors"
	"io/fs"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/resource"
)

func TestLoaderReadsAcrossSources(t *testing.T) {
	res := resource.New()
	require.NoError(t, res.Mount("disk", fstest.MapFS{"a.txt": {Data: []byte("disk a")}}))
	require.NoError(t, res.Mount("cdn", fstest.MapFS{"a.txt": {Data: []byte("cdn a")}}))

	l := resource.NewLoader(res, resource.Bytes)
	l.Queue("disk", "a.txt", "a.txt")
	l.Queue("cdn", "a.txt")
	results := l.Load()

	require.NoError(t, resource.Err(results))
	require.Len(t, results, 2)
	require.Equal(t, "disk a", string(results[0].Value))
	require.Equal(t, resource.Source("cdn"), results[1].Source)
	require.Equal(t, "cdn a", string(results[1].Value))
	require.Empty(t, l.Load())
}

func TestLoaderKeepsGoingPastFailures(t *testing.T) {
	res := resource.New()
	require.NoError(t, res.Mount("disk", fstest.MapFS{
		"one.txt": {Data: []byte("1")},
		"bad.txt": {Data: []byte("x")},
	}))

	l := resource.NewLoader(res, func(raw []byte) (int, error) { return strconv.Atoi(string(raw)) })
	l.Queue("disk", "missing.txt", "bad.txt", "one.txt")
	l.Queue("nowhere", "one.txt")
	results := l.Load()

	require.Len(t, results, 4)
	require.ErrorIs(t, results[0].Err, fs.ErrNotExist)
	require.ErrorContains(t, results[1].Err, "decode disk:bad.txt")
	require.NoError(t, results[2].Err)
	require.Equal(t, 1, results[2].Value)
	require.True(t, errs.IsNotFound(results[3].Err))

	err := resource.Err(results)
	require.ErrorIs(t, err, fs.ErrNotExist)
	var numErr *strconv.NumError
	require.True(t, errors.As(err, &numErr))
}
