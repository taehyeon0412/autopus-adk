# SPEC-AXQUAL-001 수락 기준

## 시나리오

### S1: Explicit platform 전달 시 그대로 반환
- Given: `platform` 인자가 `"opencode"`로 설정됨
- When: `resolvePlatform("opencode")` 호출
- Then: `"opencode"` 반환, PATH 탐색 없음

### S2: PATH에 claude 바이너리 존재
- Given: PATH에 `claude` 실행 파일만 존재
- When: `resolvePlatform("")` 호출
- Then: `"claude"` 반환

### S3: PATH에 codex만 존재 (claude 없음)
- Given: PATH에 `codex` 실행 파일만 존재 (claude 없음)
- When: `resolvePlatform("")` 호출
- Then: `"codex"` 반환

### S4: PATH에 gemini만 존재
- Given: PATH에 `gemini` 실행 파일만 존재
- When: `resolvePlatform("")` 호출
- Then: `"gemini"` 반환

### S5: PATH에 아무 바이너리도 없음
- Given: PATH가 빈 디렉토리를 가리킴 (인식 가능한 바이너리 없음)
- When: `resolvePlatform("")` 호출
- Then: fallback으로 `"claude"` 반환

### S6: PATH에 여러 바이너리 존재 시 우선순위
- Given: PATH에 `codex`와 `gemini` 모두 존재 (claude 없음)
- When: `resolvePlatform("")` 호출
- Then: 우선순위에 따라 `"codex"` 반환 (codex > gemini)

### S7: 템플릿 TODO에 @AX:EXCLUDE 마커 존재
- Given: `agent_create.go`와 `skill_create.go` 파일
- When: `@AX:TODO` 패턴으로 grep 실행
- Then: 템플릿 내 TODO는 매치되지 않음 (@AX:EXCLUDE로 제외 처리됨)

### S8: @AX:TODO 태그 제거 확인
- Given: SPEC 구현 완료 후 코드베이스
- When: `grep -r "@AX:TODO" internal/cli/pipeline_run.go`
- Then: 매치 결과 없음

### S9: 기존 테스트 회귀 없음
- Given: 새 테스트 추가 후
- When: `go test ./internal/cli/ -count=1` 실행
- Then: 기존 테스트 포함 전체 통과
