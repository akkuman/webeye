package req

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestWebxGetRedirectURL(t *testing.T) {
	tests := []struct{
		input HttpRawData
		wantRedirectURL string
		wantIs30X bool
	} {
		{
			input: HttpRawData{
				URL: url.URL{
					Scheme: "http",
					Host: "localhost",
				},
				Header: make(http.Header),
				StatusCode: 200,
				Body: []byte("<!doctype html>\n<html>\n\t<head>\n\t    <title>Loading...</title>\n\t</head>\n\t<body>\n\t\t<script type=\"text/javascript\">\n\t\t\tlocation.href = \"./ui/\";\n\t\t</script>\n\t</body>\n</html>\n"),
			},
			wantRedirectURL: "http://localhost/ui/",
			wantIs30X: false,
		},
		{
			input: HttpRawData{
				URL: url.URL{
					Scheme: "http",
					Host: "10.50.134.248:8000",
				},
				Header: make(http.Header),
				StatusCode: 200,
				Body: []byte("<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\n<html xmlns=\"http://www.w3.org/1999/xhtml\">\n<head>\n<meta http-equiv=\"Content-Type\" content=\"text/html; charset=gbk2312\" />\n<title></title>\n</head>\n\n<body style=\"   padding:0; margin:0; font:14px/1.5 Microsoft Yahei, \u5b8b\n\u4f53,sans-serif; color:#555;\">\n\n<div style=\"margin:0 auto;width:980px;\">\n      <div style=\"background: url('http://404.safedog.cn/images/safedogsite/head.png') no-repeat;height:300px;\">\n      \t<div style=\"width:300px;height:300px;cursor:pointer;background:#f00;filter: alpha(opacity=0); opacity: 0;float:left;\" onclick=\"location.href='http://www.safedog.cn'\">\n      \t</div>\n      \t<div style=\"float:right;width:430px;height:100px;padding-top:90px;padding-right:90px;font-size:22px;\">\n      \t\t<p id=\"error_code_p\"><a  id=\"eCode\">403</a>\u9519\u8bef<span style=\"font-size:16px;padding-left:15px;\">(\u53ef\u5728\u670d\u52a1\u5668\u4e0a\u67e5\u770b\u5177\u4f53\u9519\u8bef\u4fe1\u606f)</span></p>\n      \t\t<p id=\"eMsg\"></p>\n      \t<a href=\"http://bbs.safedog.cn/thread-60693-1-1.html?from=stat\" target=\"_blank\" style=\"color:#139ff8; font-size:16px; text-decoration:none\">\u7ad9\u957f\u8bf7\u70b9\u51fb</a>\n\t      <a href=\"#\" onclick=\"redirectToHost();\" style=\"color:#139ff8; font-size:16px; text-decoration:none;padding-left: 20px;\">\u8fd4\u56de\u4e0a\u4e00\u7ea7>></a>\n      \t</div>\n      </div>\t\n</div>\n\n\n\n<div style=\"width:1000px; margin:0 auto; \"> \n    <div style=\" width:980px; margin:0 auto;\">\n  <div style=\"width:980px; height:600px; margin:0 auto;\">\n   <iframe allowtransparency=\" true\" src=\"http://404.safedog.cn/sitedog_stat_new.html\"   frameborder=\"no\" border=\"0\" scrolling=\"no\" style=\"width:980px;  height:720px;\" ></iframe>\n </div>\n  </div>\n</div>\n</body>\n</html>\n\n<script>\n\nfunction redirectToHost(){\n            \t var host = location.host;\n                 location.href = \"http://\" + host;\n         }\n\n\nvar errorMsgData = {\n\t\"400\":\"\u8bf7\u6c42\u51fa\u73b0\u8bed\u6cd5\u9519\u8bef\",\n\t\"401\":\"\u6ca1\u6709\u8bbf\u95ee\u6743\u9650\",\n\t\"403\":\"\u670d\u52a1\u5668\u62d2\u7edd\u6267\u884c\u8be5\u8bf7\u6c42\",\n\t\"404\":\"\u6307\u5b9a\u7684\u9875\u9762\u4e0d\u5b58\u5728\",\n\t\"405\":\"\u8bf7\u6c42\u65b9\u6cd5\u5bf9\u6307\u5b9a\u7684\u8d44\u6e90\u4e0d\u9002\u7528\",\n\t\"406\":\"\u5ba2\u6237\u7aef\u65e0\u6cd5\u63a5\u53d7\u76f8\u5e94\u6570\u636e\",\n\t\"408\":\"\u7b49\u5f85\u8bf7\u6c42\u65f6\u670d\u52a1\u5668\u8d85\u65f6\",\n\t\"409\":\"\u8bf7\u6c42\u4e0e\u5f53\u524d\u8d44\u6e90\u7684\u72b6\u6001\u51b2\u7a81\uff0c\u5bfc\u81f4\u8bf7\u6c42\u65e0\u6cd5\u5b8c\u6210\",\n\t\"410\":\"\u8bf7\u6c42\u7684\u8d44\u6e90\u5df2\u4e0d\u5b58\u5728\uff0c\u5e76\u4e14\u6ca1\u6709\u8f6c\u63a5\u5730\u5740\",\n\t\"500\":\"\u670d\u52a1\u5668\u5c1d\u8bd5\u6267\u884c\u8bf7\u6c42\u65f6\u9047\u5230\u4e86\u610f\u5916\u60c5\u51b5\",\n\t\"501\":\"\u670d\u52a1\u5668\u4e0d\u5177\u5907\u6267\u884c\u8be5\u8bf7\u6c42\u6240\u9700\u7684\u529f\u80fd\",\n\t\"502\":\"\u7f51\u5173\u6216\u4ee3\u7406\u670d\u52a1\u5668\u4ece\u4e0a\u6e38\u670d\u52a1\u5668\u6536\u5230\u7684\u54cd\u5e94\u65e0\u6548\",\n\t\"503\":\"\u670d\u52a1\u5668\u6682\u65f6\u65e0\u6cd5\u5904\u7406\u8be5\u8bf7\u6c42\",\n\t\"504\":\"\u5728\u7b49\u5f85\u4e0a\u6e38\u670d\u52a1\u5668\u54cd\u5e94\u65f6\uff0c\u7f51\u5173\u6216\u4ee3\u7406\u670d\u52a1\u5668\u8d85\u65f6\",\n\t\"505\":\"\u670d\u52a1\u5668\u4e0d\u652f\u6301\u8bf7\u6c42\u4e2d\u6240\u7528\u7684 HTTP \u7248\u672c\",\n\t\"1\":\"\u65e0\u6cd5\u89e3\u6790\u670d\u52a1\u5668\u7684 DNS \u5730\u5740\",\n\t\"2\":\"\u8fde\u63a5\u5931\u8d25\",\n\t\"-7\":\"\u64cd\u4f5c\u8d85\u65f6\",\n\t\"-100\":\"\u670d\u52a1\u5668\u610f\u5916\u5173\u95ed\u4e86\u8fde\u63a5\",\n\t\"-101\":\"\u8fde\u63a5\u5df2\u91cd\u7f6e\",\n\t\"-102\":\"\u670d\u52a1\u5668\u62d2\u7edd\u4e86\u8fde\u63a5\",\n\t\"-104\":\"\u65e0\u6cd5\u8fde\u63a5\u5230\u670d\u52a1\u5668\",\n\t\"-105\":\"\u65e0\u6cd5\u89e3\u6790\u670d\u52a1\u5668\u7684 DNS \u5730\u5740\",\n\t\"-109\":\"\u65e0\u6cd5\u8bbf\u95ee\u8be5\u670d\u52a1\u5668\",\n\t\"-138\":\"\u65e0\u6cd5\u8bbf\u95ee\u7f51\u7edc\",\n\t\"-130\":\"\u4ee3\u7406\u670d\u52a1\u5668\u8fde\u63a5\u5931\u8d25\",\n\t\"-106\":\"\u4e92\u8054\u7f51\u8fde\u63a5\u5df2\u4e2d\u65ad\",\n\t\"-401\":\"\u4ece\u7f13\u5b58\u4e2d\u8bfb\u53d6\u6570\u636e\u65f6\u51fa\u73b0\u9519\u8bef\",\n\t\"-400\":\"\u7f13\u5b58\u4e2d\u672a\u627e\u5230\u8bf7\u6c42\u7684\u6761\u76ee\",\n\t\"-331\":\"\u7f51\u7edc IO \u5df2\u6682\u505c\",\n\t\"-6\":\"\u65e0\u6cd5\u627e\u5230\u8be5\u6587\u4ef6\u6216\u76ee\u5f55\",\n\t\"-310\":\"\u91cd\u5b9a\u5411\u8fc7\u591a\",\n\t\"-324\":\"\u670d\u52a1\u5668\u5df2\u65ad\u5f00\u8fde\u63a5\uff0c\u4e14\u672a\u53d1\u9001\u4efb\u4f55\u6570\u636e\",\n\t\"-346\":\"\u6536\u5230\u4e86\u6765\u81ea\u670d\u52a1\u5668\u7684\u91cd\u590d\u6807\u5934\",\n\t\"-349\":\"\u6536\u5230\u4e86\u6765\u81ea\u670d\u52a1\u5668\u7684\u91cd\u590d\u6807\u5934\",\n\t\"-350\":\"\u6536\u5230\u4e86\u6765\u81ea\u670d\u52a1\u5668\u7684\u91cd\u590d\u6807\u5934\",\n\t\"-118\":\"\u8fde\u63a5\u8d85\u65f6\"\n};\n\nvar eCode = document.getElementById(\"eCode\").innerHTML;\nvar eMsg = errorMsgData[eCode];\ndocument.title = eMsg;\ndocument.getElementById(\"eMsg\").innerHTML = eMsg;\n</script>\n<script type=\"text/javascript\" src=\"http://404.safedog.cn/Scripts/url.js\"></script>"),
			},
			wantRedirectURL: "",
			wantIs30X: false,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("Webx.getRedirectURL-%s", tc.input.URL.String()), func(t *testing.T) {
			httpClient := NewDefaultHTTPClient()
			webxIns := NewWebX(&Options{MaxRedirects: 3, RateLimit: 1000, Client: httpClient})
			redirectURL, is30X := webxIns.getRedirectURL(tc.input, nil)
			if redirectURL != tc.wantRedirectURL || is30X != tc.wantIs30X {
				t.Errorf("got %#v; want %#v", []any{redirectURL, is30X}, []any{tc.wantRedirectURL, tc.wantIs30X})
				return
			}
		})
	}
}

