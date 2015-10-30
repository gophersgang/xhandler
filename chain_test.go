package xhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAppendHanlerC(t *testing.T) {
	init := 0
	h1 := func(next HandlerC) HandlerC {
		init++
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 2)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	c := Chain{}
	c.UseC(h1)
	c.UseC(h2)
	assert.Len(t, c, 2)

	h := c.Handler(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 2, ctx.Value("test"), "second handler should overwrite first handler's context value")
	}))

	h.ServeHTTP(nil, nil)
	h.ServeHTTP(nil, nil)
	assert.Equal(t, 1, init, "handler init called once")
}

func TestAppendHanler(t *testing.T) {
	init := 0
	h1 := func(next HandlerC) HandlerC {
		return HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ctx = context.WithValue(ctx, "test", 1)
			next.ServeHTTPC(ctx, w, r)
		})
	}
	h2 := func(next http.Handler) http.Handler {
		init++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Change r and w values
			w = httptest.NewRecorder()
			r = &http.Request{}
			next.ServeHTTP(w, r)
		})
	}
	c := Chain{}
	c.UseC(h1)
	c.Use(h2)
	assert.Len(t, c, 2)

	h := c.Handler(HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		// Test ordering
		assert.Equal(t, 1, ctx.Value("test"),
			"the first handler value should be pass through the second (non-aware) one")
		// Test r and w overwrite
		assert.NotNil(t, w)
		assert.NotNil(t, r)
	}))

	h.ServeHTTP(nil, nil)
	h.ServeHTTP(nil, nil)
	assert.Equal(t, 1, init, "handler init called once")
}
