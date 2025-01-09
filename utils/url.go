package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	schemePrefixPattern = regexp.MustCompile(`^(?:[a-zA-Z0-9]+:){0,1}//`)
)

type MYURL struct {
	Scheme string `json:"scheme"` //协议
	Host   string `json:"host"`   //主机
	Port   uint16 `json:"port"`   //端口
	Path   string `json:"path"`   //路径
}

func (my_url *MYURL) ToString() string {
	if my_url.Scheme == "" {
		my_url.Scheme = "tcp"
	}
	return fmt.Sprintf("%s://%s:%d", my_url.Scheme, my_url.Host, my_url.Port)
}

// AppendURLPath 串联 uri path 路径，和 url JoinPath 行为不一致
// 比如 /a/b/c 和 /nacos/，则会生成 /a/b/c/nacos
func AppendURLPath(paths ...string) string {
	var newPath string
	for _, p := range paths {
		if p == "" {
			continue
		}
		newPath += "/" + p
	}
	for strings.Contains(newPath, "//") {
		newPath = strings.ReplaceAll(newPath, "//", "/")
	}
	return newPath
}

// ParseURL 解析 url
// 支持如下格式：
// baidu.com, baidu.com/a, baidu.com:4433/a, baidu.com:4433, http://baidu.com/a, //baidu.com/a, baidu.com/
func ParseURL(rawURL string) (*url.URL, error) {
	if schemePrefixPattern.MatchString(rawURL) {
		return url.Parse(rawURL)
	}
	return url.Parse("//"+rawURL)
}