func TestExtractRedirectURI(t *testing.T) {
	tests := []struct{
		input string
		want string
	} {
		{
			input: "<!doctype html>\n<html>\n\t<head>\n\t    <title>Loading...</title>\n\t</head>\n\t<body>\n\t\t<script type=\"text/javascript\">\n\t\t\tlocation.href = \"./ui/\";\n\t\t</script>\n\t</body>\n</html>\n",
			want: "./ui/",
		},
		{
			input: `<html>
<head>
        <title>TEST</title>
       <link href="favicon.ico" rel="icon" type="image/x-icon" />
        <meta http-equiv="X-UA-Compatible" content="chrome=1,IE=edge"/>
<meta http-equiv="refresh" content="0;url=http://localhost:8075/WebReport/ReportServer?op=fs_load&cmd=fs_signin&_=1617601373118">
    <style type="text/css">
            html, body
        {
            margin: 0px 0px;
            width: 100%;
            height: 100%;
        }
            iframe
        {
            margin: 0px 0px;
            width: 100%;
            height: 100%;
        }
    </style>
    </head>
<body>
<script type="text/javascript">
var myvalue = "jjjjj";
function fun1() {
      return "function return test";
  }
document.write("Hello World!");
</script>
</body>
</html>`,
			want: "http://localhost:8075/WebReport/ReportServer?op=fs_load&cmd=fs_signin&_=1617601373118",
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("ExtractRedirectURI-%s", tc.want), func(t *testing.T) {
			got := ExtractRedirectURI(tc.input)
			if got != tc.want {
				t.Errorf("got %#v; want %#v", got, tc.want)
			}
		})
	}
}

