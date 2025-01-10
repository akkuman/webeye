package webeye

import (
	"context"
	"fmt"
	"testing"

	"github.com/akkuman/webeye/finger"
	"github.com/akkuman/webeye/req"
	"github.com/akkuman/webeye/utils"
	"github.com/go-test/deep"
)

func TestDoFinger(t *testing.T) {
	tests := []struct{
		targetURL string
		fingerJSON string
		want []finger.WebFingerResult
		err error
	} {
		{
			targetURL: "https://58.210.187.98:4433",
			fingerJSON: `[{
				"path": "/",
				"request_method": "get",
				"request_headers": {},
				"request_data": "",
				"status_code": 0,
				"headers": {},
				"keyword": ["日志分析平台", "SFViewVersion"],
				"priority": 3,
				"favicon_hash": [],
				"name": "sangfor-ba"
			}]`,
			want: []finger.WebFingerResult{
				{Name: "sangfor-ba", RootPath: "/"},
			},
			err: nil,
		},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("DoFinger(%s)", tc.targetURL), func(t *testing.T) {
			httpClient := req.NewDefaultHTTPClient()
			webxIns := req.NewWebX(&req.Options{MaxRedirects: 3, RateLimit: 1000, Client: httpClient})
			wfs, err := finger.ParseWebFinger(tc.fingerJSON)
			if err != nil {
				t.Error(err)
				return
			}
			got, err := DoFinger(context.Background(), webxIns, tc.targetURL, *wfs)
			if !utils.ContainsErr(err, tc.err) {
				t.Errorf("error = %v; want %v", err, tc.err)
				return
			}
			if diff := deep.Equal(got, tc.want); diff != nil {
				t.Errorf("got %#v; want %#v; diff: %#v", got, tc.want, diff)
			}
		})
	}
}

func TestGetWebFinger(t *testing.T) {
	tests := []struct{
		targetURL []string
		fingerJSON string
		want []finger.WebFingerResult
		err error
	} {
		{
			targetURL: []string{
				"https://58.210.187.98:4433",
			},
			fingerJSON: `[{
				"path": "/",
				"request_method": "get",
				"request_headers": {},
				"request_data": "",
				"status_code": 0,
				"headers": {},
				"keyword": ["日志分析平台", "SFViewVersion"],
				"priority": 3,
				"favicon_hash": [],
				"name": "sangfor-ba"
			}]`,
			want: []finger.WebFingerResult{
				{Name: "sangfor-ba", RootPath: "/"},
			},
			err: nil,
		},
		{
			targetURL: []string{
				"https://fr.mingr.com.cn/",
				"https://www.dqyjc.org.cn",
				"http://erp.zschunkai.com:8075/",
			},
			fingerJSON: `[{
				"path": "/",
				"request_method": "get",
				"request_headers": {},
				"request_data": "",
				"status_code": 0,
				"headers": {},
				"keyword": ["name=\"Copyright\" content=\"FineReport\""],
				"priority": 3,
				"favicon_hash": [],
				"name": "finereport"
			}]`,
			want: []finger.WebFingerResult{
				{Name: "finereport", RootPath: "/"},
			},
			err: nil,
		},
	}
	for _, tc := range tests {
		for _, targetURL := range tc.targetURL {
			t.Run(fmt.Sprintf("DoFinger(%s)", targetURL), func(t *testing.T) {
				wfs, err := finger.ParseWebFinger(tc.fingerJSON)
				if err != nil {
					t.Error(err)
					return
				}
				got, err := GetWebFinger(context.Background(), targetURL, *wfs)
				if !utils.ContainsErr(err, tc.err) {
					t.Errorf("error = %v; want %v", err, tc.err)
					return
				}
				if diff := deep.Equal(got, tc.want); diff != nil {
					t.Errorf("got %#v; want %#v; diff: %#v", got, tc.want, diff)
				}
			})
		}
	}
}
