package xiaotao

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	keyId           string
	accessKeySecret []byte
	domainYunfang   string
)

// Init keyID, accessKey, domain
func Init(keyID string, accessKey []byte, domain string) error {
	keyId = keyID
	accessKeySecret = accessKey
	domainYunfang = domain
	// no error
	return nil
}

// GetEnquiryPriceResponseData holds the data of GetEnquiryPriceResponse.
type GetEnquiryPriceResponseData struct {
	Price         float64
	TotalPrice    float64
	StatusCode    string
	CommunityID   string `json:"communityID"`
	CommunityName string
}

// GetEnquiryPriceResponse defines the response of GetEnquiryPrice.
type GetEnquiryPriceResponse struct {
	Data      []GetEnquiryPriceResponseData
	ErrorCode string
	GUID      string `json:"guid"`
	Message   string
	Success   bool
}

// GetCommunityPeripheryAvgeragePriceResponseData holds the data of GetCommunityPeripheryAvgeragePriceResponse.
type GetCommunityPeripheryAvgeragePriceResponseData struct {
	List []GetCommunityPeripheryAvgeragePriceResponseDataEntry
}

// GetCommunityPeripheryAvgeragePriceResponseDataEntry holds avgerage price for one community.
type GetCommunityPeripheryAvgeragePriceResponseDataEntry struct {
	AvgPrice      float64 `json:"avgPrice"`
	CommunityName string  `json:"communityName"`
	Distance      float64 `json:"distance"`
	XAmap         float64 `json:"xAmap"`
	XBaidu        float64 `json:"xBaidu"`
	YAmap         float64 `json:"yAmap"`
	YBaidu        float64 `json:"yBaidu"`
}

// GetCommunityPeripheryAvgeragePriceResponse defines the response of GetCommunityPeripheryAvgeragePrice.
type GetCommunityPeripheryAvgeragePriceResponse struct {
	Data      []GetCommunityPeripheryAvgeragePriceResponseData
	ErrorCode string
	GUID      string `json:"guid"`
	Message   string
	Success   bool
}

// GetCommunityRatingResponseData holds the data of GetCommunityRatingResponse.
type GetCommunityRatingResponseData struct {
	RankWithinDistrict float64
	RankWithinCity     float64
	CommunityGrade     string
	CommunityID        string `json:"communityID"`
	CommunityName      string
}

// GetCommunityRatingResponse defines the response of GetCommunityRating.
type GetCommunityRatingResponse struct {
	Data      []GetCommunityRatingResponseData
	ErrorCode string
	GUID      string `json:"guid"`
	Message   string
	Success   bool
}

// Get signs the request and issues a GET to the specified URL.
func Get(rawurl string) (*http.Response, error) {
	u, err := url.Parse(domainYunfang + rawurl)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	now := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)

	q.Set("userKeyId", keyId)
	q.Set("time", now)

	var keys []string
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString("&")
		b.WriteString(urlEncode(k))
		b.WriteString("=")
		b.WriteString(urlEncode(q.Get(k)))
	}

	plain := "Post&" + urlEncode("/") + "&" + urlEncode(b.String()[1:])

	mac := hmac.New(sha1.New, accessKeySecret)
	mac.Write([]byte(plain))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return http.Get(domainYunfang + rawurl + "&userKeyId=" + keyId + "&timeStamp=" + now + "&accessSignature=" + urlEncode(signature))
}

// url中的特殊字符进行转义
func urlEncode(s string) string {
	if s == "" {
		return s
	}

	s = url.QueryEscape(s)

	// 仅仅依靠url.QueryEscape(s)是无法拿到云房正确的数字签名的，比如传入的字符串中有"/"、"+"等特殊字符时
	// 经过参考云房接口文档及自己的尝试，发现需要做如下的replace操作才能在各种情况下拿到我们想要的正确的云房数字签名
	s = strings.ReplaceAll(s, "+", "%2b")
	s = strings.ReplaceAll(s, "*", "%2A")
	s = strings.ReplaceAll(s, "%7E", "~")

	return s
}
