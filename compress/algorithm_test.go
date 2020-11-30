// SPDX-License-Identifier: MIT

package compress

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ WriterFunc = NewDeflate
	_ WriterFunc = NewGzip
	_ WriterFunc = NewBrotli
	_ WriterFunc = newErrorWriter
)

func newErrorWriter(w io.Writer) (Writer, error) {
	return nil, errors.New("error")
}

func TestCompress_AddAlgorithm(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", 0), "application/xml", "text/*", "application/json")
	a.NotNil(c)

	a.False(c.AddAlgorithm("br", NewBrotli))
	a.Equal(1, len(c.algorithms))

	a.True(c.AddAlgorithm("br", NewBrotli))
	a.Equal(1, len(c.algorithms))

	a.Panic(func() {
		c.AddAlgorithm("", NewGzip)
		a.Equal(1, len(c.algorithms))
	})
	a.Panic(func() {
		c.AddAlgorithm("gzip", nil)
		a.Equal(1, len(c.algorithms))
	})
}

func TestCompress_SetAlgorithm(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", 0), "application/xml", "text/*", "application/json")

	a.Equal(0, len(c.algorithms))
	c.SetAlgorithm("gzip", NewGzip)
	a.Equal(1, len(c.algorithms))
	c.SetAlgorithm("gzip", nil)
	a.Equal(0, len(c.algorithms))

	c.SetAlgorithm("gzip", NewGzip)
	c.SetAlgorithm("br", NewBrotli)
	a.Equal(2, len(c.algorithms))

	a.Panic(func() {
		c.SetAlgorithm("identity", NewGzip)
	})
}

func TestCompress_findAlgorithm(t *testing.T) {
	a := assert.New(t)

	// 空值相当于 identity
	c := newCompress(a, "application/xml", "text/*", "application/json")
	r := httptest.NewRequest(http.MethodDelete, "/", nil)
	name, f, na := c.findAlgorithm(r)
	a.False(na).Empty(name).Nil(f)

	// identity
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "identity")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Empty(name).Nil(f)

	// *
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "*")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Equal(name, "deflate").NotNil(f)

	// br
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "br")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Equal(name, "br").NotNil(f)

	// br;q=0.9,gzip,deflate
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "br;q=0.9,gzip,deflate")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Equal(name, "gzip").NotNil(f)

	// identity,br;q=0.9,gzip,deflate
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "identity,br;q=0.9,gzip,deflate")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Empty(name).Nil(f)

	// br;q=0.9,gzip,deflate,identity
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "br;q=0.9,gzip,deflate,identity")
	name, f, na = c.findAlgorithm(r)
	a.False(na).Equal(name, "gzip").NotNil(f)

	// n1,n2,identity;q=0
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "n1,n2,identity;q=0")
	name, f, na = c.findAlgorithm(r)
	a.True(na).Empty(name).Nil(f)

	// n1,n2,*;q=0
	c = newCompress(a, "application/xml", "text/*", "application/json")
	r = httptest.NewRequest(http.MethodDelete, "/", nil)
	r.Header.Add("accept-encoding", "n1,n2,*;q=0")
	name, f, na = c.findAlgorithm(r)
	a.True(na).Empty(name).Nil(f)
}
