package lsp

// Commander는 LSP 커맨드 인터페이스이다.
type Commander interface {
	Diagnostics(path string) ([]Diagnostic, error)
	References(symbol string) ([]Location, error)
	Rename(oldName, newName string) error
	Symbols(path string) ([]Symbol, error)
	Definition(symbol string) (*Location, error)
}

// MockClient는 테스트용 모의 LSP 클라이언트이다.
type MockClient struct {
	diagnostics map[string][]Diagnostic
	refs        map[string][]Location
	symbols     map[string][]Symbol
	definitions map[string]*Location
}

// NewMockClient는 모의 클라이언트를 생성한다.
func NewMockClient(diagnostics []Diagnostic) *MockClient {
	c := &MockClient{
		diagnostics: make(map[string][]Diagnostic),
		refs:        make(map[string][]Location),
		symbols:     make(map[string][]Symbol),
		definitions: make(map[string]*Location),
	}

	// 진단 메시지를 파일별로 인덱싱
	for _, d := range diagnostics {
		c.diagnostics[d.File] = append(c.diagnostics[d.File], d)
	}

	return c
}

// SetRefs는 심볼별 참조 목록을 설정한다.
func (c *MockClient) SetRefs(symbol string, locs []Location) {
	c.refs[symbol] = locs
}

// SetSymbols는 파일별 심볼 목록을 설정한다.
func (c *MockClient) SetSymbols(path string, syms []Symbol) {
	c.symbols[path] = syms
}

// SetDefinition는 심볼별 정의 위치를 설정한다.
func (c *MockClient) SetDefinition(symbol string, loc *Location) {
	c.definitions[symbol] = loc
}

// Diagnostics는 파일의 진단 메시지를 반환한다.
func (c *MockClient) Diagnostics(path string) ([]Diagnostic, error) {
	return c.diagnostics[path], nil
}

// References는 심볼의 참조 위치 목록을 반환한다.
func (c *MockClient) References(symbol string) ([]Location, error) {
	return c.refs[symbol], nil
}

// Rename는 심볼을 이름 변경한다.
func (c *MockClient) Rename(_, _ string) error {
	return nil
}

// Symbols는 파일의 심볼 목록을 반환한다.
func (c *MockClient) Symbols(path string) ([]Symbol, error) {
	return c.symbols[path], nil
}

// Definition는 심볼의 정의 위치를 반환한다.
func (c *MockClient) Definition(symbol string) (*Location, error) {
	loc, ok := c.definitions[symbol]
	if !ok {
		return nil, nil
	}
	return loc, nil
}
