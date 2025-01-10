package req

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"time"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/akkuman/webeye/cache"
	"github.com/akkuman/webeye/finger"
	"github.com/akkuman/webeye/utils"
	mapset "github.com/deckarep/golang-set/v2"
	req "github.com/imroc/req/v3"
	"go.uber.org/ratelimit"
)

var emialReg = regexp.MustCompile(`(?mi)[A-Za-z0-9.\-+_]+@[a-z0-9.\-+_]+\.[a-z]+`)

type HttpRawData struct {
	URL         url.URL            //当前 URL
	Header      http.Header        //响应头
	StatusCode  int                //状态码
	Body        []byte             //响应体
	Title       string             //标题
	URLS        []string           //提取到的 URL
	ICPS        []string           //提取到的 ICP 备案
	Email       []string           //提取到的 Email
	FaviconHash []Favicon          //图标 hash
	X509Cert    []x509.Certificate //证书
}

func (hrd *HttpRawData) FaviconHashList() []string {
	var favicons []string
	for _, v := range hrd.FaviconHash {
		if v.MD5 != "" {
			favicons = append(favicons, v.MD5)
		}
		if v.MMH3 != "" {
			favicons = append(favicons, v.MMH3)
		}
	}
	return favicons
}

type Options struct {
	// 最大跟随跳转次数
	MaxRedirects int
	// 请求 web 的最大速率每秒，可以理解为 N 个请求/s
	RateLimit int
	// 用那个客户端请求
	Client *req.Client
	Cache cache.Cache
}

type WebFingerPrintRequest struct {
	Path          string            `json:"path"`            //指纹请求路径
	RequestMethod string            `json:"request_method"`  //请求方式
	RequestHeader map[string]string `json:"request_headers"` //请求头
	RequestData   string            `json:"request_data"`    //请求体 base64 编码后字符串
}

type HttpDataStruct struct {
	HttpRawData []HttpRawData
	Error       error
}

type WebX struct {
	opt     *Options
	limiter ratelimit.Limiter
	client  *req.Client
	cache   cache.Cache
}

func NewWebX(opt *Options) *WebX {
	x := new(WebX)
	x.opt = opt
	// 增加限速器
	if opt.RateLimit > 0 {
		x.limiter = ratelimit.New(opt.RateLimit)
	} else {
		x.limiter = ratelimit.NewUnlimited()
	}
	x.client = opt.Client
	x.cache = opt.Cache
	return x
}

// Request 发送请求，wf 为请求指纹，如果传 nil，代表为首页或 favicon 指纹
func (x *WebX) Request(ctx context.Context, targetURL string, wf *finger.WebFinger) ([]HttpRawData, error) {
	d, _ := json.Marshal(map[string]any{
		"url": targetURL,
		"finger": wf,
	})
	cacheKey := utils.MD5Hex(d)
	var v []HttpRawData
	if x.cache != nil {
		if err := x.cache.Get(cacheKey, &v); err == nil {
			// 如果有缓存就返回缓存
			return v, nil
		}
	}
	// 没有中缓存就请求一次
	hrds, err := x.doWebHTMLRequest(ctx, targetURL, wf)
	if err != nil {
		return hrds, err
	}
	//保存缓存
	if x.cache != nil {
		x.cache.Set(cacheKey, hrds, 24*time.Hour)
	}
	return hrds, nil
}

// getResponse 发送请求获取响应，注意：返回的 *http.Response 将不能再被读取 body
func (x *WebX) getResponse(ctx context.Context, rawURL string, wf *finger.WebFinger) ([]byte, *req.Response, error) {
	x.limiter.Take()
	targetURL := rawURL
	request := x.client.R().DisableAutoReadResponse().SetContext(ctx)
	requestMethod := http.MethodGet
	if wf != nil {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, nil, err
		}
		newURL := &url.URL{
			Scheme: parsedURL.Scheme,
			Host: parsedURL.Host,
			Path: utils.AppendURLPath(parsedURL.Path, wf.Request.Path),
		}
		targetURL = newURL.String()
		// 非首页 favicon 指纹
		request.SetHeaders(wf.Request.RequestHeader)
		request.SetBodyBytes(wf.Request.RequestData)
		if wf.Request.RequestMethod != "" {
			requestMethod = wf.Request.RequestMethod
		}
	}
	resp, err := request.Send(requestMethod, targetURL)
	if err != nil {
		return []byte{}, resp, err
	}
	defer resp.Body.Close()
	// websockets don't have a readable body
	if resp.StatusCode == http.StatusSwitchingProtocols {
		return []byte{}, resp, err
	}
	reader := io.LimitReader(resp.Body, 2*1024*1024)
	respbody, err := io.ReadAll(reader)
	if err != nil {
		// Edge case - some servers respond with gzip encoding header but uncompressed body, in this case the standard library configures the reader as gzip, triggering an error when read.
		// The bytes slice is not accessible because of abstraction, therefore we need to perform the request again tampering the Accept-Encoding header
		return []byte{}, resp, err
	}
	return respbody, resp, err
}

