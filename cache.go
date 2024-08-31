package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var cacheFile string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cacheDir := filepath.Join(home, ".local", "share", "compaa")
	cacheFile = filepath.Join(cacheDir, ".cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Println("Warn: fails creating cache directory:", err)
	}
}

// キャッシュエントリを保持する構造体
type CacheEntry struct {
	ETag         string
	LastModified string
	Body         []byte
}

// キャッシュを保持する構造体
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.Mutex
}

// 新しいキャッシュを作成
func NewCache() *Cache {
	c := &Cache{
		entries: make(map[string]*CacheEntry),
	}
	if err := c.LoadFromFile(cacheFile); err != nil {
		fmt.Println("Warn: fails loading cache:", err)
	}
	return c
}

func (c *Cache) Close() {
	if err := c.SaveToFile(cacheFile); err != nil {
		fmt.Println("Warn: fails saving cache:", err)
	}
}

// キャッシュからエントリを取得
func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, found := c.entries[key]
	if !found {
		return nil, false
	}
	return entry, true
}

// キャッシュにエントリを追加
func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry
}

// キャッシュをファイルに保存
func (c *Cache) SaveToFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 一時ファイルを利用して安全に保存
	tempFile, err := os.CreateTemp("", "cache-compaa.tmp")
	if err != nil {
		return err
	}

	// キャッシュを一時ファイルにエンコード
	encoder := gob.NewEncoder(tempFile)
	if err := encoder.Encode(c.entries); err != nil {
		return err
	}
	tempFile.Close()

	// 一時ファイルを最終的なキャッシュファイルにリネーム
	return os.Rename(tempFile.Name(), filePath)
}

// ファイルからキャッシュを読み込む
func (c *Cache) LoadFromFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // ファイルが存在しない場合はエラーではない
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&c.entries)
}

// カスタムトランスポート
type CacheTransport struct {
	Transport http.RoundTripper
	Cache     *Cache
}

// 新しいカスタムトランスポートを作成
func NewCacheTransport(transport http.RoundTripper) *CacheTransport {
	cache := NewCache()
	return &CacheTransport{
		Transport: transport,
		Cache:     cache,
	}
}

func (c *CacheTransport) Close() {
	c.Cache.Close()
}

func (c *CacheTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// キャッシュをチェック
	if entry, found := c.Cache.Get(req.URL.String()); found {
		// ETagとLast-Modifiedヘッダーを設定
		if entry.ETag != "" {
			req.Header.Set("If-None-Match", entry.ETag)
		}
		if entry.LastModified != "" {
			req.Header.Set("If-Modified-Since", entry.LastModified)
		}
	}

	// 通常のリクエストを行う
	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// ステータスコードが304の場合、キャッシュからレスポンスを返す
	if resp.StatusCode == http.StatusNotModified {
		if entry, found := c.Cache.Get(req.URL.String()); found {
			resp.Body = io.NopCloser(bytes.NewReader(entry.Body))
			resp.StatusCode = http.StatusOK
			return resp, nil
		}
	}

	// レスポンスボディを読み取る
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// キャッシュに保存
	entry := &CacheEntry{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Body:         body,
	}
	c.Cache.Set(req.URL.String(), entry)

	// 新しいレスポンスを作成して返す
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}
