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
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("Webx.getRedirectURL-%s", tc.input.URL.String()), func(t *testing.T) {
			httpClient := NewDefaultHTTPClient()
			webxIns := NewWebX(&Options{MaxRedirects: 3, RateLimit: 1000, Client: httpClient})
			redirectURL, is30X := webxIns.getRedirectURL(tc.input)
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