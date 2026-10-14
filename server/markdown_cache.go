package server

import (
	"container/list"
	"sync"
)

var renderedMarkdown = newMarkdownCache(64 << 20)

func renderMarkdownCached(src string) string {
	if html, ok := renderedMarkdown.get(src); ok {
		return html
	}
	html := renderMarkdown(src)
	renderedMarkdown.put(src, html)
	return html
}

type markdownEntry struct {
	src  string
	html string
}

type markdownCache struct {
	mu       sync.Mutex
	maxBytes int
	bytes    int
	order    *list.List
	entries  map[string]*list.Element
}

func newMarkdownCache(maxBytes int) *markdownCache {
	return &markdownCache{
		maxBytes: maxBytes,
		order:    list.New(),
		entries:  map[string]*list.Element{},
	}
}

func (x *markdownCache) get(src string) (string, bool) {
	x.mu.Lock()
	defer x.mu.Unlock()

	el, ok := x.entries[src]
	if !ok {
		return "", false
	}
	x.order.MoveToFront(el)
	return el.Value.(markdownEntry).html, true
}

func (x *markdownCache) put(src, html string) {
	size := len(src) + len(html)
	if size > x.maxBytes {
		return
	}

	x.mu.Lock()
	defer x.mu.Unlock()

	if _, ok := x.entries[src]; ok {
		return
	}
	x.entries[src] = x.order.PushFront(markdownEntry{src: src, html: html})
	x.bytes += size

	for x.bytes > x.maxBytes {
		oldest := x.order.Back()
		entry := x.order.Remove(oldest).(markdownEntry)
		delete(x.entries, entry.src)
		x.bytes -= len(entry.src) + len(entry.html)
	}
}
