# SPEC-TELE-001: 파이프라인 옵저버빌리티 & 텔레메트리

**Status**: completed
**Created**: 2026-03-22
**Domain**: TELE
**Priority**: Must Have (Tier 1)

## 목적

업계에서 OpenTelemetry가 89% 도입되었으나, AI 코딩 하네스 경쟁사(Superpowers, MoAI, Octopus) 중 아무도 에이전트 옵저버빌리티를 구현하지 않았다. 이것은 선점 기회다.

Autopus-ADK의 멀티에이전트 파이프라인(--team)은 현재 실행 추적/메트릭/로그가 전혀 없어 디버깅과 비용 최적화가 불가능하다. 에이전트 실행 기록, 파이프라인 메트릭, 비용 추정을 제공하는 텔레메트리 시스템을 구현한다.

## 요구사항

### R1: 에이전트 실행 기록
WHEN 에이전트(planner, executor, validator, tester, reviewer)가 실행되면 THE SYSTEM SHALL 에이전트명, 시작/종료 시간, 소요 시간, 상태(PASS/FAIL), 수정 파일 수, 추정 토큰 수를 기록한다.

### R2: 파이프라인 실행 기록
WHEN 파이프라인이 시작되면 THE SYSTEM SHALL SPEC-ID, Phase별 소요 시간, 전체 소요 시간, 재시도 횟수, 최종 상태를 기록한다.

### R3: JSONL 저장
THE SYSTEM SHALL 텔레메트리 데이터를 `.autopus/telemetry/{date}-{spec-id}.jsonl` 형식으로 저장한다.

### R4: 자연어 조회
WHEN 사용자가 "비용", "cost", "텔레메트리", "파이프라인 결과" 등의 키워드를 사용하면 THE SYSTEM SHALL IntentRule 매칭을 통해 텔레메트리 리포트를 자동 표시한다.

### R5: 비용 추정
WHEN 텔레메트리 데이터가 존재하면 THE SYSTEM SHALL 모델별 토큰 단가 테이블을 기반으로 추정 비용을 계산한다.

### R6: 비용 자동 표시
WHEN 파이프라인이 완료되면 THE SYSTEM SHALL 완료 요약에 추정 비용을 자동 포함한다.

### R7: Quality Mode 비용 비교
WHEN `auto go --team` 실행 시 Quality Mode 선택 UI에서 THE SYSTEM SHALL Ultra와 Balanced의 예상 비용 차이를 표시한다.

### R8: 하네스 통합
WHEN `auto go --team` 파이프라인이 실행되면 THE SYSTEM SHALL 각 Phase 시작/종료 시점에 텔레메트리 기록 지시를 자동으로 삽입한다.

### R9: 텔레메트리 설정
WHERE autopus.yaml에 `telemetry` 섹션이 있으면 THE SYSTEM SHALL enabled, retention_days, cost_tracking 설정을 로드한다.

### R10: 텔레메트리 자동 정리
WHEN 파이프라인이 시작되면 THE SYSTEM SHALL retention_days(기본 30일)를 초과한 텔레메트리 파일을 백그라운드에서 자동 삭제한다.

## 의존성 제약

- `pkg/cost` → `pkg/telemetry` (단방향): cost 패키지가 telemetry 타입을 임포트
- `pkg/telemetry` → `pkg/cost` 의존 금지: 아키텍처 규칙 "pkg 간 양방향 의존 금지" 준수
- 비용 계산이 필요한 경우 telemetry는 `CostEstimator` 인터페이스를 정의하고 cost가 구현

## Zero-Command 통합

- **사용자 향 명령어/플래그**: 없음 — 사용자는 자연어로만 접근
- **내부 CLI**: `auto telemetry record` — 에이전트가 파이프라인 실행 중 호출하는 내부 전용 명령 (사용자 직접 호출 불필요)
- **동작 방식**: --team 파이프라인 실행 시 메인 세션이 Agent() 호출 전후에 `auto telemetry record`를 자동 호출. 완료 요약에 비용 자동 표시. 기록은 `.autopus/telemetry/`에 JSONL로 축적.
- **사용자 체감**: 파이프라인 완료 시 "소요: 4m 32s | 추정 비용: $0.45 (Balanced)" 자동 표시
- **자연어 접근**: "비용 얼마 들었어?" → 최근 파이프라인 비용 표시 / "지난번이랑 비교해줘" → 비교 리포트
- **끄기**: `autopus.yaml` → `telemetry.enabled: false`

## 생성/수정 파일

| 파일 | 유형 | 설명 |
|------|------|------|
| `pkg/telemetry/types.go` | 신규 (~80줄) | AgentRun, PipelineRun, Event 타입 |
| `pkg/telemetry/recorder.go` | 신규 (~120줄) | JSONL 기록기 |
| `pkg/telemetry/reader.go` | 신규 (~100줄) | JSONL 읽기 + 필터 |
| `pkg/telemetry/reporter.go` | 신규 (~120줄) | 요약/비교 리포트 |
| `pkg/cost/pricing.go` | 신규 (~60줄) | 모델별 토큰 단가 테이블 |
| `pkg/cost/estimator.go` | 신규 (~100줄) | 비용 추정 엔진 |
| `pkg/cost/report.go` | 신규 (~80줄) | 비용 리포트 생성 |
| `pkg/config/schema.go` | 수정 | TelemetryConf 추가 |
| `pkg/config/defaults.go` | 수정 | TelemetryConf 기본값 (enabled: true, retention_days: 30, cost_tracking: true) |
| `.claude/commands/auto.md` | 수정 | 파이프라인 텔레메트리 통합 |
| `pkg/content/intent.go` | 수정 | 비용/텔레메트리 IntentRule 추가 |
