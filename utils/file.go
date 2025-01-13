package utils

import (
	"os"
	"strings"

	"github.com/imroc/req/v3"
)

// GetLocalFileOrWeb get file content from local or web
// it will get file from web when fileURL starts with 'http(s)://'
func GetLocalFileOrWeb(fileURL string) (content []byte, err error) {
	if strings.HasPrefix(fileURL, "http://") || strings.HasPrefix(fileURL, "https://") {
		resp, err := req.DefaultClient().R().Get(fileURL)
		if err != nil {
			return nil, err
		}
		return resp.Bytes(), nil
	}
	return os.ReadFile(fileURL)
}
