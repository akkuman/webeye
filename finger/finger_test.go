package finger

import "testing"

func TestMatchFavicon(t *testing.T) {
	wf := WebFinger{
		Name: "test",
		MatchRules: MatchRule{
			FaviconHash: []string{},
		},
	}
	if wf.MatchFavicon(nil) == true {
		t.Error("wf.MatchFavicon(nil) == true")
		return
	}
	if wf.MatchFavicon([]string{}) == true {
		t.Error("wf.MatchFavicon([]string{}) == true")
		return
	}
}