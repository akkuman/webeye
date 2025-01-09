package utils

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/go-test/deep"
)

func TestAppendURLPath(t *testing.T) {
	tests := []struct{
		args []string
		want string
	} {
		{[]string{"/a/b/c", "/nacos/"}, "/a/b/c/nacos/"},
		{[]string{"/a/b/c", "/nacos//"}, "/a/b/c/nacos/"},
		{[]string{"/a/b/c/", "/nacos//"}, "/a/b/c/nacos/"},
		{[]string{"/a/b/c/", "nacos/"}, "/a/b/c/nacos/"},
		{[]string{"/a/b/c", ""}, "/a/b/c"},
		{[]string{"/a/b/c", "/nacos"}, "/a/b/c/nacos"},
	}

	for _, tc := range tests {
		got := AppendURLPath(tc.args...)
		if got != tc.want {
			t.Errorf("AppendURLPath(%v) = %s; want %s", tc.args, got, tc.want)
		}
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct{
		rawURL string
		want *url.URL
		err error
	} {
		{"baidu.com", &url.URL{Scheme: "", Host: "baidu.com", Path: ""}, nil},
		{"baidu.com/a", &url.URL{Scheme: "", Host: "baidu.com", Path: "/a"}, nil},
		{"baidu.com:4433/a", &url.URL{Scheme: "", Host: "baidu.com:4433", Path: "/a"}, nil},
		{"baidu.com:4433", &url.URL{Scheme: "", Host: "baidu.com:4433", Path: ""}, nil},
		{"http://baidu.com/a", &url.URL{Scheme: "http", Host: "baidu.com", Path: "/a"}, nil},
		{"//baidu.com/a", &url.URL{Scheme: "", Host: "baidu.com", Path: "/a"}, nil},
		{"baidu.com/", &url.URL{Scheme: "", Host: "baidu.com", Path: "/"}, nil},
		{"tcp://baidu.com/", &url.URL{Scheme: "tcp", Host: "baidu.com", Path: "/"}, nil},
		{"tcp://baidu.com/a/b/c", &url.URL{Scheme: "tcp", Host: "baidu.com", Path: "/a/b/c"}, nil},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("ParseURL(%s)", tc.rawURL), func(t *testing.T) {
			got, err := ParseURL(tc.rawURL)
			if !ContainsErr(err, tc.err) {
				t.Errorf("error = %v; want %v", err, tc.err)
				return
			}
			if diff := deep.Equal(got, tc.want); diff != nil {
				t.Errorf("got %#v; want %#v; diff: %#v", got, tc.want, diff)
			}
		})
		
	}
}