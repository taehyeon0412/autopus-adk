# SPEC-ORCH-004 리서치

## 기존 코드 분석

### 전략 등록 구조

**`pkg/orchestra/types.go`**
- `Strategy` 타입: `string` 기반 (L11)
- 기존 상수: `StrategyConsensus`, `StrategyPipeline`, `StrategyDebate`, `StrategyFastest` (L13-18)
- `ValidStrategies` 슬라이스: 유효성 검증에 사용 (L21)
- `IsValid()` 메서드: `ValidStrategies` 순회로 검증 (L24-31)
- `OrchestraConfig` 구조체: `KeepRelayOutput` 필드를 추가할 위치 (L69-79)

**`pkg/orchestra/strategy.go`**
- `strategyHandlers` 맵: 전략별 후처리 핸들러 등록 (L13-18)
- `StrategyFunc` 타입: `func(ctx, responses, cfg) (string, string, error)` (L10)
- 새 전략 추가 시 이 맵에 `StrategyRelay: handleRelay` 엔트리 추가

**`pkg/orchestra/runner.go`**
- `RunOrchestra` 함수: 메인 진입점 (L17)
- 실행 전략 분기: switch문 (L50-62) — `StrategyRelay` 케이스 추가 필요
- 병합 처리 분기: switch문 (L71-84) — relay 포맷팅 케이스 추가 필요
- `runProvider` 함수: 단일 프로바이더 실행 (L221) — relay에서 재사용

### 순차 실행 패턴 (Pipeline 참조)

**`pkg/orchestra/runner.go:runPipeline`** (L162-178)
- 순차 실행 + 이전 출력을 다음 프롬프트에 추가하는 패턴
- relay와 유사하나 차이점:
  1. relay는 agentic 플래그 추가 필요
  2. relay는 파일 기반 결과 저장
  3. relay는 구조화된 맥락 주입 (단순 append가 아닌 `## Previous Analysis` 섹션)

### Provider 실행 구조

**`pkg/orchestra/runner.go:runProvider`** (L221-289)
- `provider.Args`를 복사 후 사용 (L225)
- `PromptViaArgs` 분기: true면 Args 마지막에 프롬프트 추가, false면 stdin으로 전달
- `newCommand(ctx, provider.Binary, args...)`: testable command 생성
- relay에서는 Args에 agentic 플래그를 추가하여 `runProvider`를 그대로 호출 가능

### CLI 통합 구조

**`internal/cli/orchestra.go`**
- `runOrchestraCommand` 함수 (L127): 공통 실행 로직
- `IsValid()` 검증 (L170): ValidStrategies에 relay가 등록되면 자동 통과
- 프로바이더 기본 Args (L224-231): claude `-p`, codex `-q`, gemini `-p`

**`internal/cli/orchestra_config.go`**
- `resolveStrategy`: config의 strategy 필드에 "relay" 지정 가능 (L87-101)
- 추가 수정 불필요 — config 레벨에서 "relay"를 string으로 지정하면 자동 반영

### Temp 파일 관리 패턴

**`pkg/orchestra/pane_runner.go`**
- `os.MkdirTemp` 사용 패턴 (output 파일 저장)
- sentinel 기반 완료 감지 — relay에서는 불필요 (동기 실행)

**`pkg/orchestra/job.go`** (존재 확인)
- `CleanupStaleJobs` 함수: `/tmp` 기반 정리 패턴 참조
- relay의 temp dir도 유사한 패턴으로 정리

## 프로바이더별 Agentic 모드 플래그

| 프로바이더 | 기본 Args | Agentic 추가 플래그 | 설명 |
|-----------|----------|-------------------|------|
| claude | `-p` | `--allowedTools Edit,Read,Bash,Write` | Claude Code의 도구 사용 활성화 |
| codex | `-q` | `--approval-mode full-auto` | Codex의 전자동 승인 모드 |
| gemini | `-p` | (없음) | Gemini CLI는 기본적으로 도구 접근 가능 |

> Note: 정확한 `--allowedTools` 값은 구현 시 Claude Code의 최신 CLI 스펙을 확인해야 한다. 위 값은 BS-001 브레인스톰 결과 기반.

## 설계 결정

### D1: runPipeline 재사용 vs 별도 함수

**결정**: `runRelay`를 별도 함수로 구현

**이유**: pipeline과 relay의 핵심 차이점이 3가지(agentic 플래그, 파일 저장, 구조화된 프롬프트)이므로 runPipeline을 수정하면 기존 동작에 영향을 줄 수 있다. 별도 함수로 격리하여 하위 호환성을 보장한다.

**대안 검토**:
- runPipeline에 옵션 파라미터 추가 → 기존 코드 복잡도 증가, 테스트 영향 우려
- 공통 순차 실행 함수 추출 → 과도한 추상화, 두 함수의 차이가 충분히 크다

### D2: Provider Args 수정 방식

**결정**: `runRelay` 내부에서 provider.Args를 복사 후 agentic 플래그 append

**이유**: `runProvider`는 `provider.Args`를 `append([]string{}, provider.Args...)`로 복사하므로 (L225), relay에서 Args를 수정해도 원본 ProviderConfig에는 영향이 없다. 그러나 명시적으로 relay 전용 복사를 수행하여 의도를 명확히 한다.

### D3: 결과 파일 형식

**결정**: `/tmp/autopus-relay-{jobID}/{provider}.md` 형식의 마크다운 파일

**이유**: 마크다운은 후속 프로바이더가 직접 읽을 수 있는 범용 형식이며, 디버깅 시에도 유용하다. jobID는 `crypto/rand`로 생성하여 충돌을 방지한다 (pane_runner.go의 기존 패턴 참조).

### D4: 실패 시 동작

**결정**: fail-fast — 실패한 프로바이더 이후의 프로바이더는 건너뛰고 부분 결과 반환

**이유**: relay는 이전 결과에 의존하는 순차 체인이므로, 중간 실패 시 이후 프로바이더에 불완전한 맥락이 전달되는 것보다 부분 결과를 반환하는 것이 낫다. pipeline 전략의 기존 동작(L168-170)과도 일관성을 유지한다.