func TestStatusCodeTitle(t *testing.T) {
	tests := []struct{
		input string
		wantStatusCode int
		wantTitle string
	} {
		{
			"https://113.55.8.9:8084/",
			200,
			"电子资源馆外访问系统",
		},
		{
			"http://113.55.126.44",
			200,
			"",
		},
		{
			"https://mlzy.lzszyyy.com",
			200,
			"",
		},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			httpClient := NewDefaultHTTPClient()
			webxIns := NewWebX(&Options{MaxRedirects: 4, RateLimit: 1000, Client: httpClient})
			hrds, err := webxIns.Request(context.Background(), tc.input, nil)
			if err != nil {
				t.Error(err)
				return
			}
			if len(hrds) == 0 {
				t.Error("响应链长度为 0")
				return
			}
			var statusCode int
			var title string
			for _, hrd := range hrds {
				if hrd.StatusCode != 0 {
					statusCode = hrd.StatusCode
				}
				if hrd.Title != "" {
					title = hrd.Title
				}
			}
			if statusCode != tc.wantStatusCode || title != tc.wantTitle {
				t.Errorf(
					"got title:%s status_code:%d; want title:%s status_code:%d",
					title,
					statusCode,
					tc.wantTitle,
					tc.wantStatusCode,
				)
			}
		})
	}
}