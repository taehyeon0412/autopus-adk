# SPEC-ORCH-003 리서치

## 기존 코드 분석

### Pane 오케스트레이션 흐름 (현재)
**파일**: `pkg/orchestra/pane_runner.go`
- `RunPaneOrchestra()` (line 30): 현재 진입점. Terminal이 nil이거나 plain이면 `RunOrchestra()`로 폴백
- `splitProviderPanes()` (line 79): 각 프로바이더에 대해 `Terminal.SplitPane()` + `os.CreateTemp()` 호출. detach 모드에서 그대로 재사용 가능
- `sendPaneCommands()` (line 103): 각 pane에 `Terminal.SendCommand()` 호출. 실패 시 `skipWait=true` 마킹. 재사용 가능
- `collectPaneResults()` (line 120): goroutine으로 `waitForSentinel()` 병렬 호출 — 이것이 blocking의 원인. detach 모드에서는 이 호출을 건너뜀
- `waitForSentinel()` (line 195): 500ms ticker로 sentinel 파일 polling. CheckStatus/Wait에서 재사용
- `hasSentinel()` (line 212): 파일을 열어 sentinel 마커 검색. CheckStatus에서 직접 호출
- `readOutputFile()` (line 229): 파일 읽기 + sentinel 제거. CollectResults에서 재사용
- `mergeByStrategy()` (line 239): 전략별 merge. CollectResults에서 재사용
- `cleanupPanes()` (line 263): pane 종료 + temp 파일 삭제. Cleanup에서 재사용
- `randomHex()` (line 273): 8char hex 생성. Job ID 생성에 재사용

### Plain 오케스트레이션 (현재)
**파일**: `pkg/orchestra/runner.go`
- `RunOrchestra()` (line 17): non-pane terminal일 때 진입점. Terminal이 non-plain이면 `RunPaneOrchestra()`로 위임
- 상호 위임 구조: `RunOrchestra` ↔ `RunPaneOrchestra` (terminal 타입에 따라)

### Terminal 인터페이스
**파일**: `pkg/terminal/terminal.go`
- `Terminal` interface: Name(), SplitPane(), SendCommand(), Close() 등
- `PaneID` type: string alias

**파일**: `pkg/terminal/cmux.go`
- `CmuxAdapter`: cmux CLI wrapper
- `Close()` (line 83): surface:N 또는 workspace:N ref로 pane/workspace 종료

**파일**: `pkg/terminal/detect.go`
- `DetectTerminal()`: cmux → tmux → plain 순서로 감지

### CLI 구조
**파일**: `internal/cli/orchestra.go`
- `newOrchestraCmd()` (line 31): 서브커맨드 트리 루트. review, plan, secure, brainstorm 등록
- `runOrchestraCommand()` (line 135): 모든 서브커맨드의 공통 실행 함수. 여기에 auto-detach 분기를 추가
- line 188: `Terminal: terminal.DetectTerminal()` — 현재 감지된 terminal을 config에 설정
- line 197: `orchestra.RunOrchestra(ctx, cfg)` — 현재는 항상 RunOrchestra 호출 (RunOrchestra 내에서 pane 위임)

### 기존 Types
**파일**: `pkg/orchestra/types.go`
- `OrchestraConfig`: Providers, Strategy, Prompt, TimeoutSeconds, JudgeProvider, Terminal
- `OrchestraResult`: Strategy, Responses, Merged, Duration, Summary, FailedProviders
- `ProviderResponse`: Provider, Output, Duration, TimedOut 등

### 보안 유틸
**파일**: `pkg/orchestra/pane_shell.go`
- `sanitizeProviderName()`: 프로바이더 이름에서 위험 문자 제거
- `shellEscapeArg()`: 단일 인자 shell escape
- `uniqueHeredocDelimiter()`: heredoc delimiter 충돌 방지

## 설계 결정

### D1: Auto-Detach vs Explicit Flag
**결정**: pane 터미널 감지 시 자동 detach, `--no-detach`로 opt-out
**근거**: Bash 도구의 120초 timeout은 pane 모드에서 거의 항상 문제를 일으킴. 사용자가 의식적으로 blocking을 선택하는 경우만 예외 처리
**대안 검토**: `--detach` opt-in 방식 → 사용자가 매번 flag를 붙여야 해서 UX 저하. timeout 문제가 기본 동작에서 발생하므로 기본값을 detach로 설정하는 것이 합리적

### D2: Job 디렉토리 위치
**결정**: `/tmp/autopus-orch-{8hex}/`
**근거**: OS가 /tmp를 자동 정리하므로 누수 위험 최소화. 프로젝트 디렉토리를 오염시키지 않음
**대안 검토**: `.autopus/jobs/` — git에 포함되거나 프로젝트 디렉토리 오염. `$XDG_RUNTIME_DIR` — 리눅스 전용, macOS 미지원

### D3: Sentinel 재사용
**결정**: 기존 `__AUTOPUS_DONE__` 마커 그대로 사용
**근거**: pane_runner.go의 `buildPaneCommand()`가 이미 sentinel을 출력 파일 끝에 append. 새 마커를 도입하면 기존 코드와 불일치 발생
**대안 검토**: JSON 완료 마커 — 파싱 복잡도 증가, 기존 `hasSentinel()` 수정 필요

### D4: Status 값
**결정**: running | partial | done | timeout | error (5개 상태)
**근거**: 
- running: 아직 sentinel이 없는 프로바이더 존재, timeout 미도달
- partial: 일부 프로바이더만 sentinel 완료
- done: 전체 프로바이더 sentinel 완료
- timeout: timeout_at 초과
- error: job.json 파싱 실패 등 시스템 오류

### D5: 파일 크기 제한 준수
**결정**: 각 새 파일 120줄 이하, 기존 파일 수정은 30줄 이내
**근거**: 프로젝트 규칙 300줄 하드 제한. 200줄 미만 타겟
**대안 검토**: job.go에 모든 로직 집중 → 300줄 초과 위험. 파일 분리가 안전

### D6: collectPaneResults 재사용 불가
**결정**: detach 모드에서 collectPaneResults()는 호출하지 않음
**근거**: collectPaneResults()는 내부적으로 waitForSentinel()을 goroutine으로 병렬 호출하며 wg.Wait()으로 blocking. 이것이 정확히 timeout 문제의 원인. detach 모드에서는 pane 생성 + 명령 전송까지만 수행하고, sentinel polling은 별도 wait/status 명령에서 수행
