# SPEC-ORCH-004: Orchestra Agentic Relay Mode

**Status**: completed
**Created**: 2026-03-26
**Domain**: ORCH

## 목적

현재 orchestra의 모든 전략(consensus, pipeline, debate, fastest)은 프로바이더를 단순 텍스트 입출력 서브프로세스로 실행한다. 프로바이더가 코드 탐색이나 도구 사용 권한 없이 실행되므로 분석 깊이가 제한적이고, 프로바이더 간 맥락 공유가 없어 독립적 결과만 생성된다.

Relay 전략은 각 프로바이더를 **도구 접근 권한이 포함된 agentic one-shot 모드**로 순차 실행하되, 이전 프로바이더의 분석 결과를 다음 프로바이더의 프롬프트에 주입하는 릴레이 패턴을 구현한다. 이를 통해:

1. 프로바이더가 파일 시스템, 코드 탐색 등 도구를 직접 사용할 수 있다
2. 이전 프로바이더의 분석 결과가 다음 프로바이더의 맥락으로 전달된다
3. 순차적 깊이 증가(depth-first relay)를 통해 점진적으로 정제된 결과를 얻는다

## 요구사항

### REQ-1: Relay Strategy 등록

WHEN the system initializes the strategy registry,
THE SYSTEM SHALL register `StrategyRelay` ("relay") in `types.go`의 `ValidStrategies` 목록과 `strategy.go`의 `strategyHandlers` 맵에 등록하여 기존 4개 전략과 동등한 수준으로 사용 가능하게 한다.

### REQ-2: Agentic Provider Flags

WHEN a provider is executed under the relay strategy,
THE SYSTEM SHALL 각 프로바이더 CLI에 도구 접근 권한 플래그를 추가하여 agentic one-shot 모드로 실행한다.

프로바이더별 agentic 플래그 매핑:

| Provider | Binary | Base Mode | Agentic Flags | Notes |
|----------|--------|-----------|---------------|-------|
| claude | `claude` | `-p "{prompt}"` | `--allowedTools "Read,Grep,Bash,Glob"` | 도구 목록은 릴레이 전용 — Edit 제외 (읽기 전용) |
| codex | `codex` | `"{prompt}"` | `--approval-mode full-auto --quiet` | full-auto는 도구 사용 자동 승인 |
| gemini | `gemini` | `"{prompt}"` | (없음) | gemini CLI는 기본적으로 도구 접근 가능, 추가 플래그 불필요 |

WHERE a provider binary does not support agentic flags (감지 불가 시), THE SYSTEM SHALL 기존 one-shot 모드로 fallback하고 stderr에 경고를 출력한다.

### REQ-3: Sequential Relay Execution

WHEN the relay strategy is selected,
THE SYSTEM SHALL 프로바이더를 설정된 순서대로 순차 실행하며, 각 프로바이더의 stdout 결과를 `/tmp/autopus-relay-{jobID}/{provider}.md` 파일에 저장한다. jobID는 기존 `randomHex(8)` 함수를 사용하여 생성한다 (detach 모드와 동일한 유틸).

### REQ-3a: Relay Error Handling

WHEN a provider in the relay chain fails (non-zero exit code or timeout),
THE SYSTEM SHALL 해당 프로바이더를 건너뛰고(skip) 다음 프로바이더로 계속 진행한다. 실패한 프로바이더의 결과는 빈 문자열로 처리되며, 최종 결과에 `[SKIPPED: {provider} — {error}]` 마커를 포함한다.
WHERE all providers fail, THE SYSTEM SHALL 에러를 반환한다.

### REQ-4: Context Injection

WHEN a provider other than the first is executed in relay mode,
THE SYSTEM SHALL 이전 프로바이더의 결과 파일 내용을 현재 프로바이더의 프롬프트에 `## Previous Analysis by {provider}` 섹션으로 주입한다.

### REQ-5: CLI Activation

WHEN the user specifies `--strategy relay` or config에 `strategy: relay`를 설정한 경우,
THE SYSTEM SHALL relay 전략을 활성화하고, 기존 orchestra 서브커맨드(review, plan, secure, brainstorm)에서 동일하게 사용 가능하게 한다.

### REQ-6: Backward Compatibility

WHILE the relay strategy is added,
THE SYSTEM SHALL 기존 4개 전략(consensus, pipeline, debate, fastest)의 동작에 어떠한 영향도 주지 않는다.

### REQ-7: Temp Directory Management

WHEN relay execution completes (성공 또는 실패),
THE SYSTEM SHALL `/tmp/autopus-relay-{jobID}/` 디렉토리를 정리한다. 단, `--keep-relay-output` 플래그가 설정된 경우 결과 파일을 보존한다.

### REQ-8: Relay Result Formatting

WHEN all relay providers have completed execution,
THE SYSTEM SHALL 각 단계의 결과를 `## Relay Stage N: (by {provider})` 형식으로 병합하고, 최종 프로바이더의 결과를 primary output으로 표시한다.

### REQ-9: Pipeline과의 차별화

Relay 전략은 기존 pipeline 전략과 다음 점에서 차별화된다:
- **도구 접근**: relay는 프로바이더에 agentic 플래그를 부여하여 코드 탐색/도구 사용 가능. pipeline은 텍스트 only.
- **맥락 주입 형식**: relay는 `## Previous Analysis by {provider}` 구조화 섹션으로 전체 이전 결과를 주입. pipeline은 이전 응답을 프롬프트 끝에 단순 append.
- **결과 파일 저장**: relay는 각 단계를 파일로 저장하여 디버깅/감사 가능. pipeline은 메모리 only.

### REQ-10: Pane Mode 지원 (Phase 2)

Relay 전략의 pane mode(cmux/tmux 인터랙티브 세션) 지원은 이 SPEC의 범위 밖이다. 초기 구현은 standard execution만 지원하며, pane relay는 후속 SPEC(SPEC-ORCH-005)으로 분리한다.
WHERE the terminal is pane-capable AND strategy is relay, THE SYSTEM SHALL standard execution으로 fallback하고 "relay pane mode not yet supported" 경고를 출력한다.

## 생성 파일 상세

| 파일 | 위치 | 역할 |
|------|------|------|
| `relay.go` | `pkg/orchestra/relay.go` | relay 전략의 핵심 실행 로직 (`runRelay`, `buildRelayPrompt`, `FormatRelay`) |
| `relay_test.go` | `pkg/orchestra/relay_test.go` | relay 전략 유닛 테스트 |
| `types.go` 수정 | `pkg/orchestra/types.go` | `StrategyRelay` 상수 추가, `ValidStrategies` 확장, `OrchestraConfig`에 `KeepRelayOutput` 필드 추가 |
| `strategy.go` 수정 | `pkg/orchestra/strategy.go` | `strategyHandlers`에 relay 핸들러 등록 |
| `runner.go` 수정 | `pkg/orchestra/runner.go` | `RunOrchestra`의 switch문에 `StrategyRelay` 케이스 추가 |
| `orchestra.go` 수정 | `internal/cli/orchestra.go` | strategy 도움말에 "relay" 추가, `--keep-relay-output` 플래그 추가 |
