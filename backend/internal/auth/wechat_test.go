package auth

import (
	"encoding/json"
	"testing"
)

func TestWechatAccessTokenResp(t *testing.T) {
	t.Run("success_response", func(t *testing.T) {
		raw := `{"access_token":"at","openid":"oid","unionid":"uid","errcode":0,"errmsg":"ok"}`
		var resp WechatAccessTokenResp
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.AccessToken != "at" {
			t.Errorf("AccessToken = %q, want at", resp.AccessToken)
		}
		if resp.OpenID != "oid" {
			t.Errorf("OpenID = %q, want oid", resp.OpenID)
		}
		if resp.UnionID != "uid" {
			t.Errorf("UnionID = %q, want uid", resp.UnionID)
		}
		if resp.ErrCode != 0 {
			t.Errorf("ErrCode = %d, want 0", resp.ErrCode)
		}
	})

	t.Run("error_response", func(t *testing.T) {
		raw := `{"errcode":40029,"errmsg":"invalid code"}`
		var resp WechatAccessTokenResp
		if err := json.Unmarshal([]byte(raw), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.ErrCode != 40029 {
			t.Errorf("ErrCode = %d, want 40029", resp.ErrCode)
		}
		if resp.ErrMsg != "invalid code" {
			t.Errorf("ErrMsg = %q, want 'invalid code'", resp.ErrMsg)
		}
	})

	t.Run("is_error", func(t *testing.T) {
		tests := []struct {
			errCode int
			isErr   bool
		}{
			{0, false},
			{40029, true},
			{40163, true},
			{40013, true},
		}
		for _, tt := range tests {
			resp := WechatAccessTokenResp{ErrCode: tt.errCode}
			if (resp.ErrCode != 0) != tt.isErr {
				t.Errorf("ErrCode=%d isErr=%v, want %v", tt.errCode, resp.ErrCode != 0, tt.isErr)
			}
		}
	})
}

func TestWechatUserInfo(t *testing.T) {
	raw := `{"openid":"oid","nickname":"测试","headimgurl":"https://img.url","errcode":0}`
	var info WechatUserInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.OpenID != "oid" {
		t.Errorf("OpenID = %q, want oid", info.OpenID)
	}
	if info.Nickname != "测试" {
		t.Errorf("Nickname = %q, want 测试", info.Nickname)
	}
	if info.HeadImgURL != "https://img.url" {
		t.Errorf("HeadImgURL = %q, want https://img.url", info.HeadImgURL)
	}
}

func TestWechatErrorResponse(t *testing.T) {
	raw := `{"errcode":40125,"errmsg":"invalid appsecret"}`
	var info WechatUserInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.ErrCode != 40125 {
		t.Errorf("ErrCode = %d, want 40125", info.ErrCode)
	}
	if info.ErrMsg != "invalid appsecret" {
		t.Errorf("ErrMsg = %q", info.ErrMsg)
	}
}