// 获取跳转 URL
func (x *WebX) getRedirectURL(http_raw_data HttpRawData) (redirectURL string, is30X bool) {
	u, _ := url.Parse(http_raw_data.URL.String())
	// 协议头跳转
	dnspod := []string{"dnspod.qcloud.com", "www.wendns.com"}
	location := http_raw_data.Header.Get("Location")
	if location == "" {
		location = http_raw_data.Header.Get("location")
	}
	if location != "" {
		is30X = true
		uu, err := u.Parse(location)
		if err != nil {
			return
		}
		// IP 跳转到云服务商备案提示页面
		isDnsPod := slices.Contains(dnspod, uu.Hostname()) || strings.Contains(uu.Hostname(), "dnspod")
		if isDnsPod {
			return
		}
		redirectURL = uu.String()
		return
	}
	// 非协议头跳转
	redirectURI := ExtractRedirectURI(string(http_raw_data.Body))
	if redirectURI != "" {
		uu, err := u.Parse(redirectURI)
		if err != nil {
			return
		}
		redirectURL = uu.String()
	}
	return
}

// 首页加跳转请求，返回元数据列表，wf 不为 nil，代表是自定义请求
func (x *WebX) doWebHTMLRequest(ctx context.Context, targetURL string, wf *finger.WebFinger) ([]HttpRawData, error) {
	// 首页请求
	HttpRawDataList := make([]HttpRawData, 0)
	body, httpresp, err := x.getResponse(ctx, targetURL, wf)
	if err != nil {
		return HttpRawDataList, err
	}
	hrd, err := x.responseToHttpRawData(httpresp.Response, body)
	if err != nil {
		return []HttpRawData{hrd}, err
	}
	if wf != nil {
		// 对于自定义请求的情况，不需要进行跟随跳转，也不需要请求 favicon，直接执行自定义请求即可
		return []HttpRawData{hrd}, nil
	}
	hrd.FaviconHash = x.getFavicon(ctx, httpresp.Response, hrd.Body)
	HttpRawDataList = append(HttpRawDataList, hrd)
	currentRedirectCount := 0
	for {
		// 计算跳转次数，达到最大值退出
		currentRedirectCount += 1
		if currentRedirectCount > x.opt.MaxRedirects {
			break
		}
		redirectURL, _ := x.getRedirectURL(hrd)
		if redirectURL == "" {
			// 没有更多的跳转
			break
		}
		newURL, err := hrd.URL.Parse(redirectURL)
		if err != nil {
			break
		}
		if newURL.String() == hrd.URL.String() {
			// 如果此次跳转和上一次跳转相同，则退出
			break
		}
		body, httpresp, err := x.getResponse(ctx, newURL.String(), nil)
		if err != nil {
			return HttpRawDataList, err
		}
		hrd, err = x.responseToHttpRawData(httpresp.Response, body)
		hrd.FaviconHash = x.getFavicon(ctx, httpresp.Response, hrd.Body)
		if err != nil {
			break
		}
		HttpRawDataList = append(HttpRawDataList, hrd)
	}
	return HttpRawDataList, nil
}

