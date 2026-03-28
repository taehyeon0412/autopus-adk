# SPEC-PERM-001 수락 기준

## 시나리오

### S1: bypass 모드에서 커맨드 실행

- Given: 부모 프로세스 트리에 `claude --dangerously-skip-permissions`가 존재하는 환경
- When: `auto permission detect` 커맨드를 실행하면
- Then: stdout에 "bypass"가 출력되고, exit code 0으로 종료

### S2: safe 모드에서 커맨드 실행

- Given: 부모 프로세스 트리에 `--dangerously-skip-permissions` 플래그가 없는 환경
- When: `auto permission detect` 커맨드를 실행하면
- Then: stdout에 "safe"가 출력되고, exit code 0으로 종료

### S3: 프로세스 검사 실패 시 fail-safe

- Given: 프로세스 트리 검사가 실패하는 환경 (권한 부족 등)
- When: `DetectPermissionMode()` 함수가 호출되면
- Then: 에러를 무시하고 "safe"를 반환

### S4: 환경변수 오버라이드

- Given: `AUTOPUS_PERMISSION_MODE=bypass` 환경변수가 설정된 상태
- When: `auto permission detect` 커맨드를 실행하면
- Then: 프로세스 트리 검사를 건너뛰고 "bypass"를 출력

### S5: JSON 출력 모드

- Given: bypass 모드 환경
- When: `auto permission detect --json` 커맨드를 실행하면
- Then: `{"mode":"bypass","parent_pid":<N>,"flag_found":true}` 형태의 JSON이 출력

### S6: agent-pipeline bypass 모드 동작

- Given: PERMISSION_MODE가 "bypass"로 감지된 파이프라인 실행
- When: Phase 1 planner 에이전트가 스폰되면
- Then: mode가 "bypassPermissions"로 설정되어 도구 승인 프롬프트 없이 실행

### S7: agent-pipeline safe 모드 동작 (하위 호환)

- Given: PERMISSION_MODE가 "safe"로 감지된 파이프라인 실행
- When: Phase 1 planner 에이전트가 스폰되면
- Then: mode가 "plan"으로 설정되어 기존 동작과 동일

### S8: auto-router Step 0.5 실행

- Given: `/auto go SPEC-ID` 커맨드로 파이프라인이 시작
- When: Route A (서브에이전트 파이프라인)가 선택되면
- Then: Phase 1 시작 전에 `auto permission detect`가 실행되고 PERMISSION_MODE 변수가 설정

### S9: 환경변수 잘못된 값

- Given: `AUTOPUS_PERMISSION_MODE=invalid` 환경변수가 설정된 상태
- When: `DetectPermissionMode()` 함수가 호출되면
- Then: 유효하지 않은 값은 무시하고 프로세스 트리 검사로 폴백

## 비기능 요구사항

### NF1: 성능

- `auto permission detect` 실행 시간이 500ms 미만이어야 함
- 프로세스 트리 순회 깊이 최대 20단계 (무한 루프 방지)

### NF2: 보안

- 프로세스 트리 검사 결과를 로그에 기록하지 않음 (보안 플래그 노출 방지)
- 환경변수 오버라이드는 로컬 실행 환경에서만 유효

### NF3: 파일 크기

- 모든 신규 Go 소스 파일은 200줄 미만 (300줄 하드 리밋)
