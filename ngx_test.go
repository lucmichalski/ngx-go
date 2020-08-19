package ngx

import (
	"reflect"
	"testing"
)

var positiveMarshal = []struct {
	Data     Access
	Expected string
}{
	{Access{RemoteAddr: "$remote_addr", RemoteUser: "$remote_user", TimeLocal: "$time_local", Request: "$request", Status: 200, BodyBytesSent: 0, HTTPReferer: "$http_referer", HTTPUserAgent: "$http_user_agent"}, "$remote_addr - $remote_user [$time_local] \"$request\" 200 0 \"$http_referer\" \"$http_user_agent\""},
}

var positiveUnmarshal = []struct {
	Fmt       string
	Data      string
	Expected  map[string]string
	Marshaled string
}{
	{CombinedFmt, CombinedFmt, map[string]string{"remote_addr": "$remote_addr", "remote_user": "$remote_user", "time_local": "$time_local", "request": "$request", "status": "$status", "body_bytes_sent": "$body_bytes_sent", "http_referer": "$http_referer", "http_user_agent": "$http_user_agent"}, CombinedFmt},
	{`\$request\$request_body\$header_cookie\`, `\request\request_body\header_cookie\`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}, `\request\request_body\header_cookie\`},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\request\"request_body\"\"header_cookie\"`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}, `\request\"request_body\"\"header_cookie\"`},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\requ\\\"est\"request_body\"\"header_cookie\"`, map[string]string{"request": "requ\\\"est", "request_body": "request_body", "header_cookie": "header_cookie"}, `\requ\\\"est\"request_body\"\"header_cookie\"`},
	{`escape=json;{"$key":"$value"}`, `{"$key":"$value"}`, map[string]string{"key": "$key", "value": "$value"}, `{"$key":"$value"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065y":"\r\f\t\uf755\n"}`, map[string]string{"key": "$key", "value": "\r\f\t\xef\x9d\x95\n"}, "{\"$key\":\"\\r\\f\\t\uf755\\n\"}"},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09"}`, map[string]string{"key": "$key", "value": "🌉"}, `{"$key":"🌉"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"surrogate pair : \ud83c\udf09"}`, map[string]string{"key": "$key", "value": "surrogate pair : 🌉"}, `{"$key":"surrogate pair : 🌉"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09\ud83c\udf09is\u0020surrogate\u0020pair"}`, map[string]string{"key": "$key", "value": "🌉🌉is surrogate pair"}, `{"$key":"🌉🌉is surrogate pair"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09\ud83c\udf09\ud83c\udf09\ud83c\udf09\""}`, map[string]string{"key": "$key", "value": "🌉🌉🌉🌉\""}, `{"$key":"🌉🌉🌉🌉\""}`},
	{`escape=json;{"$$$key":"$$$value"}`, `{"$key":"$value"}`, map[string]string{"key": "key", "value": "value"}, `{"$key":"$value"}`},
}

func TestMarshal(t *testing.T) {
	for _, tc := range positiveMarshal {
		got, err := MarshalToString(tc.Data)
		if err != nil {
			t.Fatalf("failed to marshal data %q: %v", tc.Data, err)
		}
		if got != tc.Expected {
			t.Fatalf("corrupted data in marshal: expecting %q, got %q", tc.Expected, got)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range positiveUnmarshal {
		ngx, err := Compile(tc.Fmt)
		if err != nil {
			t.Fatalf("failed to compile format %q: %v", tc.Fmt, err)
		}

		got := make(map[string]string)
		if err := ngx.UnmarshalFromString(tc.Data, &got); err != nil {
			t.Fatalf("failed to unmarshal data %q: %v", tc.Data, err)
		}

		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in unmarshal: expecting %q, got %q", tc.Expected, got)
		}

		marshaled, err := ngx.MarshalToString(got)
		if err != nil {
			t.Fatalf("failed to marshal data %q: %v", got, err)
		}
		if marshaled != tc.Marshaled {
			t.Fatalf("corrupted data in marshal: expecting %q, got %q", tc.Marshaled, marshaled)
		}
	}
}

func BenchmarkUnmarshalFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]string)
		if err := UnmarshalFromString(CombinedFmt, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(CombinedFmt)
	for i := 0; i < b.N; i++ {
		m := make(map[string]string)
		if err := Unmarshal(data, &m); err != nil {
			b.Fatal(err)
		}
	}
}
