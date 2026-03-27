# SPEC-ORCH-002: cmux 모니터링 대시보드 — Agent Pipeline 실행 상태 시각화

**Status**: completed
**Created**: 2026-03-25
**Domain**: ORCH

## 목적

Agent Pipeline (`/auto go SPEC-ID`) 실행 시, 서브에이전트와 Agent Teams는 Claude Code 내부 API(Agent tool, SendMessage)로 동작하여 외부에서 실행 과정을 직접 관찰할 수 없다. SPEC-ORCH-001이 orchestra 멀티프로바이더의 pane 실행 시각화를 해결했다면, 이 SPEC은 Agent Pipeline의 **상태 모니터링** 문제를 해결한다.

cmux 감지 시 2개 pane을 생성하여: (1) 통합 로그 스트림, (2) 파이프라인 대시보드를 표시함으로써, 사용자가 에이전트 활동과 파이프라인 진행 상태를 실시간으로 모니터링할 수 있게 한다.

## 요구사항

### R1: cmux 감지 및 모니터링 pane 생성 (P0)

WHEN 파이프라인이 시작되면, THE SYSTEM SHALL `pkg/terminal.DetectTerminal()`을 호출하여 cmux 사용 가능 여부를 판별한다.
- cmux 감지 시 (기본 레이아웃): 2개 pane을 수평 분할로 생성 (로그 pane + 대시보드 pane)
- cmux 감지 시 (`--team` 5-pane 옵션): 1+4 레이아웃 지원 — 메인 대시보드 + lead/builder/tester/guardian 역할별 pane (2x2 quadrant). `--monitor full` 플래그로 활성화
- cmux 미감지 시: 모니터링 pane 없이 기존 동작 유지 (graceful skip)
- 상태 모델을 먼저 생성하고 pane을 late-bind하여 fallback 전환을 용이하게 한다

### R2: JSONL 이벤트 로그 기록 (P0)

WHILE 파이프라인이 실행 중일 때, THE SYSTEM SHALL `/tmp/autopus-pipeline-{spec-id}.jsonl`에 구조화된 이벤트를 기록한다.
- 이벤트 형식: `{"ts":"ISO8601","agent":"name","phase":"N","type":"event_type","msg":"message"}`
- 이벤트 타입: `agent_spawned`, `phase_started`, `phase_completed`, `gate_result`, `message_sent`, `error`, `metric`
- 에이전트 스폰, Phase 전환, Gate 결과, 에러 이벤트를 기록한다
- 로그 파일 경로는 에이전트 프롬프트에 주입하여, 에이전트도 자체 활동을 기록할 수 있게 한다
- 사람이 읽을 수 있는 텍스트 로그(`*.log`)도 병행 기록한다 (tail -f용)

### R3: 통합 로그 pane (P0)

WHEN 모니터링 pane이 생성되면, THE SYSTEM SHALL pane 1에서 `tail -f /tmp/autopus-pipeline-{spec-id}.log`를 실행하여 모든 에이전트 활동의 실시간 로그 스트림을 표시한다.

### R4: 파이프라인 "Control Tower" 대시보드 pane (P0)

WHEN 모니터링 pane이 생성되면, THE SYSTEM SHALL pane 2에서 파이프라인 "Control Tower" 대시보드를 표시한다.
- Phase별 진행 상태 (pending/running/done/failed) — `✓/→/○/✗` 아이콘 사용
- 역할별 상태: 에이전트 이름, 현재 Phase, 마지막 메시지 요약
- 경과 시간 및 예상 완료 시간 (ETA) — 이전 Phase 소요 시간 기반 추정
- 현재 블로커 (있는 경우): Gate 실패 이유, 재시도 횟수
- 대시보드는 이벤트 로그 기반으로 갱신된다 (Phase 전환, Gate 결과 시)

### R5: 에이전트 프롬프트 로그 경로 주입 (P0)

