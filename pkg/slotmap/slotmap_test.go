package slotmap_test

import (
	"math"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/slotmap"
)

func TestSlotMap(t *testing.T) {
	t.Run("Should initialize with correct length and capacity", func(t *testing.T) {
		slotMap := slotmap.New[int](10)
		require.NotNil(t, slotMap)
		require.Equal(t, 0, slotMap.Len())
		require.Equal(t, 10, slotMap.Cap())
	})

	t.Run("Should clamp negative seed to zero", func(t *testing.T) {
		slotMap := slotmap.New[int](-5)
		require.NotNil(t, slotMap)
		require.Equal(t, 0, slotMap.Len())
		require.Equal(t, 0, slotMap.Cap())

		k := slotMap.Insert(7)
		v, err := slotMap.Get(k)
		require.NoError(t, err)
		require.Equal(t, 7, v)
	})

	t.Run("Should panic when seed exceeds uint32 index capacity", func(t *testing.T) {
		require.Panics(t, func() {
			_ = slotmap.New[int](math.MaxUint32 + 1)
		})
	})

	t.Run("Should insert and retrieve a value correctly", func(t *testing.T) {
		slotMap := slotmap.New[int](2)

		k := slotMap.Insert(42)
		v, err := slotMap.Get(k)

		require.NoError(t, err)
		require.Equal(t, 42, v)
	})

	t.Run("Should keep free list valid after grow", func(t *testing.T) {
		slotMap := slotmap.New[int](1)

		k := slotMap.Insert(10)
		_, err := slotMap.Delete(k)
		require.NoError(t, err)

		slotMap.Grow(2)

		k2 := slotMap.Insert(20)
		k3 := slotMap.Insert(30)
		k4 := slotMap.Insert(40)

		v2, err := slotMap.Get(k2)
		require.NoError(t, err)
		require.Equal(t, 20, v2)

		v3, err := slotMap.Get(k3)
		require.NoError(t, err)
		require.Equal(t, 30, v3)

		v4, err := slotMap.Get(k4)
		require.NoError(t, err)
		require.Equal(t, 40, v4)

		require.Equal(t, 3, slotMap.Len())
	})

	t.Run("Should delete value and invalidate stale key", func(t *testing.T) {
		slotMap := slotmap.New[int](2)

		k := slotMap.Insert(99)
		require.Equal(t, 1, slotMap.Len())

		deleted, err := slotMap.Delete(k)
		require.NoError(t, err)
		require.Equal(t, 99, deleted)
		require.Equal(t, 0, slotMap.Len())

		_, err = slotMap.Get(k)
		require.ErrorIs(t, err, slotmap.ErrNotFound)

		err = slotMap.Set(k, 100)
		require.ErrorIs(t, err, slotmap.ErrNotFound)

		_, err = slotMap.Delete(k)
		require.ErrorIs(t, err, slotmap.ErrNotFound)
	})

	t.Run("Should return false for invalid keys", func(t *testing.T) {
		slotMap := slotmap.New[int](1)

		k := slotmap.K{}

		_, err := slotMap.Get(k)
		require.ErrorIs(t, err, slotmap.ErrInvalidKey)

		err = slotMap.Set(k, 1)
		require.ErrorIs(t, err, slotmap.ErrInvalidKey)

		_, err = slotMap.Delete(k)
		require.ErrorIs(t, err, slotmap.ErrInvalidKey)
	})

	t.Run("Should range over occupied entries and support early stop", func(t *testing.T) {
		slotMap := slotmap.New[int](4)

		k1 := slotMap.Insert(10)
		slotMap.Insert(20)
		k3 := slotMap.Insert(30)

		_, err := slotMap.Delete(k1)
		require.NoError(t, err)

		seen := map[int]int{}
		slotMap.Range(func(_ slotmap.K, value int) bool {
			seen[value]++
			return true
		})

		require.Equal(t, 0, seen[10])
		require.Equal(t, 1, seen[20])
		require.Equal(t, 1, seen[30])

		visited := 0
		slotMap.Range(func(k slotmap.K, _ int) bool {
			visited++
			return k != k3
		})
		require.True(t, visited >= 1)
		require.True(t, visited <= 2)
	})
}

