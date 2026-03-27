# SPEC-ORCH-004 구현 계획

## 태스크 목록

- [ ] T1: `types.go`에 `StrategyRelay` 상수 및 `ValidStrategies` 확장
  - `StrategyRelay Strategy = "relay"` 추가
  - `ValidStrategies` 슬라이스에 relay 추가
  - `OrchestraConfig`에 `KeepRelayOutput bool` 필드 추가

- [ ] T2: `pkg/orchestra/relay.go` 신규 생성 — 핵심 실행 로직
  - `runRelay(ctx, cfg) ([]ProviderResponse, error)`: 순차 실행 + 파일 저장 + 프롬프트 주입
  - `buildRelayPrompt(original, previousResults)`: 이전 결과를 포함한 릴레이 프롬프트 구성
  - `agenticArgs(provider) []string`: 프로바이더별 agentic 도구 플래그 반환
  - `FormatRelay(responses) string`: 릴레이 결과 포맷팅
  - `cleanupRelayDir(dir string, keep bool)`: temp 디렉토리 정리

- [ ] T3: `strategy.go`에 relay 핸들러 등록
  - `strategyHandlers` 맵에 `StrategyRelay: handleRelay` 추가
  - `handleRelay` 함수 구현

- [ ] T4: `runner.go`의 `RunOrchestra` switch문에 relay 케이스 추가
  - `case StrategyRelay:` 블록에서 `runRelay` 호출
  - 병합 처리 switch문에도 relay 케이스 추가

- [ ] T5: `internal/cli/orchestra.go` CLI 통합
  - strategy 도움말 문자열에 "relay" 추가
  - `--keep-relay-output` 플래그 추가 (review, plan, secure, brainstorm)
  - `OrchestraConfig.KeepRelayOutput` 매핑

- [ ] T6: `pkg/orchestra/relay_test.go` 유닛 테스트
  - 순차 실행 순서 검증
  - 프롬프트 주입 내용 검증
  - agentic 플래그 매핑 검증
  - temp 디렉토리 생성/정리 검증
  - 기존 전략 비간섭 검증

## 구현 전략

### 기존 코드 활용

- **`runPipeline` 패턴 참조**: relay는 pipeline과 유사한 순차 실행이나, 차이점은 (1) agentic 플래그 추가 (2) 파일 기반 결과 저장 (3) 구조화된 맥락 주입이다.
- **`runProvider` 재사용**: 단일 프로바이더 실행은 기존 `runProvider` 함수를 그대로 사용한다. agentic 플래그는 provider의 Args를 확장하여 전달한다.
- **`command` 인터페이스**: 테스트 시 `newCommand`를 mock하여 실제 바이너리 실행 없이 검증한다.

### 변경 범위

| 파일 | 변경 유형 | 예상 라인 |
|------|----------|----------|
| `types.go` | 수정 (상수 + 필드 추가) | +5줄 |
| `strategy.go` | 수정 (핸들러 등록) | +10줄 |
| `runner.go` | 수정 (switch 케이스) | +10줄 |
| `relay.go` | **신규** | ~150줄 |
| `relay_test.go` | **신규** | ~180줄 |
| `orchestra.go` | 수정 (플래그 + 도움말) | +15줄 |

### 설계 원칙

1. **relay.go 단일 파일**: 핵심 로직을 한 파일에 집중 (300줄 제한 내)
2. **Args 확장 방식**: provider.Args를 복사 후 agentic 플래그를 append하여 원본 불변
3. **fail-fast**: 릴레이 중 한 프로바이더가 실패하면 이후 프로바이더를 건너뛰고 부분 결과 반환
