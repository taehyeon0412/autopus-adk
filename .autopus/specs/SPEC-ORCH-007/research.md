# SPEC-ORCH-007 리서치

## 기존 코드 분석

### 결과 수집 현재 흐름 (SPEC-ORCH-006)

현재 `interactive.go`의 결과 수집 경로:

```
waitAndCollectResults()
  → waitForCompletion() [interactive.go:207]
    → ReadScreen 폴링 (500ms) + isPromptVisible() [interactive_detect.go:59]
    → isOutputIdle() (10초 idle) [interactive_detect.go:77]
  → ReadScreen(Scrollback: true) [interactive.go:187]
  → cleanScreenOutput() [interactive_detect.go:87]
    → stripANSI() [interactive_detect.go:14]
    → filterPromptLines() [interactive_detect.go:29]
```

핵심 문제:
- `waitForCompletion()`: ReadScreen 폴링 + idle 감지의 이중 전략이 false positive/negative를 모두 유발
- `cleanScreenOutput()`: ANSI 이스케이프 + 프롬프트 패턴 필터링이 프로바이더마다 다른 출력 형식에 대응 불가
- `DefaultCompletionPatterns()` [types.go:89]: 하드코딩된 3개 패턴만 존재

### Hook 모드로 대체될 코드 경로

| 현재 함수 | 위치 | Hook 모드 대체 |
|-----------|------|---------------|
| `waitForCompletion()` | interactive.go:207 | `waitForHookDone()` — done 파일 os.Stat 폴링 |
| `cleanScreenOutput()` | interactive_detect.go:87 | 제거 — JSON response 필드 직접 사용 |
| `isPromptVisible()` | interactive_detect.go:59 | fallback 전용으로 격리 |
| `isOutputIdle()` | interactive_detect.go:77 | fallback 전용으로 격리 |
| `pollUntilPrompt()` | interactive.go:122 | 세션 준비 감지에는 유지 (hook은 완료 시그널만) |

### 유지되는 코드 경로

- `startPipeCapture()` [interactive.go:81]: 디버깅/로깅용으로 유지
- `launchInteractiveSessions()` [interactive.go:92]: pane에 바이너리 실행은 동일
- `waitForSessionReady()` [interactive.go:111]: 세션 준비는 ReadScreen 유지 (hook은 완료만 감지)
- `sendPrompts()` [interactive.go:146]: 프롬프트 주입은 동일
- `mergeByStrategy()` [pane_runner.go:252]: 병합 로직은 동일 (입력이 깨끗해짐)

### 세션 ID 관리

기존 `randomHex()` [pane_runner.go:286]: 8자 hex 생성. Hook 모드에서는 16자(randomHex() x2)로 세션 ID를 생성하여 `/tmp/autopus/{session-id}/` 디렉토리를 관리한다. 이미 `runRelay()`에서 동일 패턴 사용 중 [relay.go:23].

### 프로바이더 설정 파일 위치

| Provider | 설정 파일 | Hook 등록 위치 |
|----------|----------|---------------|
| Claude Code | `.claude/settings.json` | `hooks.Stop` 배열 |
| Gemini CLI | `.gemini/settings.json` | `hooks.AfterAgent` 배열 |
| opencode | `opencode.json` | `experimental.text.complete` plugin |

### 어댑터별 Hook 주입 포인트

- **Claude**: `pkg/adapter/claude/claude_settings.go` — 기존 `mergeHooks()` 함수가 사용자 hook을 보존하면서 autopus hook을 추가하는 로직 구현 완료. Stop hook 엔트리 추가 필요.
- **Gemini**: `pkg/adapter/gemini/gemini.go` — 현재 hook 관련 코드 없음. AfterAgent hook 주입 로직 신규 추가 필요.
- **opencode**: `pkg/adapter/opencode/` — 디렉토리 미존재. 신규 어댑터 생성 필요. `PlatformAdapter` 인터페이스 [adapter.go] 구현.

### Config 마이그레이션 분석

`pkg/config/migrate.go`:
- `defaultProviderEntries`: claude, codex, gemini 3개만 정의 [line 5-9]
- `PlatformToProvider()`: claude-code, codex, gemini-cli 3개 매핑 [line 110-121]
- opencode 엔트리 추가: `"opencode": {Binary: "opencode", Args: []string{}, PaneArgs: []string{}}`
- codex → opencode 마이그레이션 함수 추가 필요

### 전략별 Hook 연동 분석

**Debate** [debate.go]:
- `runDebate()` [line 14]: Phase 1에서 `runParallel()` → 인터랙티브 모드에서는 pane 병렬 실행
- `runRebuttalRound()` [line 52]: `prevResponses[].Output` 사용 → hook 결과로 대체 시 Output이 깨끗한 JSON response가 됨
- `buildRebuttalPrompt()` [line 94]: `r.Output` 직접 참조 → hook 모드에서 이미 깨끗한 텍스트이므로 변경 불필요

**Relay** [relay.go]:
- `buildRelayPrompt()` [line 88]: `r.output` 사용 → hook 결과로 자동 대체
- `runRelay()`: 비인터랙티브 모드에서는 `runProvider()` 직접 호출 — hook 미적용
- `relay_pane.go`: 인터랙티브 relay에서 hook 적용 필요

