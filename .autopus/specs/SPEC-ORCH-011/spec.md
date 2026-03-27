# SPEC-ORCH-011: opencode/gemini 프롬프트 전달 안정성 수정

**Status**: completed
**Created**: 2026-03-27
**Domain**: ORCH (Orchestra)
**Ref**: BS-010

## Minimal PRD

### 1. Problem

interactive pane 모드에서 opencode와 gemini 프로바이더에 프롬프트가 정상 전달되지 않아 멀티프로바이더 오케스트레이션이 실패한다. opencode는 `SendCommand` 기반 paste를 수신하지 못해 빈 응답을 반환하고, gemini는 긴 프롬프트가 `tmux send-keys` 버퍼 제한으로 잘려 "메시지가 잘렸다"를 반복한다.

### 2. Requirements (P0 only, EARS format)

> **REQ-1**: WHEN interactive pane 모드에서 opencode 프로바이더에 프롬프트를 전달할 때,
> THE SYSTEM SHALL opencode의 입력 메커니즘에 맞는 방식(임시 파일 경유 또는 stdin redirect)으로 프롬프트를 전달하여 정상 응답을 받아야 한다.

> **REQ-2**: WHEN interactive pane 모드에서 2000자 이상의 긴 프롬프트를 gemini 프로바이더에 전달할 때,
> THE SYSTEM SHALL 프롬프트를 청크 단위로 분할하거나 파일 경유 방식으로 전달하여 truncation 없이 전체 프롬프트가 전달되어야 한다.

> **REQ-3**: WHEN 프로바이더별 프롬프트 전달 방식이 다를 때,
> THE SYSTEM SHALL ProviderConfig에 interactive 모드 전용 프롬프트 전달 전략(`interactive_input` 필드)을 지원하여 프로바이더별 맞춤 전달이 가능해야 한다.

> **REQ-4**: WHERE sendPrompts() 또는 executeRound()에서 프롬프트 전달이 실패할 때,
> THE SYSTEM SHALL 에러를 FailedProvider에 기록하고 해당 프로바이더를 skipWait으로 마킹하여 전체 오케스트레이션이 중단되지 않아야 한다.

### 3. Technical Notes

- **tmux send-keys 제한**: tmux send-keys는 단일 호출에 ~500바이트 수준 제한이 있으며, 긴 문자열은 잘릴 수 있음. `load-buffer` + `paste-buffer` 조합으로 우회 가능.
- **cmux send 제한**: cmux send도 유사한 제한이 있을 수 있으나, socket API 기반이므로 tmux보다 나을 수 있음. 확인 필요.
- **opencode 입력 방식**: opencode의 TUI는 `send-keys` 기반 paste를 받지 못할 수 있음. `opencode run -m model "prompt"` 같이 args 직접 전달 또는 파일 경유 필요.
- **기존 코드 영향**: `sendPrompts()`, `executeRound()`, `TmuxAdapter.SendCommand()`, `CmuxAdapter.SendCommand()` 수정 필요.
- **호환성**: claude, codex 등 기존 프로바이더의 SendCommand 기반 전달은 영향 없어야 함.

### 4. Out of Scope

- 프로바이더 자동 감지(바이너리 이름에서 입력 방식 추론) — 수동 설정으로 시작
- Structured Output Protocol (BS-010 ICE #2) — 별도 SPEC
- Circuit Breaker / Graceful Degradation (BS-010 ICE #1) — 별도 SPEC

### 5. Key Q&A

1. **Q: opencode의 정확한 입력 메커니즘은?**
   A: opencode `run` 서브커맨드는 마지막 인자로 프롬프트를 받음. interactive TUI 모드에서는 자체 입력 버퍼를 사용하며, tmux send-keys 기반 paste를 무시할 수 있음. → `buildInteractiveLaunchCmd`에서 `run` 플래그를 제거하는 것이 근본 원인 (line 21).

2. **Q: tmux send-keys의 정확한 버퍼 제한은?**
   A: tmux 소스코드 기준 단일 `send-keys` 호출의 실질적 제한은 터미널 라인 버퍼에 의존하며, ~500-2000바이트에서 잘릴 수 있음. `load-buffer`/`paste-buffer` 조합은 제한 없음.

3. **Q: cmux send도 동일한 truncation 문제가 있는가?**
   A: cmux는 socket API 기반이므로 tmux보다 제한이 적을 가능성이 높으나, 실제 테스트로 확인 필요.

4. **Q: interactive_input 필드의 기본값은?**
   A: `sendkeys`(현재 동작). 대안: `paste-buffer`, `file-redirect`, `args-relaunch`.

5. **Q: `buildInteractiveLaunchCmd`에서 `run` 플래그를 제거하는 것이 opencode 문제의 근본 원인인가?**
   A: 맞음. opencode의 interactive 모드는 `opencode`만 실행하지만, `run` 없이는 TUI가 뜨고 send-keys 입력을 자체 입력 위젯에서 받음. 그러나 TUI 입력 위젯과 tmux send-keys 간 호환이 보장되지 않음. → 대안: non-interactive `opencode run -m model "prompt"` 으로 실행하되 interactive pane에서 출력만 캡처.

## 요구사항

> **R1**: WHEN opencode 프로바이더가 interactive pane 모드에서 실행될 때,
> THE SYSTEM SHALL `buildInteractiveLaunchCmd`에서 opencode를 non-interactive 모드(`opencode run -m <model> "<prompt>"`)로 실행하거나, 프롬프트를 프로바이더별 맞춤 방식으로 전달해야 한다.

> **R2**: WHEN tmux 터미널에서 500바이트 이상의 프롬프트를 전달할 때,
> THE SYSTEM SHALL `send-keys` 대신 `load-buffer`/`paste-buffer` 조합을 사용하여 truncation 없이 전체 프롬프트가 전달되어야 한다.

> **R3**: WHEN cmux 터미널에서 긴 프롬프트를 전달할 때,
> THE SYSTEM SHALL 동일한 truncation 방지 로직을 적용하거나, cmux의 버퍼 제한이 충분함을 검증해야 한다.

> **R4**: WHERE 프로바이더별 프롬프트 전달 방식이 다를 때,
> THE SYSTEM SHALL `sendPrompts()` 및 `executeRound()`에서 프로바이더의 `interactive_input` 설정에 따라 분기 처리해야 한다.

## 생성 파일 상세

| 파일 | 역할 |
|------|------|
| `pkg/terminal/tmux.go` | `SendCommand` → `SendLongText` 추가 (load-buffer/paste-buffer) |
| `pkg/terminal/cmux.go` | 긴 텍스트 전달 검증 및 필요시 분할 로직 추가 |
| `pkg/terminal/terminal.go` | `Terminal` 인터페이스에 `SendLongText` 메서드 추가 |
| `pkg/orchestra/interactive.go` | `sendPrompts()` 수정 — 프로바이더별 분기 |
| `pkg/orchestra/interactive_debate.go` | `executeRound()` 수정 — 동일 분기 적용 |
| `pkg/orchestra/interactive_launch.go` | opencode `run` 플래그 처리 수정 |
| `pkg/orchestra/types.go` | `ProviderConfig`에 `InteractiveInput` 필드 추가 |
| `pkg/config/schema.go` | `ProviderEntry`에 `interactive_input` YAML 필드 추가 |
