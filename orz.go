package orz

import (
	"bytes"
	"sync"
	"time"
	"unicode"
)

func Maybe[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

func CamelToSnake(s string) string {
	var buf bytes.Buffer
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				buf.WriteByte('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// Debounce 防抖函数：该函数会从上一次被调用后，延迟 wait 毫秒后调用 func 方法。
func Debounce(wait time.Duration) func(f func()) {
	var lock sync.Mutex
	var timer *time.Timer

	return func(f func()) {
		lock.Lock()
		defer lock.Unlock()

		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(wait, f)
	}
}
