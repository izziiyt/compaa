package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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

type CacheEntry struct {
	ETag         string
	LastModified string
	Body         []byte
	Expire       time.Time
}

type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

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

func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, found := c.entries[key]
	if !found {
		return nil, false
	}
	return entry, true
}

func (c *Cache) Set(key string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = entry
}

func (c *Cache) SaveToFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tempFile, err := os.CreateTemp("", "cache-compaa.tmp")
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(tempFile)
	if err := encoder.Encode(c.entries); err != nil {
		return err
	}
	tempFile.Close()

	return os.Rename(tempFile.Name(), filePath)
}

func (c *Cache) LoadFromFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&c.entries)
}

type CacheTransport struct {
	Transport http.RoundTripper
	Cache     *Cache
}

func NewCacheTransport() *CacheTransport {
	cache := NewCache()
	return &CacheTransport{
		Transport: http.DefaultTransport,
		Cache:     cache,
	}
}

func (c *CacheTransport) Close() {
	c.Cache.Close()
}

func (c *CacheTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if entry, found := c.Cache.Get(req.URL.String()); found {
		if entry.Expire.After(time.Now()) {
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(entry.Body)),
			}
			return resp, nil
		}
		if entry.ETag != "" {
			req.Header.Set("If-None-Match", entry.ETag)
		}
		if entry.LastModified != "" {
			req.Header.Set("If-Modified-Since", entry.LastModified)
		}
	}

	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		entry, found := c.Cache.Get(req.URL.String())
		if found {
			resp.Body = io.NopCloser(bytes.NewReader(entry.Body))
			age := extendAge(resp)
			entry.Expire = time.Now().Add(time.Duration(age) * time.Second)
			c.Cache.Set(req.URL.String(), entry)
			resp.StatusCode = http.StatusOK
			return resp, nil
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	etag := resp.Header.Get("ETag")
	lm := resp.Header.Get("Last-Modified")
	age := extendAge(resp)
	if etag != "" || lm != "" || age > 0 {
		entry := &CacheEntry{
			ETag:         etag,
			LastModified: lm,
			Expire:       time.Now().Add(time.Duration(age) * time.Second),
			Body:         body,
		}
		c.Cache.Set(req.URL.String(), entry)
	}

	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

func getMaxAge(cacheControl string) int {
	directives := strings.Split(cacheControl, ",")
	for _, directive := range directives {
		directive = strings.TrimSpace(directive)
		if strings.HasPrefix(directive, "max-age=") {
			ageStr := strings.TrimPrefix(directive, "max-age=")
			age, err := strconv.Atoi(ageStr)
			if err != nil {
				return 0
			}
			return age
		}
	}
	return 0
}

func extendAge(resp *http.Response) int {
	age, err := strconv.Atoi(resp.Header.Get("Age"))
	if err != nil {
		age = 0
	}
	maxage := getMaxAge(resp.Header.Get("Cache-Control"))
	diff := max(maxage-age, 0)
	return diff
}
