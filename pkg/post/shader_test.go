package post_test

import (
	"testing"

	"github.com/adm87/finch/pkg/post"
)

func TestShaderPanicsOnNil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Shader(nil) did not panic")
		}
	}()
	post.Shader(nil, nil)
}
