# SPEC-ORCH-005 수락 기준

## 시나리오

### S1: Relay Pane 순차 실행 — 정상 흐름

- Given: cmux 또는 tmux 터미널이 감지되고, strategy가 relay이며, 3개 프로바이더(claude, codex, gemini)가 설정됨
- When: orchestra review를 relay 전략으로 실행
- Then:
  - 프로바이더가 순차적으로 각각 별도 pane에서 실행됨
  - 각 프로바이더는 `-p` 플래그 없이 인터랙티브 모드로 실행됨
  - 두 번째 프로바이더 pane의 프롬프트에 "## Previous Analysis by claude" 섹션이 포함됨
  - 세 번째 프로바이더 pane의 프롬프트에 claude와 codex의 이전 분석이 모두 포함됨
  - 최종 결과가 "## Relay Stage N: (by {provider})" 형식으로 출력됨

### S2: 중간 프로바이더 실패 — Skip Continue

- Given: relay pane mode가 활성이고 3개 프로바이더가 설정됨
- When: 두 번째 프로바이더가 실행 중 타임아웃 발생
- Then:
  - 두 번째 프로바이더가 `[SKIPPED: {provider} — context deadline exceeded]`로 기록됨
  - 세 번째 프로바이더가 정상 실행됨 (첫 번째 프로바이더 결과만 맥락으로 주입)
  - 최종 결과에 3개 stage가 모두 포함됨 (두 번째는 SKIPPED)

### S3: Plain 터미널 Fallback

- Given: 터미널이 plain이거나 nil이고, strategy가 relay
- When: orchestra를 실행
- Then:
  - standard relay 실행이 동작함 (SPEC-ORCH-004와 동일한 `-p` 모드)
  - "relay pane mode not yet supported" 경고가 더 이상 출력되지 않음
  - 결과가 정상적으로 반환됨

### S4: Pane Cleanup

- Given: relay pane mode로 3개 프로바이더 실행 완료
- When: 실행이 완료되거나 중간에 에러 발생
- Then:
  - 생성된 모든 pane이 Terminal.Close로 정리됨
  - 임시 output 파일이 삭제됨
  - relay temp 디렉토리가 정리됨 (--keep-relay-output 미설정 시)

### S5: 기존 전략 회귀 없음

- Given: consensus, pipeline, debate, fastest 전략
- When: pane-capable 터미널에서 각 전략을 실행
- Then:
  - 모든 기존 전략이 이전과 동일하게 동작함
  - 병렬 pane 실행이 정상 작동함

### S6: Sentinel 완료 감지

- Given: relay pane에서 프로바이더가 실행 중
- When: 프로바이더가 작업을 완료하고 출력 파일에 `__AUTOPUS_DONE__`이 기록됨
- Then:
  - 500ms 이내에 완료가 감지됨
  - 출력에서 sentinel 마커가 제거되고 결과가 수집됨
  - 다음 프로바이더 pane 실행이 시작됨

### S7: 맥락 주입 내용 검증

- Given: 첫 번째 프로바이더가 "Found 3 security issues" 결과를 생성
- When: 두 번째 프로바이더 pane 명령이 구성됨
- Then:
  - heredoc 내 프롬프트에 원본 사용자 프롬프트가 포함됨
  - `## Previous Analysis by claude` 섹션 하위에 "Found 3 security issues"가 포함됨
  - 쉘 특수문자가 이스케이프 처리됨

### S8: 전체 프로바이더 실패

- Given: relay pane mode에서 모든 프로바이더가 실패
- When: 마지막 프로바이더까지 실패
- Then:
  - 에러가 반환됨 ("relay: all providers failed" 또는 유사)
  - 모든 pane이 정리됨
  - 부분 결과가 FailedProviders에 기록됨

### S9: 파일 크기 제한

- Given: relay_pane.go 구현 완료
- When: 코드 리뷰
- Then:
  - `relay_pane.go`가 300줄 이하 (목표: 200줄 이하)
  - `pane_runner.go`가 기존 280줄에서 300줄을 초과하지 않음
