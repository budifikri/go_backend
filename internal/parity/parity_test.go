//go:build parity

package parity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func envOrSkip(t *testing.T, key string) string {
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("set %s to run parity tests", key)
	}
	return v
}

func postJSON(t *testing.T, baseURL, path string, body any, token string) (int, map[string]any) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return res.StatusCode, out
}

func getJSON(t *testing.T, baseURL, path string, token string) (int, map[string]any) {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return res.StatusCode, out
}

func getTokenFromLoginResp(t *testing.T, resp map[string]any) string {
	data, ok := resp["data"].(map[string]any)
	if !ok {
		return ""
	}
	tok, _ := data["token"].(string)
	return tok
}

func assertHasKeys(t *testing.T, obj map[string]any, keys ...string) {
	for _, k := range keys {
		if _, ok := obj[k]; !ok {
			t.Fatalf("missing key %q in %v", k, obj)
		}
	}
}

func TestBasicParityShapes(t *testing.T) {
	tsBase := envOrSkip(t, "TS_BASE_URL")
	goBase := envOrSkip(t, "GO_BASE_URL")
	username := envOrSkip(t, "PARITY_USERNAME")
	password := envOrSkip(t, "PARITY_PASSWORD")

	// Login
	statusTS, loginTS := postJSON(t, tsBase, "/api/auth/login", map[string]any{"username": username, "password": password}, "")
	statusGO, loginGO := postJSON(t, goBase, "/api/auth/login", map[string]any{"username": username, "password": password}, "")
	if statusTS != 200 || statusGO != 200 {
		t.Fatalf("login status mismatch/failed: ts=%d go=%d", statusTS, statusGO)
	}

	tokenTS := getTokenFromLoginResp(t, loginTS)
	tokenGO := getTokenFromLoginResp(t, loginGO)
	if tokenTS == "" || tokenGO == "" {
		t.Fatalf("missing token: ts=%q go=%q", tokenTS, tokenGO)
	}

	// Check paginated list shape (products)
	_, prodTS := getJSON(t, tsBase, "/api/products?limit=1&offset=0", tokenTS)
	_, prodGO := getJSON(t, goBase, "/api/products?limit=1&offset=0", tokenGO)
	assertHasKeys(t, prodTS, "success", "data")
	assertHasKeys(t, prodGO, "success", "data")
	if _, ok := prodTS["pagination"]; !ok {
		t.Fatalf("ts missing pagination")
	}
	if _, ok := prodGO["pagination"]; !ok {
		t.Fatalf("go missing pagination")
	}

	// GRN list uses custom pagination shape; just ensure both have pagination.
	_, grnTS := getJSON(t, tsBase, "/api/grn?page=1&limit=1", tokenTS)
	_, grnGO := getJSON(t, goBase, "/api/grn?page=1&limit=1", tokenGO)
	if _, ok := grnTS["pagination"]; !ok {
		t.Fatalf("ts missing grn pagination")
	}
	if _, ok := grnGO["pagination"]; !ok {
		t.Fatalf("go missing grn pagination")
	}

	// Logout (best-effort)
	_, _ = postJSON(t, tsBase, "/api/auth/logout", map[string]any{}, tokenTS)
	_, _ = postJSON(t, goBase, "/api/auth/logout", map[string]any{}, tokenGO)

	_ = fmt.Sprintf("parity ok")
}