**Consensus** [consensus.go]:
- `MergeConsensus()`: `responses[].Output` 직접 비교 — hook 결과가 Output에 들어가므로 변경 불필요

핵심 발견: `ProviderResponse.Output`에 hook 결과를 저장하면 debate/relay/consensus 전략 코드 변경이 최소화된다.

## 프로바이더별 Hook API 분석

### Claude Code Stop Hook

```json
// .claude/settings.json
{
  "hooks": {
    "Stop": [{
      "matcher": "",
      "hooks": [{
        "type": "command",
        "command": "/path/to/hook-claude-stop.sh"
      }]
    }]
  }
}
```

- 입력: stdin으로 JSON 전달 (`last_assistant_message` 필드 포함)
- 트리거: 모든 assistant 응답 완료 시
- fire-and-forget: hook 실패가 CLI 동작에 영향 없음

### Gemini CLI AfterAgent Hook

```json
// .gemini/settings.json
{
  "hooks": {
    "AfterAgent": [{
      "command": "/path/to/hook-gemini-afteragent.sh"
    }]
  }
}
```

- 입력: stdin으로 JSON 전달 (`prompt_response` 필드 포함)
- 트리거: 에이전트 턴 완료 시

### opencode Plugin

```json
// opencode.json
{
  "experimental": {
    "plugins": [{
      "name": "autopus-result",
      "event": "text.complete",
      "command": "bun /path/to/hook-opencode-complete.ts"
    }]
  }
}
```

- 입력: 환경변수 또는 stdin으로 `text` 필드 전달
- 트리거: 텍스트 완성 완료 시
- TypeScript 기반 (bun 런타임)

## 설계 결정

### D1: 파일 폴링 vs fsnotify

**결정**: `os.Stat()` 폴링 (200ms 간격)

**근거**:
- fsnotify는 외부 의존성 추가 필요 (프로젝트 정책: stdlib 우선)
- 200ms 폴링은 NFR 목표(500ms 이내 감지)를 충족
- PRD 기술 제약에서 명시적으로 폴링 우선 지정 [PRD Section 7]
- 3개 프로바이더 x 200ms = 무시할 수 있는 CPU 오버헤드

### D2: 프로바이더별 파일 분리 vs 단일 파일

**결정**: 프로바이더별 파일 분리 (`{provider}-result.json`, `{provider}-done`)

**근거**:
- 동시 3 프로바이더 쓰기 충돌 방지 (PRD Q&A Q6)
- 개별 프로바이더 완료 감지 가능 (fastest 전략에 필수)
- 디버깅 시 프로바이더별 결과 독립 확인 가능

### D3: Hook 모드 활성화 방식

**결정**: 프로바이더별 hook 설정 존재 여부를 자동 감지하여 혼합 모드 지원

**근거**:
- Graceful degradation 요구사항 (R8)
- 사용자가 일부 프로바이더만 hook 설정할 수 있음
- `OrchestraConfig.HookMode`는 전체 플래그, 실제 실행은 프로바이더별 분기

### D4: Hook 결과를 ProviderResponse.Output에 저장

**결정**: hook 결과의 `response` 필드를 기존 `ProviderResponse.Output`에 저장

**근거**:
- debate/relay/consensus 전략 코드 변경 최소화 (리서치 분석 결과)
- `mergeByStrategy()`, `MergeConsensus()`, `buildRebuttalPrompt()` 등이 모두 `.Output` 참조
- 새 필드 추가보다 기존 인터페이스 활용이 호환성 유지에 유리

### D5: Shell 스크립트 vs Go 바이너리 Hook

**결정**: POSIX shell 스크립트 (Claude/Gemini), TypeScript (opencode)

**근거**:
- 프로바이더 CLI가 hook을 외부 명령으로 실행 — Go 바이너리 불필요
- Shell 스크립트는 의존성 없이 즉시 실행 가능
- opencode plugin은 TypeScript 기반 — bun 런타임 활용
- PRD 기술 제약: `jq` 의존성 없이 기본 shell 도구만 사용

### D6: /tmp vs XDG_CACHE_HOME

**결정**: `/tmp/autopus/{session-id}/` (PRD Q5 미결정 상태에서 /tmp 우선 사용)

**근거**:
- 세션 임시 데이터이므로 캐시가 아닌 임시 저장소가 적합
- macOS/Linux 모두 `/tmp` 존재 보장
- 오케스트레이션 완료 후 자동 삭제 (NFR)
- XDG는 향후 Q5 결정에 따라 마이그레이션 가능

## 위험 분석

### Hook 스크립트의 jq-free JSON 파싱

Shell에서 `jq` 없이 JSON 필드를 추출해야 한다. 방법:
- `grep` + `sed` 조합으로 단일 필드 추출
- Python one-liner 활용 (macOS/Linux 기본 설치): `python3 -c "import json,sys; print(json.load(sys.stdin)['field'])"`
- Go 바이너리 내 JSON 추출 헬퍼 (`auto hook-extract`) 제공

권장: Python3 one-liner 방식 (macOS/Linux 양쪽 기본 설치, 안정적 JSON 파싱)

### opencode experimental API 안정성

opencode의 `experimental.text.complete`는 실험적 API이다. 버전 핀(>=0.1)으로 호환성 관리하고, API 변경 시 hook 스크립트 템플릿만 업데이트하면 된다.
