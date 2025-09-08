package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"garmin-connect/garth/types"
)

var oauthConsumer *types.OAuthConsumer

// LoadOAuthConsumer loads OAuth consumer credentials
func LoadOAuthConsumer() (*types.OAuthConsumer, error) {
	if oauthConsumer != nil {
		return oauthConsumer, nil
	}

	// First try to get from S3 (like the Python library)
	resp, err := http.Get("https://thegarth.s3.amazonaws.com/oauth_consumer.json")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			var consumer types.OAuthConsumer
			if err := json.NewDecoder(resp.Body).Decode(&consumer); err == nil {
				oauthConsumer = &consumer
				return oauthConsumer, nil
			}
		}
	}

	// Fallback to hardcoded values
	oauthConsumer = &types.OAuthConsumer{
		ConsumerKey:    "fc320c35-fbdc-4308-b5c6-8e41a8b2e0c8",
		ConsumerSecret: "8b344b8c-5bd5-4b7b-9c98-ad76a6bbf0e7",
	}
	return oauthConsumer, nil
}

// GenerateNonce generates a random nonce for OAuth
func GenerateNonce() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// GenerateTimestamp generates a timestamp for OAuth
func GenerateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// PercentEncode URL encodes a string
func PercentEncode(s string) string {
	return url.QueryEscape(s)
}

// CreateSignatureBaseString creates the base string for OAuth signing
func CreateSignatureBaseString(method, baseURL string, params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStrs []string
	for _, key := range keys {
		paramStrs = append(paramStrs, PercentEncode(key)+"="+PercentEncode(params[key]))
	}
	paramString := strings.Join(paramStrs, "&")

	return method + "&" + PercentEncode(baseURL) + "&" + PercentEncode(paramString)
}

// CreateSigningKey creates the signing key for OAuth
func CreateSigningKey(consumerSecret, tokenSecret string) string {
	return PercentEncode(consumerSecret) + "&" + PercentEncode(tokenSecret)
}

// SignRequest signs an OAuth request
func SignRequest(consumerSecret, tokenSecret, baseString string) string {
	signingKey := CreateSigningKey(consumerSecret, tokenSecret)
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// CreateOAuth1AuthorizationHeader creates the OAuth1 authorization header
func CreateOAuth1AuthorizationHeader(method, requestURL string, params map[string]string, consumerKey, consumerSecret, token, tokenSecret string) string {
	oauthParams := map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            GenerateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        GenerateTimestamp(),
		"oauth_version":          "1.0",
	}

	if token != "" {
		oauthParams["oauth_token"] = token
	}

	// Combine OAuth params with request params
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	// Parse URL to get base URL without query params
	parsedURL, _ := url.Parse(requestURL)
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	// Create signature base string
	baseString := CreateSignatureBaseString(method, baseURL, allParams)

	// Sign the request
	signature := SignRequest(consumerSecret, tokenSecret, baseString)
	oauthParams["oauth_signature"] = signature

	// Build authorization header
	var headerParts []string
	for key, value := range oauthParams {
		headerParts = append(headerParts, PercentEncode(key)+"=\""+PercentEncode(value)+"\"")
	}
	sort.Strings(headerParts)

	return "OAuth " + strings.Join(headerParts, ", ")
}

// Min returns the smaller of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DateRange generates a date range from end date backwards for n days
func DateRange(end time.Time, days int) []time.Time {
	dates := make([]time.Time, days)
	for i := 0; i < days; i++ {
		dates[i] = end.AddDate(0, 0, -i)
	}
	return dates
}

// CamelToSnake converts a camelCase string to snake_case
func CamelToSnake(s string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// CamelToSnakeDict recursively converts map keys from camelCase to snake_case
func CamelToSnakeDict(m map[string]interface{}) map[string]interface{} {
	snakeDict := make(map[string]interface{})
	for k, v := range m {
		snakeKey := CamelToSnake(k)
		// Handle nested maps
		if nestedMap, ok := v.(map[string]interface{}); ok {
			snakeDict[snakeKey] = CamelToSnakeDict(nestedMap)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			// Handle slices of maps
			var newSlice []interface{}
			for _, item := range nestedSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					newSlice = append(newSlice, CamelToSnakeDict(itemMap))
				} else {
					newSlice = append(newSlice, item)
				}
			}
			snakeDict[snakeKey] = newSlice
		} else {
			snakeDict[snakeKey] = v
		}
	}
	return snakeDict
}

// FormatEndDate converts various date formats to time.Time
func FormatEndDate(end interface{}) time.Time {
	if end == nil {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}

	switch v := end.(type) {
	case string:
		t, _ := time.Parse("2006-01-02", v)
		return t
	case time.Time:
		return v
	default:
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
}

// GetLocalizedDateTime converts GMT and local timestamps to localized time
func GetLocalizedDateTime(gmtTimestamp, localTimestamp int64) time.Time {
	localDiff := localTimestamp - gmtTimestamp
	offset := time.Duration(localDiff) * time.Millisecond
	loc := time.FixedZone("", int(offset.Seconds()))
	gmtTime := time.Unix(0, gmtTimestamp*int64(time.Millisecond)).UTC()
	return gmtTime.In(loc)
}
