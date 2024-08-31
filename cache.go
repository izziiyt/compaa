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

type CacheEntry struct {
	ETag         string
	LastModified string
	Body         []byte
}

type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.Mutex
}

func NewCache() *Cache {
	c := &Cache{
		entries: make(map[string]*CacheEntry),
	}
	if err := c.LoadFromFile(cacheFile); err != nil {
		fmt.Println("Warn: fails loading cache:", err)
	}
	fmt.Println("Cache loaded", c.entries)
	return c
}

func (c *Cache) Close() {
	if err := c.SaveToFile(cacheFile); err != nil {
		fmt.Println("Warn: fails saving cache:", err)
	}
}

func (c *Cache) Get(key string) (*CacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
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
	if entry, found := c.Cache.Get(req.URL.String()); found {
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

	if resp.StatusCode == http.StatusNotModified {
		if entry, found := c.Cache.Get(req.URL.String()); found {
			resp.Body = io.NopCloser(bytes.NewReader(entry.Body))
			resp.StatusCode = http.StatusOK
			return resp, nil
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	entry := &CacheEntry{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Body:         body,
	}
	c.Cache.Set(req.URL.String(), entry)

	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}
