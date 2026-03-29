# SPEC-ORCHCFG-001 구현 계획

## 태스크 목록

- [ ] T1: `config.CommandEntry`에 `ConsensusThreshold float64` 필드 추가 (`pkg/config/schema.go`)
- [ ] T2: `handleConsensus()`에서 `cfg.ConsensusThreshold` 전달 (`pkg/orchestra/strategy.go`)
- [ ] T3: `resolveThreshold()` 함수 구현 (`internal/cli/orchestra_config.go`)
- [ ] T4: CLI `--threshold` 플래그 등록 및 `resolveThreshold` 호출 (`internal/cli/orchestra.go` 또는 관련 커맨드 파일)
- [ ] T5: 임계값 범위 검증 로직 추가 (0.0 < threshold <= 1.0)
- [ ] T6: `resolveThreshold` 단위 테스트 (`internal/cli/orchestra_config_test.go`)
- [ ] T7: `handleConsensus` cfg 전달 테스트 (`pkg/orchestra/strategy_test.go`)

## 구현 전략

### 기존 코드 활용

이 기능은 기존 패턴을 최대한 활용한다:

1. **resolve 패턴 재사용**: `resolveStrategy()`, `resolveProviders()`, `resolveJudge()` 함수가 이미 "CLI 플래그 → 커맨드 설정 → 기본값" 3단계 폴백을 구현하고 있다. `resolveThreshold()`도 동일한 패턴을 따른다.

2. **ConsensusThreshold 필드 활용**: `OrchestraConfig.ConsensusThreshold`가 이미 존재하며 `interactive_debate_helpers.go`의 `consensusReached()`에서 사용 중이다. `handleConsensus()`만 이 필드를 읽도록 수정하면 된다.

3. **CommandEntry 확장**: `config.CommandEntry`에 필드 하나만 추가하면 YAML 파싱이 자동으로 동작한다.

### 변경 범위

- 순수 추가 변경 (기존 동작 변경 없음 — 0이면 기본값 0.66 폴백)
- `strategy.go`의 `handleConsensus` 함수 1줄 수정 (0.66 → cfg.ConsensusThreshold 전달 + 기본값 처리)
- CLI 관련 파일에 `--threshold` 플래그와 `resolveThreshold` 추가
- 테스트 파일에 검증 케이스 추가

### 의존성

- T1은 독립적으로 실행 가능 (스키마 변경)
- T2는 독립적으로 실행 가능 (이미 `cfg OrchestraConfig` 매개변수를 받고 있음)
- T3~T4는 T1 완료 후 진행
- T5는 T3~T4와 함께 구현
- T6~T7은 해당 구현 태스크 완료 후 진행
