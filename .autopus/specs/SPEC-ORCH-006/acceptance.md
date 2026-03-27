# SPEC-ORCH-006 수락 기준

## 시나리오

### S1: cmux 환경에서 인터랙티브 pane 실행

- Given: cmux 터미널이 감지되고, `Interactive: true`로 OrchestraConfig가 설정됨
- When: `RunPaneOrchestra()`를 3개 프로바이더(claude, codex, gemini)로 실행
- Then: 3개 pane이 생성되고, 각 pane에서 프로바이더 CLI가 인터랙티브 세션으로 시작되며, 프롬프트가 전송되고, 완료 후 결과가 수집되어 merge된 `OrchestraResult`가 반환됨

### S2: tmux 환경에서 인터랙티브 pane 실행

- Given: tmux 터미널이 감지되고, `Interactive: true`로 설정됨
- When: `RunPaneOrchestra()`를 2개 프로바이더로 실행
- Then: tmux `pipe-pane` 및 `capture-pane`을 사용하여 S1과 동일한 결과가 반환됨

### S3: plain 터미널 fallback

- Given: plain 터미널이 감지됨 (cmux/tmux 없음)
- When: `RunPaneOrchestra()`를 `Interactive: true`로 실행
- Then: 기존 `RunOrchestra()` 비대화식 모드로 fallback되어 정상 결과 반환

### S4: ReadScreen 메서드 동작

- Given: cmux 어댑터의 pane에서 텍스트가 출력 중
- When: `ReadScreen(ctx, paneID, ReadScreenOpts{Scrollback: true, Lines: 50})`을 호출
- Then: 최대 50줄의 화면 내용이 문자열로 반환됨

### S5: PipePane 스트리밍 동작

- Given: cmux 어댑터의 pane이 존재
- When: `PipePaneStart(ctx, paneID, "/tmp/output.txt")`를 호출하고, pane에서 출력이 발생
- Then: 출력이 `/tmp/output.txt`에 연속 기록됨
- When: `PipePaneStop(ctx, paneID)`를 호출
- Then: 스트리밍이 중지됨

### S6: 완료 감지 — 프롬프트 패턴

- Given: claude CLI 세션이 실행 중이고, 응답을 완료하여 입력 프롬프트(`>`)가 다시 표시됨
- When: `waitForCompletion()`이 `ReadScreen` 폴링 중
- Then: 프롬프트 패턴 매칭으로 완료가 감지되고, 결과 수집이 시작됨

### S7: 완료 감지 — idle timeout

- Given: codex CLI 세션이 실행 중이고, 프롬프트 패턴을 알 수 없음
- When: pipe-pane 출력 파일에 10초간 새 출력이 없음
- Then: idle 감지로 완료가 판정되고, 결과 수집이 시작됨

### S8: 타임아웃 처리

- Given: gemini CLI 세션이 실행 중이고, 타임아웃이 30초로 설정됨
- When: 30초 경과 후에도 세션이 완료되지 않음
- Then: 해당 pane이 종료되고, 30초 시점의 부분 결과가 `TimedOut: true`로 수집됨

### S9: ANSI 이스케이프 제거

- Given: ReadScreen 결과에 ANSI 색상 코드(`\033[32m`, `\033[0m` 등)가 포함됨
- When: 결과 필터링 단계 실행
- Then: 모든 ANSI 이스케이프 시퀀스가 제거된 깨끗한 텍스트만 merge 로직에 전달됨

### S10: 기존 sentinel 모드 무영향

- Given: `Interactive: false`(기본값)로 OrchestraConfig가 설정됨
- When: `RunPaneOrchestra()`를 실행
- Then: 기존 sentinel 기반 `buildPaneCommand()` 경로로 실행되며 동작 변경 없음

### S11: pane_args 설정 적용

- Given: claude 프로바이더에 `pane_args: []`가 설정됨 (빈 배열)
- When: 인터랙티브 모드에서 claude pane을 시작
- Then: `claude` 바이너리만 실행됨 (인자 없이 인터랙티브 모드 진입)
