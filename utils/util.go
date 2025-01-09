package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func MD5Hex(dat []byte) string {
	hash := md5.Sum(dat)
   	return hex.EncodeToString(hash[:])
}

// ContainsErr 检查输出的 err 和期望的 err 是否符合
// 注意，只会检查 err.Error() 是否有包含关系
func ContainsErr(gotErr, wantErr error) bool {
	if gotErr == nil && wantErr == nil {
		return true
	}
	if gotErr == nil && wantErr != nil {
		return false
	}
	if gotErr != nil && wantErr == nil {
		return false
	}
	if !strings.Contains(gotErr.Error(), wantErr.Error()) {
		return false
	}
	return true
}