WHEN 에이전트를 스폰할 때, THE SYSTEM SHALL 에이전트 프롬프트에 로그 파일 경로를 포함하여, 에이전트가 자체 활동 로그를 기록할 수 있게 한다.
- 주입 형식: `## Pipeline Monitor\nLog file: /tmp/autopus-pipeline-{spec-id}.log\nWrite structured log entries: [timestamp] [your-role] [phase] message`

### R6: pane 정리 및 로그 삭제 (P0)

WHEN 파이프라인이 완료되면, THE SYSTEM SHALL 생성된 모니터링 pane을 닫고, 임시 로그 파일을 삭제한다.
- SPEC-ORCH-001의 `cleanupPanes` 패턴을 재활용한다

### R7: 서브에이전트 및 Agent Teams 모드 지원 (P0)

WHEN `--team` 플래그 유무에 관계없이, THE SYSTEM SHALL 동일한 모니터링 메커니즘을 제공한다.
- 기본 모드 (서브에이전트): 메인 세션이 로그 기록 담당
- `--team` 모드: 각 팀원이 SendMessage 시 로그 기록

### R8: 대시보드 렌더링 (P1)

THE SYSTEM SHALL `auto pipeline dashboard {spec-id}` CLI 명령을 제공하여 대시보드를 렌더링한다.
- 체크포인트 파일 (`.autopus/pipeline-state/{specID}.yaml`) 기반으로 상태를 읽는다
- 메인 세션이 Phase 전환 시 `SendCommand`로 대시보드 pane에 갱신 명령을 전송한다

### R9: 로그 기록 실패 안전성 (P1)

WHEN 로그 파일 기록에 실패하면, THE SYSTEM SHALL 파이프라인 실행을 중단하지 않고 경고만 출력한다.
- 모니터링은 부가 기능이므로 파이프라인 핵심 흐름에 영향을 주지 않는다

### R10: 역할별 ANSI 색상 코딩 (P1)

WHEN 로그를 텍스트 pane에 표시할 때, THE SYSTEM SHALL 에이전트 역할에 따라 ANSI 색상을 적용한다.
- lead: Cyan (`\033[36m`)
- builder: Green (`\033[32m`)
- tester: Yellow (`\033[33m`)
- guardian/reviewer: Red (`\033[31m`)
- auditor: Magenta (`\033[35m`)
- 대시보드 Phase 상태에도 동일 색상 적용

### R11: 세션 재생 아티팩트 보존 (P1)

WHEN 파이프라인이 완료되면, THE SYSTEM SHALL JSONL 이벤트 로그를 `.autopus/pipeline-state/{spec-id}/events.jsonl`에 보존한다.
- 텍스트 로그(`/tmp/...`)는 삭제하되, JSONL 이벤트 로그는 영구 보존
- `auto pipeline replay {spec-id}` 명령으로 세션 재생 가능 (후속 구현)
- 사후 분석(post-mortem)에 활용: "왜 builder가 멈췄는지", "어디서 시간이 소요되었는지"

## 생성 파일 상세

### `pkg/pipeline/monitor.go` (신규)
모니터링 세션 관리: pane 생성, 로그 파일 초기화, pane 정리. `MonitorSession` 구조체가 로그 pane ID, 대시보드 pane ID, 로그 파일 경로를 관리한다.

### `pkg/pipeline/logger.go` (신규)
구조화된 파이프라인 로그 기록기. `PipelineLogger` 구조체가 로그 형식화 및 파일 기록을 담당한다. 에이전트 프롬프트 주입용 로그 경로 반환 메서드 포함.

### `pkg/pipeline/dashboard.go` (신규)
대시보드 렌더링 로직. 체크포인트 파일 기반으로 Phase 상태, 에이전트 목록, 경과 시간을 포맷팅하여 출력한다.

### `internal/cli/pipeline_dashboard.go` (신규)
`auto pipeline dashboard {spec-id}` CLI 명령 핸들러. `pkg/pipeline/dashboard.go`를 호출하여 결과를 stdout에 출력한다.

### `.claude/skills/autopus/agent-pipeline.md` (수정)
에이전트 스폰 시 로그 경로 주입 패턴 추가. Phase 전환 시 대시보드 갱신 지시 추가.
