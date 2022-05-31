package vncpasswd

import (
	"github.com/gogf/gf/test/gtest"
	"github.com/gogf/gf/util/gconv"
	"testing"
)

func TestVncByGo(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		var plaintext = "lzmlzm11"
		b := AuthVNCEncode(gconv.Bytes(plaintext))
		t.Log(string(AuthVNCDecrypt(b)))
		t.Assert(AuthVNCDecrypt(b), plaintext)
	})
}
