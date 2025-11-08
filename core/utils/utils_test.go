package utils_test

import (
	"sort"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/core/utils"
)

func TestSelector(t *testing.T) {
	t.Run("selector selects an element from given list", func(t *testing.T) {
		items := []string{"a", "b", "c", "d", "e", "f"}
		ans := utils.FirstThatMatch(
			func(item string) bool {
				return item == "c" || item == "d" || item == "e"
			},
			"error",
			items...,
		)
		assert.Equal(t, "c", ans)
	})
	t.Run("selector selects will return default value", func(t *testing.T) {
		items := []string{"a", "b", "c", "d", "e", "f"}
		ans := utils.FirstThatMatch(
			func(item string) bool {
				return item == "g"
			},
			"error",
			items...,
		)
		assert.Equal(t, "error", ans)
	})
}

func TestNonZeroSelector(t *testing.T) {
	t.Run("MayFirstNonZero", func(t *testing.T) {
		items := []int{0, 1, 2, 3, 4, 5}
		ans := utils.FirstNonZeroForced(items...)
		assert.Equal(t, 1, ans)
	})
	t.Run("MayFirstNon", func(t *testing.T) {
		items := []int{0, 1, 2, 3, 4, 5}
		ans, err := utils.FirstNonZero(items...)
		assert.NoError(t, err)
		assert.Equal(t, 1, ans)
	})
	t.Run("MayFirstNon", func(t *testing.T) {
		items := []int{0, 0, 0, 0, 0}
		ans, err := utils.FirstNonZero(items...)
		assert.Error(t, err)
		assert.Equal(t, 0, ans)
	})
}

func TestZipChannel(t *testing.T) {
	t.Run("zip channel test", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)
		go func() {
			for i := 0; i < 5; i++ {
				ch1 <- i
			}
			close(ch1)
		}()
		go func() {
			for i := 0; i < 5; i++ {
				ch2 <- i
			}
			close(ch2)
		}()
		zipped := utils.ZipChannels(ch1, ch2)
		ans := make([]int, 0)
		for val := range zipped {
			ans = append(ans, val)
		}
		sort.Ints(ans)
		assert.Equal(t, []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}, ans)
	})
}

func TestList(t *testing.T) {
	t.Run("list generator", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		ans := make([]int, 0)
		ans = append(ans, list.Slice()...)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, ans)
	})
	t.Run("list fold", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		ans := utils.Fold(list, 0, func(lastValue int, current int) int {
			return lastValue + current
		})
		assert.Equal(t, 15, ans)
	})
	t.Run("list map", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		ans := utils.Map(list, strconv.Itoa)
		assert.Equal(t, []string{"1", "2", "3", "4", "5"}, ans.Slice())
	})
	t.Run("list remove", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		list.Remove(3)
		assert.Equal(t, []int{1, 2, 4, 5}, list.Slice())
	})
	t.Run("list remove (non-present item)", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		list.Remove(10)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, list.Slice())
	})
	t.Run("list length", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		assert.Equal(t, 5, list.Len())
		list.Remove(3)
		assert.Equal(t, 4, list.Len())
	})
	t.Run("list all tester", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		pass := list.All(func(i int) bool {
			return i < 10
		})
		fail := list.All(func(i int) bool {
			return i < 3
		})
		assert.True(t, pass)
		assert.False(t, fail)
	})
	t.Run("list any tester", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		pass := list.Any(func(i int) bool {
			return i < 3
		})
		fail := list.Any(func(i int) bool {
			return i < 0
		})
		assert.True(t, pass)
		assert.False(t, fail)
	})
	t.Run("list contains tester", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		pass := list.Contains(3)
		fail := list.Contains(6)
		list.Remove(3, 4)
		fail2 := list.Contains(3)
		assert.True(t, pass)
		assert.False(t, fail)
		assert.False(t, fail2)
	})
	t.Run("list empty check", func(t *testing.T) {
		list := utils.NewList(1, 2, 3, 4, 5)
		assert.False(t, list.IsEmpty())
		assert.True(t, list.IsNotEmpty())
		list.Remove(1, 2, 3, 4, 5)
		assert.True(t, list.IsEmpty())
		assert.False(t, list.IsNotEmpty())
		list.Add(1, 2, 3, 4, 5)
		assert.False(t, list.IsEmpty())
		assert.True(t, list.IsNotEmpty())
	})
}

func TestEscapedSplit(t *testing.T) {
	t.Run("Panic when last rune is escape rune '\\'",
		func(t *testing.T) {
			str := "must-panic\\"
			assert.Panics(t, func() {
				utils.EscapedSplit(str, '-')
			})
		},
	)
	t.Run("Normal input returns slice with single item (whole input)",
		func(t *testing.T) {
			str := "nothing to split"
			result := utils.EscapedSplit(str, '-')
			assert.Equal(t, []string{str}, result)
		},
	)
	t.Run("Normal input (with escaped splitter) returns slice with single item (whole input)",
		func(t *testing.T) {
			str := "does\\ not\\ split"
			expectedResult := []string{"does not split"}
			result := utils.EscapedSplit(str, ' ')
			assert.Equal(t, expectedResult, result)
		},
	)
	t.Run("Normal escape does nothing to non split character",
		func(t *testing.T) {
			str := "does\\nnot\\nsplit"
			expectedResult := []string{"does\\nnot\\nsplit"}
			result := utils.EscapedSplit(str, ' ')
			assert.Equal(t, expectedResult, result)
		},
	)
	t.Run("Normal input with and without escaped splitter returns correct slice",
		func(t *testing.T) {
			str := "this splits but this\\ will\\ not"
			expectedResult := []string{"this", "splits", "but", "this will not"}
			result := utils.EscapedSplit(str, ' ')
			assert.Equal(t, expectedResult, result)
		},
	)
}
