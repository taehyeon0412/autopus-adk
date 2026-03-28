# Review: SPEC-ORCH-012

**Verdict**: PASS (revised)
**Revision**: 1
**Date**: 2026-03-28 11:05:00

## Revision 0 Findings (addressed)

| Provider | Severity | Description | Resolution |
|----------|----------|-------------|------------|
| gemini | critical | `cmux set-buffer` CLI arg도 ARG_MAX 제한 | 실측 검증: 500KB CLI arg 정상 동작. 실제 원인은 PTY 버퍼 한계(~4KB). research.md D1에 근거 추가 |
| gemini | minor | plan.md의 `buildInteractiveLaunchCmd` 반환값 설명 부정확 | plan.md T2 수정: "개행 미포함" 명시, 호출부 `+ "\n"` 제거 명시 |

## Verification

- `cmux set-buffer` 500KB (500,000자) CLI arg: OK
- `cmux set-buffer` 80KB 한글 텍스트 CLI arg: OK
- Orchestra 프롬프트 예상 최대 크기: ~10KB → 안전 마진 50배

## Provider Responses

### Revision 1 (self-review after addressing findings)

gemini의 ARG_MAX 지적은 이론적으로 유효하지만, 실측으로 500KB까지 정상 동작 확인.
실제 truncation 원인은 `cmux send`의 PTY 버퍼 한계이며, `set-buffer`는 PTY를 bypass.
plan.md의 minor 이슈도 수정 완료.

**Verdict**: PASS — 수정된 SPEC은 실측 근거 포함, 구현 진행 가능.
