# SPEC-ORCH-004 수락 기준

## 시나리오

### S1: Relay 전략 등록 확인

- Given: orchestra 패키지가 초기화된 상태
- When: `Strategy("relay").IsValid()`를 호출
- Then: `true`를 반환한다

### S2: Agentic 플래그 매핑

- Given: claude, codex, gemini 프로바이더가 설정된 relay 실행 환경
- When: 각 프로바이더의 실행 커맨드가 구성될 때
- Then: claude에는 `--allowedTools` 관련 플래그가 포함되고, codex에는 `--approval-mode full-auto`가 포함되며, gemini는 기본 Args를 유지한다

### S3: 순차 실행 및 결과 파일 저장

- Given: 3개 프로바이더(A, B, C)가 relay 전략으로 설정된 상태
- When: `runRelay`가 실행될 때
- Then: A → B → C 순서로 실행되며, 각 결과가 `/tmp/autopus-relay-{jobID}/A.md`, `B.md`, `C.md`에 저장된다

### S4: 프롬프트 릴레이 주입

- Given: 프로바이더 A가 "코드에 XSS 취약점 발견"이라는 결과를 출력한 상태
- When: 프로바이더 B가 실행될 때
- Then: B의 프롬프트에 `## Previous Analysis by A` 섹션이 포함되고, A의 전체 출력이 주입된다

### S5: CLI 활성화

- Given: `auto orchestra review --strategy relay` 명령이 입력된 상태
- When: 커맨드가 파싱될 때
- Then: relay 전략이 선택되고 오류 없이 실행된다

### S6: 하위 호환성

- Given: 기존 전략(consensus, pipeline, debate, fastest)이 설정된 상태
- When: 기존 전략으로 orchestra를 실행할 때
- Then: relay 추가 이전과 동일한 동작을 한다 (기존 테스트 전체 통과)

### S7: Temp 디렉토리 정리

- Given: relay 실행이 완료된 상태에서 `--keep-relay-output`이 미설정
- When: 결과가 반환된 후
- Then: `/tmp/autopus-relay-{jobID}/` 디렉토리가 삭제된다

### S8: Temp 디렉토리 보존

- Given: relay 실행이 완료된 상태에서 `--keep-relay-output`이 설정
- When: 결과가 반환된 후
- Then: `/tmp/autopus-relay-{jobID}/` 디렉토리와 결과 파일이 보존된다

### S9: 부분 실패 처리

- Given: 3개 프로바이더 중 2번째가 실행 실패한 상태
- When: relay가 오류를 감지할 때
- Then: 1번째 프로바이더의 결과는 보존하고, 3번째는 건너뛰며, 부분 결과와 실패 정보를 반환한다

### S10: 결과 포맷팅

- Given: 3개 프로바이더가 모두 성공적으로 완료된 상태
- When: 결과가 병합될 때
- Then: `## Relay Stage 1: (by A)`, `## Relay Stage 2: (by B)`, `## Relay Stage 3: (by C)` 형식으로 포맷되고, 요약에는 "릴레이: 3단계 완료"가 포함된다
