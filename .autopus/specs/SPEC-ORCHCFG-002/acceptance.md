# SPEC-ORCHCFG-002 수락 기준

## 시나리오

### S1: 신규 설치 시 기본 프로바이더가 codex
- Given: autopus.yaml이 없는 신규 환경
- When: `auto init`으로 기본 설정을 생성
- Then: Orchestra providers에 codex가 포함되고, opencode는 포함되지 않음
- Then: codex의 args가 `[exec, --approval-mode, full-auto, --quiet, -m, gpt-5.4]`임

### S2: 기존 opencode config 자동 마이그레이션
- Given: autopus.yaml에 opencode 프로바이더가 설정된 기존 환경
- When: config를 로드하여 MigrateOrchestraConfig 실행
- Then: opencode 엔트리가 삭제되고 codex 엔트리가 추가됨
- Then: 모든 command의 providers 리스트에서 opencode가 codex로 교체됨

### S3: codex config가 이미 있는 경우 중복 방지
- Given: autopus.yaml에 codex 프로바이더가 이미 설정된 환경
- When: config를 로드하여 MigrateOrchestraConfig 실행
- Then: 기존 codex 설정이 유지되고 중복 엔트리가 생기지 않음
- Then: changed가 false를 반환

### S4: PlatformToProvider 매핑
- Given: platform이 "opencode"인 환경
- When: PlatformToProvider("opencode") 호출
- Then: "codex"를 반환

### S5: MigrateOpencodeToTUI 호출 제거
- Given: MigrateOrchestraConfig 함수
- When: 마이그레이션 실행
- Then: MigrateOpencodeToTUI가 호출되지 않음 (함수 자체가 제거됨)

### S6: Fallback provider config에 codex 반영
- Given: config 없이 CLI에서 직접 orchestra 실행
- When: buildProviderConfigs(["codex"]) 호출
- Then: binary=codex, args에 exec와 --approval-mode이 포함된 ProviderConfig 반환

### S7: opencode 완료 패턴 하위 호환
- Given: DefaultCompletionPatterns() 호출
- When: 패턴 목록 확인
- Then: opencode의 `Ask anything` 패턴과 codex의 `codex>` 패턴이 모두 존재

### S8: DefaultFullConfig의 codex 반영
- Given: DefaultFullConfig() 호출
- When: Orchestra 섹션 확인
- Then: Providers에 "codex" 키가 존재하고 "opencode" 키는 존재하지 않음
- Then: Commands의 모든 provider 리스트에 "codex"가 포함되고 "opencode"는 미포함

### S9: codex PromptViaArgs 설정
- Given: 새로 생성되는 codex provider entry
- When: PromptViaArgs 필드 확인
- Then: false (stdin 기반 프롬프트 전달)

### S10: 기존 테스트 통과
- Given: 모든 테스트 파일 업데이트 완료
- When: `go test ./pkg/config/... ./internal/cli/... ./pkg/orchestra/...` 실행
- Then: 모든 테스트 통과
