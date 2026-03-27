# SPEC-ORCH-006 구현 계획

## 태스크 목록

- [ ] T1: Terminal 인터페이스 확장 — `ReadScreen`, `PipePaneStart`, `PipePaneStop` 메서드 및 `ReadScreenOpts` 구조체 추가
- [ ] T2: CmuxAdapter 구현 — `cmux read-screen`, `cmux pipe-pane` 명령 래핑
- [ ] T3: TmuxAdapter 구현 — `tmux capture-pane`, `tmux pipe-pane` 명령 래핑
- [ ] T4: PlainAdapter no-op 구현 — 새 메서드 빈 구현
- [ ] T5: `interactive_detect.go` 작성 — 프로바이더별 프롬프트 패턴, idle 감지, ANSI 스트립 유틸리티
- [ ] T6: `interactive.go` 작성 — 인터랙티브 pane 실행 플로우 전체 (pipe capture, session launch, prompt send, completion wait, result collect)
- [ ] T7: `pane_runner.go` 분기 추가 — `Interactive` 플래그 확인하여 인터랙티브 모드 위임
- [ ] T8: `types.go` 수정 — `OrchestraConfig.Interactive` 필드, 완료 패턴 설정 구조체 추가
- [ ] T9: 단위 테스트 — Terminal 어댑터 새 메서드, 완료 감지 로직, ANSI 스트립, 인터랙티브 플로우 mock 테스트
- [ ] T10: autopus.yaml 문서화 — `pane_args` 필드 설명 및 예시 추가

## 구현 전략

### 접근 방법
1. **Bottom-up**: Terminal 인터페이스 확장(T1-T4) → 감지 유틸(T5) → 핵심 로직(T6) → 통합(T7-T8) → 테스트(T9) 순서
2. **기존 코드 최대 활용**: `splitProviderPanes()`, `mergeByStrategy()`, `cleanupPanes()` 등 기존 함수를 그대로 재사용
3. **파일 분리**: 인터랙티브 로직을 `interactive.go`와 `interactive_detect.go`로 분리하여 300줄 제한 준수
4. **Fallback 보장**: 기존 sentinel 모드 코드 경로를 일절 수정하지 않음. `Interactive` 플래그가 true일 때만 새 경로 진입

### 기존 코드 활용
- `pane_runner.go`: `splitProviderPanes()`, `cleanupPanes()`, `mergeByStrategy()`, `paneInfo` 구조체 그대로 재사용
- `pane_shell.go`: `shellEscapeArg()`, `sanitizeProviderName()` 재사용
- `types.go`: `ProviderConfig.PaneArgs` 필드 이미 존재, 추가 필드만 확장

### 변경 범위
- Terminal 인터페이스 변경은 3개 어댑터(cmux, tmux, plain) 모두에 반영 필요
- `pane_runner.go`는 진입점에 분기 1줄 추가 수준
- `types.go`는 필드 2-3개 추가
- 핵심 신규 코드는 `interactive.go`(~180줄)와 `interactive_detect.go`(~120줄)

### 위험 요소
- CLI별 프롬프트 패턴이 버전에 따라 변경될 수 있음 → 설정 가능한 패턴으로 대응
- `read-screen` 출력에 ANSI 잔여물이 있을 수 있음 → ANSI 스트립 유틸리티로 대응
- cmux/tmux `pipe-pane` 동작 차이 → 각 어댑터에서 독립 구현