func (x *WebX) responseToHttpRawData(resp *http.Response, respbody []byte) (HttpRawData, error) {
	http_raw_data := HttpRawData{
		URL:        *resp.Request.URL,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		URLS:       make([]string, 0),
		Body:       respbody,
		Email:      make([]string, 0),
		ICPS:       make([]string, 0),
		Title:      "",
		X509Cert:   make([]x509.Certificate, 0),
	}
	if resp.StatusCode == http.StatusSwitchingProtocols {
		return http_raw_data, fmt.Errorf("StatusSwitchingProtocols")
	}
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := make([]x509.Certificate, 0)
		if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
			peer_cert := resp.TLS.PeerCertificates
			for i := range peer_cert {
				cert = append(cert, *peer_cert[i])
			}
		}
		http_raw_data.X509Cert = cert
	}
	title, links, icp := GetWebTitleAndUrlsAndIPC(http_raw_data.Body)
	// 不是 30x 跳转才设置标题
	if !(resp.StatusCode > 300 && resp.StatusCode < 400) {
		http_raw_data.Title = title
	}
	http_raw_data.URLS = links
	http_raw_data.ICPS = icp
	http_raw_data.Email = emialReg.FindAllString(string(http_raw_data.Body), -1)
	return http_raw_data, nil
}

func (x *WebX) getFavicon(ctx context.Context, resp *http.Response, body []byte) (favicons []Favicon) {
	// 对于服务器错误的情况，直接跳过
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return
	}
	links := ExtractFaviconLink(body)
	// 浏览器在找不到 favicon 的情况下会自动访问该路径
	links = append(links, "/favicon.ico")
	for j := range links {
		if resp.Request == nil || resp.Request.URL == nil {
			continue
		}
		iconURL, err := resp.Request.URL.Parse(links[j])
		if err != nil {
			continue
		}
		faviconURL := iconURL.String()
		if strings.HasPrefix(links[j], "http://") || strings.HasPrefix(links[j], "https://") {
			faviconURL = links[j]
		}
		fr := x.GetFav(ctx, faviconURL)
		if fr.Error == nil {
			favicons = append(favicons, fr.Favicon)
		}
	}
	return
}



// 获取标题，URL 列表，ICP 备案
func GetWebTitleAndUrlsAndIPC(body []byte) (string, []string, []string) {
	bodyReader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return "", []string{}, []string{}
	}
	urls := mapset.NewSet[string]()
	icps := mapset.NewSet[string]()
	Title := ExtractTitle(body)
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		link, exists := selection.Attr("href")
		if exists {
			u, err := url.Parse(link)
			if err == nil {
				if u.Host == "www.beian.gov.cn" || u.Host == "beian.miit.gov.cn" {
					IpcName := selection.Text()
					if len(IpcName) > 0 {
						icps.Add(strings.TrimSpace(IpcName))
					}
				}
			}
		}
	})
	return Title, urls.ToSlice(), icps.ToSlice()
}

// 图标 hash 缓存
type FavCacheStruct struct {
	Favicon Favicon
	Error   error
}

func buildFavCacheKey(favURL string) string {
	return fmt.Sprintf("favicon:%s", favURL)
}

// 根据 url 返回图标 hash
func (x *WebX) GetFav(ctx context.Context, faviconURL string) FavCacheStruct {
	var v Favicon
	if x.cache != nil {
		err := x.cache.Get(buildFavCacheKey(faviconURL), &v)
		if err == nil {
			return FavCacheStruct{
				Error:   nil,
				Favicon: v,
			}
		}
	}
	x.limiter.Take()
	resp, err := x.client.R().SetContext(ctx).DisableAutoReadResponse().Get(faviconURL)
	if err != nil {
		return FavCacheStruct{Error: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return FavCacheStruct{Error: fmt.Errorf("StatusCode Not OK")}
	}
	reader := io.LimitReader(resp.Body, 2*1024*1024)
	respbody, err := io.ReadAll(reader)
	if err != nil {
		return FavCacheStruct{Error: err}
	}
	if !strings.Contains(strings.ToLower(resp.GetContentType()), "image") || strings.Contains(string(respbody), "<html>") {
		return FavCacheStruct{Error: fmt.Errorf("ContentType Not Image")}
	}
	fr := FavCacheStruct{Error: nil, Favicon: Favicon{
		URL:  faviconURL,
		MMH3: ShodanIconHash(respbody),
		MD5:  MD5IconHash(respbody),
		Data: respbody,
	}}
	if x.cache != nil {
		x.cache.Set(buildFavCacheKey(faviconURL), fr.Favicon, 24*time.Hour)
	}
	return fr
}
