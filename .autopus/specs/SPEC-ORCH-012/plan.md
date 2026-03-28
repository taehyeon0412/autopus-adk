# SPEC-ORCH-012 구현 계획

## 태스크 목록

- [ ] T1: cmux `SendLongText` buffer 경로 구현 — `pkg/terminal/cmux.go`
  - 500B 미만: 기존 `SendCommand` 위임 유지
  - 500B 이상: `set-buffer --name <unique> <text>` → `paste-buffer --name <unique> --surface <ref>` → `delete-buffer --name <unique>` (best-effort)
  - unique 이름: `autopus-<paneID sanitized>-<unix-nano>` 형식
  - `set-buffer` 실패 시 warning 로그 + `SendCommand` fallback

- [ ] T2: `launchInteractiveSessions` 전달 경로 변경 — `pkg/orchestra/interactive.go`
  - `buildInteractiveLaunchCmd`의 반환값(개행 미포함)을 `SendLongText`로 전달
  - 호출부의 기존 `+ "\n"` suffix 제거
  - Enter는 별도 `SendCommand("\n")`로 분리

- [ ] T3: cmux `SendLongText` 단위 테스트 — `pkg/terminal/cmux_test.go`
  - short text (<500B): SendCommand 경로 확인
  - long text (>=500B): set-buffer → paste-buffer → delete-buffer 호출 순서 확인
  - unique buffer name 형식 검증
  - set-buffer 실패 시 fallback 동작 확인

- [ ] T4: interactive launch 테스트 업데이트 — `pkg/orchestra/interactive_test.go`
  - `launchInteractiveSessions` 호출 시 `SendLongText` 사용 확인
  - mock의 `sendLongTextCalls` 카운터 추가/검증

## 구현 전략

### 접근 방법

tmux 어댑터의 `SendLongText` 패턴(load-buffer/paste-buffer)을 cmux 어댑터에 적용한다. cmux는 파일 기반이 아닌 CLI argument 기반 `set-buffer`를 사용하므로, Go `exec.Command`가 shell bypass하여 특수문자를 안전하게 전달한다.

### 기존 코드 활용

- `tmux.go:123-163`: SendLongText의 500B 임계값, temp file 패턴 참조
- `cmux.go:59-67`: SendCommand의 `cmux send --surface` 패턴 재사용
- `interactive.go:134-153`: launchInteractiveSessions 수정 대상
- `cmux_test.go:45-62`: `newCmuxMockV2` 테스트 인프라 재사용

### 변경 범위

- **cmux.go**: SendLongText 메서드 ~30줄 추가 (현재 176줄 → ~200줄, 300줄 제한 내)
- **interactive.go**: launchInteractiveSessions 내 2줄 변경 (SendCommand → SendLongText + 별도 Enter)
- **cmux_test.go**: ~40줄 테스트 추가 (현재 201줄 → ~240줄)
- **interactive_test.go**: ~15줄 테스트 추가 (현재 300줄 — 주의: 한계 근접, 필요 시 분리)
