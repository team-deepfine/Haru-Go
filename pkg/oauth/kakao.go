package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	kakaoUserMeURL = "https://kapi.kakao.com/v2/user/me"
)

// KakaoUserInfo holds the user info retrieved from Kakao APIs.
type KakaoUserInfo struct {
	Sub          string  // Kakao user ID (int64 → string)
	Email        *string
	Nickname     *string
	ProfileImage *string
}

// KakaoClient fetches user information from Kakao APIs using an access token.
type KakaoClient struct {
	httpClient *http.Client
}

// NewKakaoClient creates a new Kakao OAuth client.
func NewKakaoClient() *KakaoClient {
	return &KakaoClient{
		httpClient: &http.Client{Timeout: 10 * 1e9}, // 10 seconds
	}
}

// GetUserByAccessToken fetches user info from the Kakao API using an access token
// obtained directly from the Kakao SDK (mobile flow).
func (k *KakaoClient) GetUserByAccessToken(ctx context.Context, accessToken string) (*KakaoUserInfo, error) {
	return k.getUserInfo(ctx, accessToken)
}

// getUserInfo fetches user profile from the Kakao user info API.
func (k *KakaoClient) getUserInfo(ctx context.Context, accessToken string) (*KakaoUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kakaoUserMeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kakao user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kakao user info endpoint returned status %d", resp.StatusCode)
	}

	var userResp kakaoUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("decode user info response: %w", err)
	}

	info := &KakaoUserInfo{
		Sub: strconv.FormatInt(userResp.ID, 10),
	}

	if userResp.KakaoAccount != nil {
		if userResp.KakaoAccount.Email != "" && userResp.KakaoAccount.IsEmailVerified {
			info.Email = &userResp.KakaoAccount.Email
		}
		if userResp.KakaoAccount.Profile != nil {
			if userResp.KakaoAccount.Profile.Nickname != "" {
				info.Nickname = &userResp.KakaoAccount.Profile.Nickname
			}
			if userResp.KakaoAccount.Profile.ProfileImageURL != "" {
				info.ProfileImage = &userResp.KakaoAccount.Profile.ProfileImageURL
			}
		}
	}

	return info, nil
}

// Kakao API response types (unexported).

type kakaoUserResponse struct {
	ID           int64         `json:"id"`
	KakaoAccount *kakaoAccount `json:"kakao_account"`
}

type kakaoAccount struct {
	Profile         *kakaoProfile `json:"profile"`
	Email           string        `json:"email"`
	IsEmailVerified bool          `json:"is_email_verified"`
}

type kakaoProfile struct {
	Nickname        string `json:"nickname"`
	ProfileImageURL string `json:"profile_image_url"`
}
