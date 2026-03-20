// Package template은 Go text/template 기반 렌더링 엔진을 제공한다.
package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// Engine은 템플릿 렌더링 엔진이다.
type Engine struct {
	funcMap template.FuncMap
}

// New는 기본 FuncMap으로 엔진을 생성한다.
func New() *Engine {
	return &Engine{
		funcMap: defaultFuncMap(),
	}
}

// RenderString은 문자열 템플릿을 렌더링한다.
func (e *Engine) RenderString(tmpl string, data any) (string, error) {
	t, err := template.New("").Funcs(e.funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// RenderFile은 파일 템플릿을 렌더링한다.
// ParseFiles는 파일명을 템플릿 이름으로 사용하므로 basename으로 ExecuteTemplate을 호출한다.
func (e *Engine) RenderFile(path string, data any) (string, error) {
	name := filepath.Base(path)
	t, err := template.New(name).Funcs(e.funcMap).ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("parse template file %s: %w", path, err)
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("execute template file %s: %w", path, err)
	}
	return buf.String(), nil
}

func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"indent": func(n int, s string) string {
			pad := strings.Repeat(" ", n)
			lines := strings.Split(s, "\n")
			for i, l := range lines {
				if l != "" {
					lines[i] = pad + l
				}
			}
			return strings.Join(lines, "\n")
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"trim":  strings.TrimSpace,
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
		"contains": strings.Contains,
		"langName": func(code string) string {
			names := map[string]string{
				"en": "English",
				"ko": "Korean",
				"ja": "Japanese",
				"zh": "Chinese",
			}
			if name, ok := names[code]; ok {
				return name
			}
			return code
		},
	}
}
