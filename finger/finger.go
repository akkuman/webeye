package finger

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type RequestInfo struct {
	Path          string            `json:"path"`            // 请求路径
	RequestMethod string            `json:"request_method"`  // 请求方法，默认 GET
	RequestHeader map[string]string `json:"request_headers"` // 请求头
	RequestData   []byte            `json:"request_data"`    // 请求数据
}

type MatchRule struct {
	StatusCode  int               `json:"status_code"`  // 匹配状态码
	FaviconHash []string          `json:"favicon_hash"` // 匹配图标 hash，一个匹配到了就算命中
	Headers     map[string]string `json:"headers"`      // 匹配全球头，读取键，匹配值，如果值为*或者空，只匹配键
	Keyword     []string          `json:"keyword"`      // 匹配关键词
}

type WebFinger struct {
	Name       string      `json:"name"`        // 指纹名称
	Priority   int         `json:"priority"`    // 指纹优先度
	Request    RequestInfo `json:"request"`     // 自定义请求
	MatchRules MatchRule   `json:"match_rules"` // 匹配规则
	RootPath   string      `json:"root_path"`   // 站点根路径，默认为 /
}

func (wf *WebFinger) IsIndex() bool {
	return wf.Request.Path == "/" && len(wf.Request.RequestHeader) == 0 && strings.ToLower(wf.Request.RequestMethod) == "get" && len(wf.Request.RequestData) == 0 && len(wf.MatchRules.FaviconHash) == 0
}

func (wf *WebFinger) IsFavicon() bool {
	return len(wf.MatchRules.FaviconHash) > 0
}

func (wf *WebFinger) IsCustom() bool {
	return wf.Request.Path != "/" || strings.ToLower(wf.Request.RequestMethod) != "get" || len(wf.Request.RequestHeader) != 0 || len(wf.Request.RequestData) != 0
}

// MatchKeyWord 匹配指纹判断是否命中
func (wf *WebFinger) MatchKeyWord(data []byte, headers map[string]string, status_code int) bool {
	// 匹配状态码，指纹规则中有状态码，但是和传进来的不匹配
	if wf.MatchRules.StatusCode != 0 && wf.MatchRules.StatusCode != status_code {
		return false
	}
	// 匹配 header，指纹规则中有请求头，但是没有找到键
	for k, v := range wf.MatchRules.Headers {
		if hk, ok := headers[strings.ToLower(k)]; ok {
			// *时只匹配键
			if v != "*" && !strings.Contains(hk, strings.ToLower(v)) {
				return false
			}
		} else {
			return false
		}
	}
	// 匹配正文
	// 提前判断防止 []byte->string 的转化
	if len(wf.MatchRules.Keyword) != 0 {
		bodytext := strings.ToLower(string(data))
		for _, keyword := range wf.MatchRules.Keyword {
			if !strings.Contains(bodytext, strings.ToLower(keyword)) {
				return false
			}
		}
	}
	return true
}

func (wf *WebFinger) MatchFavicon(favicons []string) bool {
	// 匹配图标
	favicon_hash_set := mapset.NewSet(favicons...)
	if len(wf.MatchRules.FaviconHash) > 0 {
		// 存在 favicon 指纹的情况下，指纹中的 iconhash 没有一个匹配到，则指纹匹配失败
		favicon_set := mapset.NewSet(wf.MatchRules.FaviconHash...)
		if len(favicon_set.Intersect(favicon_hash_set).ToSlice()) == 0 {
			return false
		}
	}
	return true
}

type WebFingerRaw struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Keyword       []string          `json:"keyword"`
	Priority      int               `json:"priority"`
	RequestMethod string            `json:"request_method"`
	RequestHeader map[string]string `json:"request_headers"`
	RequestData   string            `json:"request_data"` // base64 编码后的请求体
	FaviconHash   []string          `json:"favicon_hash"`
	RootPath      string            `json:"root_path"` // 站点根路径，默认为 /
}

// json 转为首页，特殊路径和图标 hash 指纹
func (wfr *WebFingerRaw) toWebFinger() (wf *WebFinger, err error) {
	var reqData []byte
	if len(wfr.RequestData) != 0 {
		reqData, err = base64.RawStdEncoding.DecodeString(wfr.RequestData)
		if err != nil {
			return
		}
	}
	// 处理 RootPath
	rootPath := "/"
	if wfr.RootPath != "" {
		if err = CheckRootPath(wfr.RootPath); err != nil {
			return nil, err
		}
		rootPath = wfr.RootPath
	}
	request := RequestInfo{
		Path:          wfr.Path,
		RequestMethod: strings.ToLower(wfr.RequestMethod),
		RequestHeader: wfr.RequestHeader,
		RequestData:   reqData,
	}
	match_rules := MatchRule{
		Keyword:     wfr.Keyword,
		Headers:     lowerMap(wfr.Headers),
		FaviconHash: wfr.FaviconHash,
		StatusCode:  wfr.StatusCode,
	}
	wf = &WebFinger{
		Name:       wfr.Name,
		Priority:   wfr.Priority,
		Request:    request,
		MatchRules: match_rules,
		RootPath:   rootPath,
	}
	return
}

type WebFingerSystem struct {
	Indexs     []WebFinger // 首页请求指纹
	CustomReqs []WebFinger // 自定义请求指纹
	Favicons   []WebFinger // favicon 指纹
}

// ParseWebFinger 解析 web 指纹，传入一个列表（json）
func ParseWebFinger(content string) (*WebFingerSystem, error) {
	wfrList := make([]WebFingerRaw, 0)
	err := json.Unmarshal([]byte(content), &wfrList)
	if err != nil {
		return nil, err
	}
	wfs := new(WebFingerSystem)
	for _, wfr := range wfrList {
		wf, err := wfr.toWebFinger()
		if err != nil {
			return nil, err
		}
		if wf.IsIndex() {
			wfs.Indexs = append(wfs.Indexs, *wf)
		} else if wf.IsFavicon() {
			wfs.Favicons = append(wfs.Favicons, *wf)
		} else if wf.IsCustom() {
			wfs.Favicons = append(wfs.CustomReqs, *wf)
		}
	}
	return wfs, nil
}

type WebFingerResult struct {
	Name string
	RootPath string
}

func NewWebFingerResult(wf WebFinger) WebFingerResult {
	return WebFingerResult{
		Name: wf.Name,
		RootPath: wf.RootPath,
	}
}

// 匹配首页和 favicon 指纹
func (wfs *WebFingerSystem) MatchIndex(data []byte, headers http.Header, statusCode int, favicons []string) []WebFingerResult {
	res := mapset.NewSet[WebFingerResult]()
	headerMap := HTTPHeadersToMap(headers)
	// 首页匹配
	for _, f := range wfs.Indexs {
		if f.MatchKeyWord(data, headerMap, statusCode) {
			res.Add(NewWebFingerResult(f))
		}
	}
	// favicon 匹配
	for _, f := range wfs.Favicons {
		if f.MatchFavicon(favicons) {
			res.Add(NewWebFingerResult(f))
		}
	}
	return res.ToSlice()
}
