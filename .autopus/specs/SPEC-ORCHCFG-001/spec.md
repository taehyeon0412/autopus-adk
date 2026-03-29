# SPEC-ORCHCFG-001: 커맨드별 동적 합의 임계값

**Status**: completed
**Created**: 2026-03-29
**Domain**: ORCHCFG

## 목적

현재 Orchestra 엔진의 합의 임계값(consensus threshold)은 모든 커맨드에 동일한 값(하드코딩 0.66)이 적용된다. 그러나 보안 리뷰(높은 합의 필요)와 브레인스토밍(낮은 합의 허용)은 근본적으로 다른 수준의 합의를 요구한다. 이 SPEC은 autopus.yaml의 commands 섹션에서 커맨드별로 합의 임계값을 차등 설정할 수 있도록 하고, CLI `--threshold` 플래그로 런타임 오버라이드를 지원한다.

## 요구사항

### R1: 커맨드별 임계값 설정 (ubiquitous)

THE SYSTEM SHALL `config.CommandEntry` 구조체에 `ConsensusThreshold float64` 필드를 추가하여 autopus.yaml의 commands 섹션에서 커맨드별 합의 임계값을 설정할 수 있도록 한다.

### R2: 임계값 해석 우선순위 (ubiquitous)

THE SYSTEM SHALL 합의 임계값을 다음 우선순위로 해석한다:
1. CLI `--threshold` 플래그 값 (최우선)
2. autopus.yaml commands 섹션의 커맨드별 `consensus_threshold` 값
3. `OrchestraConfig.ConsensusThreshold` 전역 값
4. 기본값 0.66

각 단계에서 값이 0이면 다음 단계로 폴백한다.

### R3: handleConsensus 동적 임계값 전달 (ubiquitous)

THE SYSTEM SHALL `handleConsensus()` 함수에서 하드코딩된 0.66 대신 `OrchestraConfig.ConsensusThreshold` 값을 `MergeConsensus()`에 전달한다. 값이 0이면 기본값 0.66을 사용한다.

### R4: CLI --threshold 플래그 (event-driven)

WHEN 사용자가 `--threshold` 플래그와 함께 orchestra 커맨드를 실행하면, THE SYSTEM SHALL 해당 값을 `OrchestraConfig.ConsensusThreshold`에 설정하여 모든 설정 파일 값을 오버라이드한다.

### R5: 임계값 범위 검증 (unwanted)

THE SYSTEM SHALL NOT 0.0 미만 또는 1.0 초과의 임계값을 허용하지 않으며, 범위를 벗어나는 값이 지정되면 에러를 반환한다.

### R6: resolveThreshold 함수 (ubiquitous)

THE SYSTEM SHALL `resolveThreshold(conf *config.OrchestraConf, commandName string, flagValue float64) float64` 함수를 제공하여 R2의 우선순위를 구현한다. 이 함수는 기존 `resolveStrategy`, `resolveProviders`, `resolveJudge` 패턴과 일관된다.

## 생성/수정 파일 상세

| 파일 | 변경 내용 | 영향 범위 |
|------|-----------|-----------|
| `pkg/config/schema.go` | `CommandEntry`에 `ConsensusThreshold float64` 필드 추가 | 설정 파싱 |
| `pkg/orchestra/strategy.go` | `handleConsensus()`에서 `cfg.ConsensusThreshold` 사용 | 합의 전략 실행 |
| `internal/cli/orchestra_config.go` | `resolveThreshold()` 함수 추가 | CLI 설정 해석 |
| `internal/cli/orchestra.go` (또는 command.go) | `--threshold` 플래그 등록, `resolveThreshold` 호출 | CLI 인터페이스 |
| `internal/cli/orchestra_config_test.go` | `resolveThreshold` 테스트 추가 | 테스트 |
| `pkg/orchestra/strategy_test.go` | `handleConsensus` cfg 전달 테스트 | 테스트 |
