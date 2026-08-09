package engine

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
)

// TemplateEngine manages compiled html/template instances with thread-safe RWMutex caching.
type TemplateEngine struct {
	cache map[string]*template.Template
	mu    sync.RWMutex
}

// NewTemplateEngine initializes a new TemplateEngine instance.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		cache: make(map[string]*template.Template),
	}
}

// Render compiles and renders template text (subject or body) with dynamic variables (e.g. {{.Name}}).
func (te *TemplateEngine) Render(key string, tmplContent string, data interface{}) (string, error) {
	te.mu.RLock()
	compiledTmpl, exists := te.cache[key]
	te.mu.RUnlock()

	if !exists {
		var err error
		compiledTmpl, err = template.New(key).Parse(tmplContent)
		if err != nil {
			return "", fmt.Errorf("failed to parse template '%s': %w", key, err)
		}

		te.mu.Lock()
		te.cache[key] = compiledTmpl
		te.mu.Unlock()
	}

	var buf bytes.Buffer
	if err := compiledTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", key, err)
	}

	return buf.String(), nil
}
