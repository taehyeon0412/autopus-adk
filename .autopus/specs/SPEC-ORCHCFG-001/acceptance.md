# SPEC-ORCHCFG-001 수락 기준

## 시나리오

### S1: YAML 커맨드별 임계값 설정

- Given: autopus.yaml에 `secure` 커맨드의 `consensus_threshold: 0.85`가 설정됨
- When: `auto orchestra --multi secure` 실행
- Then: `MergeConsensus()`에 threshold=0.85가 전달됨

### S2: CLI --threshold 플래그 오버라이드

- Given: autopus.yaml에 `review` 커맨드의 `consensus_threshold: 0.75`가 설정됨
- When: `auto orchestra --multi review --threshold 0.9` 실행
- Then: `MergeConsensus()`에 threshold=0.9가 전달됨 (YAML 값 0.75가 아닌 CLI 값 우선)

### S3: 기본값 폴백

- Given: autopus.yaml에 `brainstorm` 커맨드의 `consensus_threshold`가 설정되지 않음 (0)
- Given: `OrchestraConfig.ConsensusThreshold`도 0
- When: consensus 전략으로 brainstorm 실행
- Then: 기본값 0.66이 사용됨

### S4: 전역 ConsensusThreshold 폴백

- Given: `OrchestraConfig.ConsensusThreshold`가 0.7로 설정됨
- Given: 해당 커맨드에 커맨드별 `consensus_threshold`가 없음
- When: consensus 전략 실행
- Then: 전역값 0.7이 사용됨

### S5: 범위 검증 — 초과

- Given: 사용자가 `--threshold 1.5` 플래그를 지정
- When: orchestra 커맨드 실행
- Then: 에러 메시지 출력, 실행 중단

### S6: 범위 검증 — 음수

- Given: 사용자가 `--threshold -0.1` 플래그를 지정
- When: orchestra 커맨드 실행
- Then: 에러 메시지 출력, 실행 중단

### S7: resolve 패턴 일관성

- Given: resolveThreshold() 함수 구현
- When: 테스트에서 (flagValue=0, commandEntry.ConsensusThreshold=0.8) 호출
- Then: 0.8 반환 (커맨드 설정 값 사용)

### S8: handleConsensus cfg 전달

- Given: `OrchestraConfig{ConsensusThreshold: 0.8}` 설정
- When: `handleConsensus(ctx, responses, cfg)` 호출
- Then: `MergeConsensus(responses, 0.8)` 호출 (하드코딩 0.66이 아님)

### S9: 하위 호환성

- Given: 기존 autopus.yaml에 `consensus_threshold` 필드가 없음
- When: 시스템 실행
- Then: 모든 기존 동작이 동일하게 유지 (기본값 0.66)
