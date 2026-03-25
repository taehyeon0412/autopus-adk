---
name: go
stack: go
tools: [go, golangci-lint, gopls]
test_framework: go test
linter: golangci-lint
---

# Go Executor Profile

## Idioms

Write idiomatic Go. Prefer clarity over cleverness.

### Error Handling

```go
// Always return errors explicitly; never ignore them
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething: %w", err)
}
```

### Interface-Driven Design

```go
// Define interfaces at the point of use, not at the point of implementation
type Reader interface {
    Read(ctx context.Context, id string) (*Entity, error)
}
```

### Struct Embedding

```go
// Use embedding for behavior composition, not inheritance
type Service struct {
    repo    Repository
    logger  *slog.Logger
}
```

### Context Propagation

```go
// Always pass context as the first parameter
func (s *Service) Get(ctx context.Context, id string) (*Entity, error) {
    return s.repo.Find(ctx, id)
}
```

## Testing Patterns

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    t.Parallel()
    cases := []struct {
        name    string
        a, b    int
        want    int
    }{
        {"positive", 1, 2, 3},
        {"zero", 0, 0, 0},
        {"negative", -1, 1, 0},
    }
    for _, tc := range cases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            got := Add(tc.a, tc.b)
            if got != tc.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
            }
        })
    }
}
```

### Use `t.Parallel()` for independent tests to speed up the suite.

### Subtests for grouped assertions

```go
t.Run("group", func(t *testing.T) {
    // assertions for this group
})
```

## Completion Criteria

- [ ] `go test -race ./...` — all tests pass, no data races
- [ ] `go vet ./...` — no issues
- [ ] `golangci-lint run` — no warnings
- [ ] Coverage 85%+
- [ ] `go build ./...` — compiles cleanly
