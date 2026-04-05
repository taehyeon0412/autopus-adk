# SPEC-AXSEC-001 구현 계획

## 태스크 목록

- [ ] T1: `pkg/experiment/cmdvalidate.go` — shell metacharacter 검증 함수 구현
- [ ] T2: `pkg/experiment/cmdvalidate_test.go` — 검증 함수 TDD 테스트 (먼저 작성)
- [ ] T3: `pkg/experiment/metric.go` — `RunMetric` 진입부에 `ValidateCommand` 호출 삽입
- [ ] T4: `pkg/experiment/metric_test.go` — injection 시도 케이스 추가 (semicolon, pipe, subshell 등)
- [ ] T5: `pkg/pipeline/branchvalidate.go` — branch name regex 검증 함수 구현
- [ ] T6: `pkg/pipeline/branchvalidate_test.go` — branch name 검증 TDD 테스트 (먼저 작성)
- [ ] T7: `pkg/pipeline/worktree.go` — `addWorktreeWithRetry` 진입부에 인라인 검증 추가, `sanitizeBranchName` error 반환
- [ ] T8: `pkg/pipeline/worktree_test.go` — 악의적 branch name 거부 테스트 추가
- [ ] T9: 기존 테스트 전체 통과 확인 (backward compatibility)

## 구현 전략

### 접근 방법: Defense-in-Depth Layering

1. **검증 함수를 별도 파일로 분리** — 파일 사이즈 제한(300줄) 준수 및 단일 책임 원칙
2. **TDD**: T2, T6 테스트를 먼저 작성하여 실패 확인 후 T1, T5 구현
3. **기존 코드 최소 변경** — 검증 호출 1줄 추가 수준으로 유지

### Shell Metacharacter 검증 설계

```
disallowed := []string{";", "|", "&&", "||", "$(", ")`", "`", "{", "}", "<", ">", "\n"}
```

- `AllowShellMeta` 옵션은 `RunMetricOption` functional option 패턴으로 제공
- 기본값은 strict (차단), opt-in으로 해제 가능
- 향후 allowlist 확장 시 config 파일에서 로드하는 방향으로 확장 가능

### Branch Name 검증 설계

```
validBranch := regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)
```

- `addWorktreeWithRetry` 내부에서 직접 검증 (sanitize 호출 여부와 무관하게 안전)
- `sanitizeBranchName` 시그니처를 `(string, error)`로 변경하여 silent replacement 대신 명시적 에러

### 변경 범위

- 새 파일 4개 (각 ~50-80줄)
- 기존 파일 수정 2개 (각 +5-10줄)
- 기존 테스트 파일 수정 2개 (각 +20-30줄)

### 의존성

- T1 ← T2 (TDD 순서)
- T3 ← T1
- T5 ← T6 (TDD 순서)
- T7 ← T5
- T4, T8은 독립 실행 가능
- T9는 모든 태스크 완료 후
