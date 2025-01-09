package finger

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	urlPathPattern = regexp.MustCompile(`^[a-zA-Z0-9\-._~!$&'()*+,;=:@\/]*$`)
)

// 把 map 的键值都转为小写
func lowerMap(m map[string]string) map[string]string {
	newMap := make(map[string]string, 0)
	for k, v := range m {
		newMap[strings.ToLower(k)] = strings.ToLower(v)
	}
	return newMap
}

// CheckRootPath 检查 root_path 是否符合规范
func CheckRootPath(rootPath string) error {
	if !urlPathPattern.MatchString(rootPath) {
		return fmt.Errorf("不是一个合法的 root_path")
	}
	if !strings.HasPrefix(rootPath, "/") || !strings.HasSuffix(rootPath, "/") {
		return fmt.Errorf("自定义 root_path 必须以 / 开头，以 / 结尾")
	}
	return nil
}

func HTTPHeadersToMap(headers http.Header) map[string]string {
	headerMap := make(map[string]string)
	for hk, hv := range headers {
		headerMap[strings.ToLower(hk)] = strings.ToLower(strings.Join(hv, "; "))
	}
	return headerMap
}