/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
// Package curl 提供HTTP请求的封装工具
package curl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// PostRequest 发送POST请求
func PostRequest(tokenStr, url string, requestBody any) ([]byte, error) {
	return sendRequest(tokenStr, url, "POST", requestBody)
}

// DeleteRequest 发送DELETE请求
func DeleteRequest(tokenStr, url string, requestBody any) ([]byte, error) {
	return sendRequest(tokenStr, url, "DELETE", nil)
}

// PutRequest 发送PUT请求
func PutRequest(tokenStr, url string, requestBody any) ([]byte, error) {
	return sendRequest(tokenStr, url, "PUT", requestBody)
}

// GetRequest 发送GET请求，支持查询参数
func GetRequest(tokenStr, urls string, queryParams map[string]any) ([]byte, error) {
	// 拼接查询参数
	if len(queryParams) > 0 {
		values := url.Values{}
		for key, value := range queryParams {
			values.Add(key, fmt.Sprintf("%v", value))
		}
		urls = urls + "?" + values.Encode()
	}
	return sendRequest(tokenStr, urls, "GET", nil)
}

// sendRequest 统一的HTTP请求发送方法
func sendRequest(tokenStr, url, method string, requestBody any) ([]byte, error) {
	var (
		body []byte
		err  error
	)

	// 序列化请求体
	if requestBody != nil {
		body, err = json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if len(tokenStr) > 0 {
		req.Header.Set("Authorization", tokenStr)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBuffer := new(bytes.Buffer)
	_, err = responseBuffer.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBuffer.Bytes(), nil
}
