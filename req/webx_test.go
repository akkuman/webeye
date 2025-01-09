package req

import (
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
