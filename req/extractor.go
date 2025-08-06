package req

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
)

var allowJSRedirect = true

// 一些提取数据的方法
var (
	cutset            = "\n\t\v\f\r"
	reTitle           = regexp.MustCompile(`(?im)<\s*title.*>(.*?)<\s*/\s*title>`)
	reFaviconLink     = regexp.MustCompile(`(?im)<\s*?link\s*?rel\s*?=\s*?"\s*?(shortcut icon|icon)\s*?"\s*?href\s*?=\s*?"\s*?(.+?)\s*?"\s*?>`)
	reRedirectURLInJS = []*regexp.Regexp{
		regexp.MustCompile(`(?im)\.?location\.(open|replace|assign)\(['"]?(?P<uri>.*?)['"]?\)`),
		regexp.MustCompile(`(?im)\.?location(?:\.href)?\s*?=\s*?['"](?P<uri>.*?)['"]`),
	}
	rePlainWord          = regexp.MustCompile(`^[\p{L}\p{N}\s]+$`)
	titleGuestKeysInJSON = []string{"msg", "message", "info"}
)

// ExtractRedirectURI from a response
func ExtractRedirectURI(data string) (redirectURI string) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(data)))
	if err != nil {
		return
	}
	// 提取 meta 跳转
	doc.Find("meta[http-equiv]").Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "meta" {
			if v, ok := s.Attr("http-equiv"); ok {
				if strings.ToLower(v) != "refresh" {
					return
				}
				if content, exsit := s.Attr("content"); exsit {
					contentL := strings.SplitN(strings.TrimSpace(content), "=", 2)
					if len(contentL) != 2 {
						return
					}
					redirectURI = strings.TrimSpace(contentL[1])
					redirectURI = strings.ReplaceAll(redirectURI, "'", "")
					redirectURI = strings.ReplaceAll(redirectURI, "\"", "")
				}
			}
		}
	})
	if redirectURI != "" {
		return
	}
	// 提取 js 跳转
	if allowJSRedirect {
		for _, r := range reRedirectURLInJS {
			subMatchMaps := ReSubMatchMap(r, data, -1)
			// 提取 js 中最后一个跳转链接
			for _, m := range subMatchMaps {
				uri, ok := m["uri"]
				if !ok {
					continue
				}
				redirectPath := strings.TrimSpace(uri)
				if strings.HasSuffix(redirectPath, "://") {
					continue
				} else if strings.Contains(redirectPath, "www.safedog.cn") {
					continue
				}
				redirectURI = redirectPath
			}
		}
	}
	return
}

// extractTitle from a response
func ExtractTitle(data []byte) (title string) {
	defer func() {
		// 移除非预期字符
		title = strings.TrimSpace(strings.Trim(title, cutset))
	}()
	data = bytes.TrimSpace(data)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		// dom 解析失败
		// 使用 html title 正则匹配
		for _, match := range reTitle.FindAllSubmatch(data, -1) {
			title = string(match[1])
			return
		}
	}
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "title" {
			title = s.Text()
		}
	})
	if title != "" {
		return
	}
	if rePlainWord.Match(data) {
		// 对于一些纯文本，可能一个网页上就一个 SUCCESS，则把此类文本的第一行作为标题返回
		lines := bytes.Split(data, []byte("\n"))
		title = strings.TrimSpace(string(lines[0]))
	} else if bytes.HasPrefix(data, []byte("{")) && bytes.HasSuffix(data, []byte("}")) {
		// 对于一些首页返回 json 字典的情况，则将一些预置可能的键值作为标题返回
		dataMap := make(map[string]interface{})
		err = json.Unmarshal(data, &dataMap)
		if err != nil {
			return
		}
		for _, k := range titleGuestKeysInJSON {
			if v, ok := dataMap[k]; ok {
				title = cast.ToString(v)
			}
		}
	}

	return
}

// extractFaviconLink 从响应体中提取 favicon 链接
func ExtractFaviconLink(data []byte) (links []string) {
	fReMatch := func(s string) {
		for _, match := range reFaviconLink.FindAllStringSubmatch(string(s), -1) {
			links = append(links, match[2])
		}
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		fReMatch(string(data))
		return
	}
	doc.Find("link[rel=\"shortcut icon\"],link[rel=\"icon\"]").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}
		links = append(links, href)
	})
	if len(links) == 0 {
		fReMatch(string(data))
	}
	return
}

func ReSubMatchMap(r *regexp.Regexp, s string, n int) (subMatchMaps []map[string]string) {
	matches := r.FindAllStringSubmatch(s, n)
	for _, match := range matches {
		subMatchMap := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if i != 0 {
				subMatchMap[name] = match[i]
			}
		}
		subMatchMaps = append(subMatchMaps, subMatchMap)
	}
	return
}
