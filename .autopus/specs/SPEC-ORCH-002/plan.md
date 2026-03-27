# SPEC-ORCH-002 구현 계획

## 태스크 목록

- [x] T1: `pkg/pipeline/events.go` — 이벤트 타입 정의 (Event 구조체, EventType 상수, JSONL 직렬화)
- [x] T2: `pkg/pipeline/logger.go` — PipelineLogger 구현 (JSONL + 텍스트 듀얼 기록, ANSI 색상, 프롬프트 주입)
- [x] T3: `pkg/pipeline/monitor.go` — MonitorSession 구현 (상태 모델 → pane late-bind, 2-pane/5-pane 레이아웃)
- [x] T4: `pkg/pipeline/dashboard.go` — "Control Tower" 대시보드 렌더링 (역할 상태, ETA, 블로커)
- [x] T5: `internal/cli/pipeline_dashboard.go` — `auto pipeline dashboard` CLI 명령 등록
- [x] T6: `.claude/skills/autopus/agent-pipeline.md` — 로그 경로 주입 및 대시보드 갱신 패턴 추가
- [x] T7: 단위 테스트 — events, logger, monitor, dashboard 각각에 대한 테스트 파일 작성
- [x] T8: 통합 검증 — cmux 환경에서 end-to-end 모니터링 동작 확인

## 구현 전략

### 기존 코드 활용

- **Terminal 인터페이스**: `pkg/terminal/terminal.go`의 `Terminal` 인터페이스 (`SplitPane`, `SendCommand`, `Close`) 를 그대로 사용. SPEC-ORCH-001에서 검증된 패턴.
- **pane 정리**: `pkg/orchestra/pane_runner.go`의 `cleanupPanes` 패턴을 `MonitorSession.Close()`에서 재활용.
- **체크포인트**: `internal/cli/pipeline.go`의 `LoadCheckpointIfContinue()` 및 `pipeline.Checkpoint` 구조체를 대시보드에서 읽기 전용으로 사용.
- **DetectTerminal**: `pkg/terminal/detect.go`의 `DetectTerminal()`로 cmux 여부 판별.

### 접근 방법

1. **로그 기록은 Go 코드 최소화 + 프롬프트 주입 조합**: Go 측에서는 Phase 전환, Gate 결과 등 핵심 이벤트만 기록. 에이전트 자체 활동 로그는 프롬프트 주입으로 에이전트가 직접 기록하게 유도.
2. **대시보드는 CLI 명령으로 구현**: `auto pipeline dashboard` 명령이 체크포인트 YAML을 읽어 상태를 렌더링. 메인 세션이 Phase 전환 시 `SendCommand`로 대시보드 pane에 갱신 명령을 재전송.
3. **`--team` 모드 호환**: 로그 파일 경로를 공유 자원으로 취급. 서브에이전트/팀원 모두 같은 파일에 append. 파일 잠금 불필요 (append-only).

### 변경 범위

- 신규 파일 5개 (Go 4 + CLI 1)
- 수정 파일 1개 (스킬 문서)
- 테스트 파일 4개
- 기존 Go 소스 코드 수정 없음 (인터페이스 변경 없음)

### 파일 크기 제한 준수

각 파일을 단일 책임으로 분리하여 200줄 이내 목표:
- `events.go`: 이벤트 타입 + JSONL 직렬화 (~60줄)
- `logger.go`: 듀얼 로그 기록 + ANSI 색상 + 프롬프트 주입 (~120줄)
- `monitor.go`: 상태 모델 + pane late-bind + 레이아웃 관리 (~150줄)
- `dashboard.go`: Control Tower 렌더링 (~150줄)
- `pipeline_dashboard.go`: CLI 바인딩 (~50줄)

### 의존성 그래프

```
T1 (events) ──┐
              ├── T2 (logger) ──┐
              │                 ├── T3 (monitor) ── T5 (CLI) ── T8 (통합)
              └── T4 (dashboard)┘                      │
                                                       T6 (skill doc)
                                                       T7 (unit tests, T1+T2+T3+T4 후)
```
