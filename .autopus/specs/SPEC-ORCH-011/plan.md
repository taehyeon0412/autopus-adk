# SPEC-ORCH-011 구현 계획

## 태스크 목록

- [ ] T1: Terminal 인터페이스에 `SendLongText` 메서드 추가
  - `pkg/terminal/terminal.go`에 `SendLongText(ctx, paneID, text string) error` 추가
  - 짧은 텍스트(<500B)는 기존 `SendCommand`로 위임, 긴 텍스트는 어댑터별 구현

- [ ] T2: TmuxAdapter에 `SendLongText` 구현 (load-buffer/paste-buffer)
  - `pkg/terminal/tmux.go`에 구현
  - 임시 파일에 텍스트 기록 → `tmux load-buffer <tmpfile>` → `tmux paste-buffer -t <target>` → 임시 파일 삭제
  - 기존 `SendCommand`는 변경 없음 (하위 호환)

- [ ] T3: CmuxAdapter에 `SendLongText` 구현 또는 기존 `SendCommand` 검증
  - `pkg/terminal/cmux.go`에 구현
  - cmux socket API의 긴 텍스트 전달 한계 확인
  - 필요시 임시 파일 경유 또는 청크 분할 방식 적용

- [ ] T4: PlainAdapter에 `SendLongText` stub 구현
  - plain 터미널은 interactive 모드 미지원이므로 no-op 또는 에러 반환

- [ ] T5: `sendPrompts()`에서 `SendLongText` 호출로 전환
  - `pkg/orchestra/interactive.go:188-215` 수정
  - `cfg.Terminal.SendCommand` → `cfg.Terminal.SendLongText` 전환
  - 프롬프트 전송 후 Enter 별도 전송 로직 유지

- [ ] T6: `executeRound()`에서 `SendLongText` 호출로 전환
  - `pkg/orchestra/interactive_debate.go:187-189` 수정
  - `sendPrompts()`와 동일한 패턴 적용

- [ ] T7: opencode `buildInteractiveLaunchCmd` 수정
  - `pkg/orchestra/interactive_launch.go:21`에서 `run` 플래그 제거 로직 재검토
  - opencode는 interactive TUI 대신 `run` 서브커맨드로 직접 실행하는 방식 고려
  - 또는 `InteractiveInput` 필드에 따라 분기 처리

- [ ] T8: `ProviderConfig`에 `InteractiveInput` 필드 추가
  - `pkg/orchestra/types.go:37` 근처에 `InteractiveInput string` 추가
  - 값: `"sendkeys"` (기본), `"paste-buffer"` (긴 텍스트), `"args"` (opencode용)
  - `pkg/config/schema.go:108` 근처에 `InteractiveInput string` YAML 매핑 추가

- [ ] T9: 테스트 작성
  - TmuxAdapter.SendLongText 단위 테스트 (짧은/긴 텍스트)
  - CmuxAdapter.SendLongText 단위 테스트
  - sendPrompts() 통합 테스트 (mock terminal)
  - opencode launch cmd 테스트

- [ ] T10: 기존 테스트 통과 확인
  - `go test ./pkg/terminal/...` 전체 통과
  - `go test ./pkg/orchestra/...` 전체 통과
  - `go test ./internal/cli/...` 전체 통과

## 구현 전략

### 접근 방법

1. **Bottom-up**: Terminal 레이어(T1-T4) → Orchestra 레이어(T5-T8) → 테스트(T9-T10)
2. 기존 `SendCommand` 인터페이스는 변경하지 않음 — 하위 호환성 보장
3. `SendLongText`는 새 메서드로 추가, 짧은 텍스트 fallback 포함

### 기존 코드 활용

- `TmuxAdapter.SendCommand` 패턴(session + paneID → target) 재사용
- `CmuxAdapter.SendCommand` 패턴(surface flag) 재사용
- `buildInteractiveLaunchCmd`의 플래그 필터링 로직 확장

### 변경 범위

- **수정 파일**: 6-8개 (terminal 3개 + orchestra 3개 + config 2개)
- **신규 파일**: 0개 (기존 파일에 메서드 추가)
- **신규 코드**: ~150-200 라인
- **서브에이전트 위임 권장**: 파일 수 3+개, 코드 200 라인 근접

### 의존성

- T1 → T2, T3, T4 (인터페이스 정의 먼저)
- T1-T4 → T5, T6 (Terminal 구현 후 Orchestra 수정)
- T8은 독립 (타입 추가만)
- T9, T10은 T1-T8 완료 후
