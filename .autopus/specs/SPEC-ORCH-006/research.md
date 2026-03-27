# SPEC-ORCH-006 리서치

## 기존 코드 분석

### Terminal 인터페이스 (`pkg/terminal/terminal.go`)
- 현재 5개 메서드: `Name()`, `CreateWorkspace()`, `SplitPane()`, `SendCommand()`, `Notify()`, `Close()`
- `@AX:ANCHOR` 주석이 있어 변경 시 모든 어댑터 반영 필요
- `PaneID` 타입은 `string` 기반으로 cmux의 `surface:N`, tmux의 pane ID 모두 수용

### CmuxAdapter (`pkg/terminal/cmux.go`, 125줄)
- `execCommand()` 헬퍼로 외부 명령 실행
- `parseCmuxRef()` — cmux CLI 출력에서 ref 추출
- `SendCommand` — `cmux send --surface <ref> <command>` 실행
- 새 메서드 추가 시 약 160줄 예상 (300줄 이내)

### TmuxAdapter (`pkg/terminal/tmux.go`, 88줄)
- `a.session` 필드로 세션명 추적, pane 타겟은 `session:paneID` 형식
- `SendCommand` — `tmux send-keys -t <target> <command> Enter`
- 새 메서드 추가 시 약 130줄 예상 (300줄 이내)

### PlainAdapter (`pkg/terminal/plain.go`, 32줄)
- 모든 메서드 no-op
- 새 메서드 3개 추가해도 약 50줄

### pane_runner.go (`pkg/orchestra/pane_runner.go`, 285줄)
- `RunPaneOrchestra()` — pane 기반 오케스트레이션 진입점
- `splitProviderPanes()` — pane 분할 + 임시 파일 생성. **재사용 대상**
- `sendPaneCommands()` — `buildPaneCommand()`로 sentinel 기반 명령 구성. **인터랙티브 모드에서 우회**
- `collectPaneResults()` — sentinel 폴링으로 결과 수집. **인터랙티브 모드에서 대체**
- `mergeByStrategy()` — 전략별 merge. **그대로 재사용**
- `cleanupPanes()` — pane 닫기 + 임시 파일 삭제. **그대로 재사용**
- `paneArgs()` — `PaneArgs` 우선 반환, 없으면 `Args` 반환. **재사용**
- 현재 285줄로 300줄 근접. 인터랙티브 로직은 반드시 별도 파일로 분리

### pane_shell.go (`pkg/orchestra/pane_shell.go`, 57줄)
- `shellEscapeArg()`, `shellEscapeArgs()` — 셸 이스케이프 유틸리티
- `sanitizeProviderName()` — 프로바이더명 경로 안전화
- **그대로 재사용**

### types.go (`pkg/orchestra/types.go`, 78줄)
- `ProviderConfig.PaneArgs` — 이미 존재하는 pane 모드 전용 인자 필드
- `OrchestraConfig` — `Interactive bool` 필드 추가 필요
- 프로바이더별 완료 패턴 설정이 필요하면 `ProviderConfig`에 `DonePattern string` 추가

### runner.go (`pkg/orchestra/runner.go`, 297줄)
- `RunOrchestra()` — terminal이 있으면 `RunPaneOrchestra()`로 위임
- 기존 비대화식 모드 경로는 변경 불필요

### autopus.yaml (root, 122줄)
- `orchestra.providers` 섹션에 `binary`, `args`, `prompt_via_args` 설정
- `pane_args` 필드는 yaml에 아직 없지만 `ProviderConfig` 구조체에는 이미 존재

## 설계 결정

### 1. 인터랙티브 로직 분리

**결정**: `interactive.go` + `interactive_detect.go` 2개 파일로 분리
**근거**: `pane_runner.go`가 이미 285줄로 300줄 근접. 인터랙티브 로직을 같은 파일에 넣으면 초과 확실. 감지 유틸과 실행 플로우를 분리하면 각 파일 200줄 이내 유지 가능.
**대안**: `pane_runner.go`를 리팩터링하여 기존 코드와 함께 넣기 → 300줄 초과로 불가

### 2. 완료 감지: ReadScreen 폴링 vs pipe-pane 파일 감시

**결정**: ReadScreen 폴링을 primary, pipe-pane 파일 idle 감지를 secondary로 사용
**근거**: ReadScreen은 ANSI 자동 제거(cmux)되어 깨끗한 텍스트 매칭 가능. pipe-pane 파일은 raw 바이트가 섞일 수 있어 패턴 매칭이 불안정. 단, ReadScreen 폴링 간격이 너무 짧으면 성능 문제가 있으므로 1-2초 간격 권장.
**대안**: sentinel 마커를 인터랙티브 세션에도 적용 → CLI 세션 내에서 임의 명령 실행이 어렵고, CLI마다 출력 포맷이 달라 sentinel 삽입 불가

### 3. Terminal 인터페이스 확장 방식

**결정**: 기존 `Terminal` 인터페이스에 직접 메서드 추가
**근거**: 현재 어댑터가 3개뿐이고 모두 이 프로젝트 내에서 관리됨. 인터페이스 분리(composition)보다 단순 확장이 변경 비용이 낮음.
**대안**: `InteractiveTerminal` 별도 인터페이스로 타입 단언(type assertion) 사용 → 호출측 코드가 복잡해짐

### 4. pane_args 활용 전략

**결정**: 인터랙티브 모드에서는 `pane_args`가 비어있으면 바이너리만 실행 (인자 없음 = 인터랙티브 모드)
**근거**: claude, codex, gemini 모두 인자 없이 실행하면 인터랙티브 모드로 진입. 기존 `pane_args` 필드를 그대로 활용하면 설정 스키마 변경 없음.

### 5. cmux `read-screen` ANSI 처리

**결정**: cmux는 ANSI 자동 제거 기능이 있으므로 cmux 어댑터에서는 추가 스트립 불필요. tmux `capture-pane -p`는 raw 출력이므로 ANSI 스트립 유틸 적용.
**근거**: cmux 문서에 `read-screen`이 ANSI 자동 제거한다고 명시. tmux는 `-e` 플래그 없이 `capture-pane`하면 raw 텍스트지만 일부 제어 시퀀스가 남을 수 있음.
