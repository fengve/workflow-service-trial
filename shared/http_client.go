package shared

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPRequester interface {
	Get(ctx context.Context, url string) ([]byte, error)
	GetWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	DeleteWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	PutWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error)
	PostWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error)
	PatchWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error)
	Post(ctx context.Context, url string, request []byte) ([]byte, error)
	PostURLEncoded(ctx context.Context, apiUrl string, data url.Values) ([]byte, error)
}

type HTTPClient struct {
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	return HttpGet(ctx, url)
}

func (c *HTTPClient) GetWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return HttpGetWithHeader(ctx, url, headers)
}

func (c *HTTPClient) DeleteWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return HttpDeleteWithHeader(ctx, url, headers)
}

func (c *HTTPClient) PutWithHeader(ctx context.Context, url string, request []byte,
	headers map[string]string) ([]byte, error) {
	return HttpPutWithHeader(ctx, url, request, headers)
}

func (c *HTTPClient) PostWithHeader(ctx context.Context, url string, request []byte,
	headers map[string]string) ([]byte, error) {
	return HttpPostWithHeader(ctx, url, request, headers)
}

func (c *HTTPClient) PatchWithHeader(ctx context.Context, url string, request []byte,
	headers map[string]string) ([]byte, error) {
	return HttpPatchWithHeader(ctx, url, request, headers)
}

func (c *HTTPClient) Post(ctx context.Context, url string, request []byte) ([]byte, error) {
	return HttpPost(ctx, url, request)
}

func (c *HTTPClient) PostURLEncoded(ctx context.Context, apiUrl string, data url.Values) ([]byte, error) {
	return HttpPostUrlEncoded(ctx, apiUrl, data)
}

func HttpGet(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s\n", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("response failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

func HttpGetWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reqWithTimeout := req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(reqWithTimeout)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s\n", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return bodyBytes, fmt.Errorf("response failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

type HttpResponse struct {
	StatusCode int
	Body       []byte
}

func HttpGetWithHeaderV2(ctx context.Context, url string, headers map[string]string) (*HttpResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reqWithTimeout := req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(reqWithTimeout)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	return &HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
	}, nil
}

// Send http POST or PATCH request with header.
func httpPostOrPatchWithHeader(
	ctx context.Context, url string, method string, request []byte, headers map[string]string,
) ([]byte, error) {

	buffer := bytes.NewBuffer(request)

	// check headers
	if headers == nil {
		headers = make(map[string]string)
	}
	// set content type if headers map does not have it
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}

	return httpPostOrPatchWithHeaderBuffer(ctx, url, method, buffer, headers)
}

// Send http POST or PATCH request with header.
func httpPostOrPatchWithHeaderBuffer(
	ctx context.Context, url string, method string, buffer *bytes.Buffer, headers map[string]string,
) ([]byte, error) {
	req, err := http.NewRequest(method, url, buffer)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reqWithTimeout := req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(reqWithTimeout)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return bodyBytes, fmt.Errorf("response failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

func HttpDeleteWithHeader(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reqWithTimeout := req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(reqWithTimeout)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s\n", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return bodyBytes, fmt.Errorf("response failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}

func HttpPutWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error) {
	return httpPostOrPatchWithHeader(ctx, url, http.MethodPut, request, headers)
}

func HttpPostWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error) {
	return httpPostOrPatchWithHeader(ctx, url, http.MethodPost, request, headers)
}

func HttpPostWithHeaderBuffer(ctx context.Context, url string, request *bytes.Buffer,
	headers map[string]string) ([]byte, error) {
	return httpPostOrPatchWithHeaderBuffer(ctx, url, http.MethodPost, request, headers)
}

func HttpPatchWithHeader(ctx context.Context, url string, request []byte, headers map[string]string) ([]byte, error) {
	return httpPostOrPatchWithHeader(ctx, url, http.MethodPatch, request, headers)
}

func HttpPost(ctx context.Context, url string, request []byte) ([]byte, error) {
	return httpPostOrPatchWithHeader(ctx, url, http.MethodPost, request, nil)
}

func HttpPostUrlEncoded(ctx context.Context, apiUrl string, data url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send http request failed %s\n", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %s", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("response failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	return bodyBytes, nil
}