func BenchmarkSlotMap(b *testing.B) {
	b.Run("Insert", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			b.Run("Growth", func(b *testing.B) {
				m := make(map[int]int)
				b.ReportAllocs()
				b.ResetTimer()
				for i := range b.N {
					m[i] = i
				}
			})
			b.Run("Preallocated", func(b *testing.B) {
				m := make(map[int]int, b.N)
				b.ReportAllocs()
				b.ResetTimer()
				for i := range b.N {
					m[i] = i
				}
			})
		})
		b.Run("slotmap", func(b *testing.B) {
			b.Run("Growth", func(b *testing.B) {
				sm := slotmap.New[int](1)
				b.ReportAllocs()
				b.ResetTimer()
				for i := range b.N {
					_ = sm.Insert(i)
				}
			})
			b.Run("Preallocated", func(b *testing.B) {
				sm := slotmap.New[int](b.N)
				b.ReportAllocs()
				b.ResetTimer()
				for i := range b.N {
					_ = sm.Insert(i)
				}
			})
		})
	})

	b.Run("Get", func(b *testing.B) {
		const size = 10_000

		b.Run("map", func(b *testing.B) {
			m, keys := benchmarkFilledMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				_ = m[keys[i%size]]
			}
		})
		b.Run("slotmap", func(b *testing.B) {
			sm, keys := benchmarkFilledSlotMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				_, _ = sm.Get(keys[i%size])
			}
		})
	})

	b.Run("Delete", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			m, keys := benchmarkFilledMap(b.N)
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				delete(m, keys[i])
			}
		})
		b.Run("slotmap", func(b *testing.B) {
			sm, keys := benchmarkFilledSlotMap(b.N)
			b.ReportAllocs()
			b.ResetTimer()
			for i := range b.N {
				_, _ = sm.Delete(keys[i])
			}
		})
	})

	b.Run("RangeFull", func(b *testing.B) {
		const size = 10_000

		b.Run("map", func(b *testing.B) {
			m, _ := benchmarkFilledMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for range b.N {
				for _, v := range m {
					sum += v
				}
			}
			_ = sum
		})
		b.Run("slotmap", func(b *testing.B) {
			sm, _ := benchmarkFilledSlotMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for range b.N {
				sm.Range(func(k slotmap.K, v int) bool {
					sum += v
					return true
				})
			}
			_ = sum
		})
	})

	b.Run("RangeEarlyExit", func(b *testing.B) {
		const size = 10_000

		b.Run("map", func(b *testing.B) {
			m, _ := benchmarkFilledMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for range b.N {
				for _, v := range m {
					sum += v
					break
				}
			}
			_ = sum
		})
		b.Run("slotmap", func(b *testing.B) {
			sm, _ := benchmarkFilledSlotMap(size)
			b.ReportAllocs()
			b.ResetTimer()
			sum := 0
			for range b.N {
				sm.Range(func(k slotmap.K, v int) bool {
					sum += v
					return false
				})
			}
			if sum == -1 { // prevents dead-code elimination of `sum` without changing behavior
				b.Fatal("unreachable")
			}
		})
	})
}

func benchmarkFilledSlotMap(size int) (*slotmap.SlotMap[int], []slotmap.K) {
	sm := slotmap.New[int](1)
	keys := make([]slotmap.K, size)
	for i := range size {
		keys[i] = sm.Insert(i)
	}
	return sm, keys
}

func benchmarkFilledMap(size int) (map[int]int, []int) {
	m := make(map[int]int, size)
	keys := make([]int, size)
	for i := range size {
		m[i] = i
		keys[i] = i
	}
	return m, keys
}
