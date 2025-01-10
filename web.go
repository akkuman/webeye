package webeye

import (
	"context"
	"fmt"
	"strings"

	"github.com/akkuman/webeye/finger"
	"github.com/akkuman/webeye/req"
	"github.com/akkuman/webeye/utils"
	mapset "github.com/deckarep/golang-set/v2"
)

func GetWebFinger(ctx context.Context, rawURL string, wfs finger.WebFingerSystem) (res []finger.WebFingerResult, err error) {
	httpClient := req.NewDefaultHTTPClient()
	webxIns := req.NewWebX(&req.Options{MaxRedirects: 3, RateLimit: 1000, Client: httpClient})
	return DoFingerAutoScheme(ctx, webxIns, rawURL, wfs)
}

// DoFingerAutoScheme 自动补充协议的指纹识别
// 支持带 http(s):// 或不带的 rawURL
func DoFingerAutoScheme(ctx context.Context, webxIns *req.WebX, rawURL string, wfs finger.WebFingerSystem) (res []finger.WebFingerResult, err error) {
	u, err := utils.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		res, err = DoFinger(ctx, webxIns, rawURL, wfs)
		return
	} else if u.Scheme == "tcp" || u.Scheme == "" {
		fingers := mapset.NewSet[finger.WebFingerResult]()
		for _, scheme := range []string{"https", "http"} {
			u.Scheme = scheme
			fingers_, err := DoFinger(ctx, webxIns, u.String(), wfs)
			fingers.Append(fingers_...)
			if err != nil {
				return fingers.ToSlice(), err
			}
		}
		return fingers.ToSlice(), nil
	}
	return nil, fmt.Errorf("不支持的 url: %s", rawURL)
}

// DoFinger 执行指纹识别
// targetURL 必须以 http 或 https 开头
func DoFinger(ctx context.Context, webxIns *req.WebX, targetURL string, wfs finger.WebFingerSystem) (res []finger.WebFingerResult, err error) {
	if !strings.HasPrefix(targetURL, "https://") && !strings.HasPrefix(targetURL, "http://") {
		return nil, fmt.Errorf("incorrect target url: %s", targetURL)
	}
	fingers := mapset.NewSet[finger.WebFingerResult]()
	// 请求首页和 favicon
	httpRawDataList, err := webxIns.Request(ctx, targetURL, nil)
	if err != nil {
		return fingers.ToSlice(), err
	}
	for _, hrd := range httpRawDataList {
		fingerResult := wfs.MatchIndex(hrd.Body, hrd.Header, hrd.StatusCode, hrd.FaviconHashList())
		fingers.Append(fingerResult...)
	}
	// 自定义请求
	// 内部实现：自定义请求将不会跟随任何跳转
	for _, wf := range wfs.CustomReqs {
		hrds, err := webxIns.Request(ctx, targetURL, &wf)
		for _, hrd := range hrds {
			if wf.MatchKeyWord(hrd.Body, finger.HTTPHeadersToMap(hrd.Header), hrd.StatusCode) {
				fingers.Add(finger.NewWebFingerResult(wf))
			}
		}
		if err != nil {
			return fingers.ToSlice(), err
		}
	}
	return fingers.ToSlice(), nil
}
