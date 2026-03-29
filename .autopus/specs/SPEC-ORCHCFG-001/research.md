# SPEC-ORCHCFG-001 리서치

## 기존 코드 분석

### 합의 임계값 현재 상태

**하드코딩 위치** (`pkg/orchestra/strategy.go:33`):
```go
func handleConsensus(_ context.Context, responses []ProviderResponse, _ OrchestraConfig) (string, string, error) {
    merged, summary := MergeConsensus(responses, 0.66)  // 하드코딩
    return merged, summary, nil
}
```

- `handleConsensus`는 `cfg OrchestraConfig` 파라미터를 이미 받고 있으나, `_`로 무시함
- 이 한 줄만 수정하면 핵심 기능이 동작함

**OrchestraConfig.ConsensusThreshold** (`pkg/orchestra/types.go:84`):
- 이미 `ConsensusThreshold float64` 필드가 존재
- 주석: "consensus threshold (0 uses default 0.66)"
- `interactive_debate_helpers.go:85`에서 `cfg.ConsensusThreshold`로 사용 중

**MergeConsensus** (`pkg/orchestra/consensus.go`):
- `MergeConsensus(responses []ProviderResponse, threshold float64)` — 이미 threshold 파라미터를 받음
- `MergeStructuredConsensus`도 동일

### resolve 패턴 (CLI 설정 해석)

**기존 3종 함수** (`internal/cli/orchestra_config.go` 또는 관련 파일):
- `resolveProviders(conf, commandName, flagProviders)` — CLI → 커맨드 → 전체
- `resolveStrategy(conf, commandName, flagStrategy)` — CLI → 커맨드 → 기본값 → "consensus"
- `resolveJudge(conf, commandName, flagJudge)` — CLI → 커맨드 → ""

테스트 (`internal/cli/orchestra_config_test.go`):
- 각 resolve 함수에 대해 Flag 오버라이드, 커맨드 설정, 기본값 폴백 테스트가 존재
- 동일한 패턴으로 `resolveThreshold` 테스트를 추가하면 됨

### config.CommandEntry (`pkg/config/schema.go:114-119`)

```go
type CommandEntry struct {
    Strategy  string   `yaml:"strategy"`
    Providers []string `yaml:"providers,flow"`
    Judge     string   `yaml:"judge,omitempty"`
}
```

- `ConsensusThreshold float64 \`yaml:"consensus_threshold,omitempty"\`` 추가 필요
- YAML 태그만 추가하면 기존 파싱 자동 동작

### autopus.yaml 현재 commands 섹션

```yaml
commands:
    brainstorm:
        strategy: debate
        providers: [claude, opencode, gemini]
        judge: claude
    plan:
        strategy: consensus
        providers: [claude, opencode, gemini]
    review:
        strategy: debate
        providers: [claude, opencode, gemini]
        judge: claude
    secure:
        strategy: consensus
        providers: [claude, opencode, gemini]
```

- `consensus_threshold` 필드가 아직 없음 (추가 대상)

## 설계 결정

### 왜 resolve 패턴인가

기존 `resolveStrategy`, `resolveProviders`, `resolveJudge`와 동일한 3단계 폴백 패턴을 사용한다. 이유:
1. 코드 일관성 — 같은 파일, 같은 패턴, 같은 테스트 구조
2. 학습 비용 최소화 — 기존 패턴을 아는 개발자가 즉시 이해 가능
3. 테스트 용이 — 기존 테스트 패턴 복사 후 값만 변경

### 왜 4단계 폴백인가 (기존 3단계 + 전역 값)

기존 resolve 함수들은 CLI → 커맨드설정 → 기본값의 3단계이다. 임계값은 `OrchestraConfig.ConsensusThreshold`라는 전역 설정이 이미 존재하므로, CLI → 커맨드별 → 전역 → 기본값의 4단계가 자연스럽다.

### 대안 검토

1. **전역 설정만 사용** — 현재 상태. 모든 커맨드에 동일한 임계값이라 보안/브레인스토밍 구분 불가. 기각.
2. **전략(strategy)에 임계값 내장** — strategy.go에서 전략별 기본 임계값. 커맨드 단위 커스터마이징 불가. 기각.
3. **커맨드별 설정 (채택)** — autopus.yaml commands 섹션에 `consensus_threshold` 추가. 기존 패턴과 일관. 최소 코드 변경.

### 리스크

- **없음**: 하위 호환성 보장 (0이면 기본값 폴백)
- **저위험**: YAML 필드 추가는 기존 파일 파싱에 영향 없음 (`omitempty` 태그)
