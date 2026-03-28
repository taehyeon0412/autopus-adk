# SPEC-PERM-001 리서치

## 기존 코드 분석

### pkg/detect/detect.go

권한 감지 함수의 자연스러운 배치 위치. 기존 패턴:

- `DetectPlatforms()` — PATH에서 CLI 바이너리 감지
- `DetectOrchestraProviders()` — orchestra 프로바이더 감지
- `IsInstalled(binary string) bool` — 바이너리 설치 확인
- `CheckParentRuleConflicts(projectDir string)` — 부모 디렉토리 규칙 충돌 감지

새 함수 `DetectPermissionMode() string`은 동일한 "Detect" 접두사 네이밍을 따른다. `CheckParentRuleConflicts`처럼 프로세스/파일시스템 환경을 검사하는 패턴이므로 같은 패키지에 적합.

### internal/cli/root.go

Cobra 커맨드 등록 패턴:
- `newXxxCmd()` 함수가 `*cobra.Command`를 반환
- `root.AddCommand(newXxxCmd())`으로 등록
- 현재 약 60줄, `newPermissionCmd()` 추가로 1줄 증가

기존 커맨드 중 유사한 패턴: `platform.go` (`auto platform` — 플랫폼 감지 결과 출력).

### internal/cli/platform.go

`auto platform` 커맨드가 감지 결과를 출력하는 패턴:

```go
func newPlatformCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "platform",
        Short: "Detect installed coding CLI platforms",
        RunE: func(cmd *cobra.Command, args []string) error {
            platforms := detect.DetectPlatforms()
            // ... 출력 로직
        },
    }
    return cmd
}
```

`permission.go`도 동일한 구조를 따른다.

### content/skills/agent-pipeline.md

Pipeline Overview (line 32-46)에서 각 Phase별 에이전트의 mode가 텍스트로 기술됨:
- `planner (model: depends on quality mode, plan)`
- `validator (haiku, plan)`
- `reviewer + security-auditor (parallel)`

Agent Spawning 섹션 (line 95~)에서 실제 Agent() 호출 예시가 있으며, `permissionMode = "bypassPermissions"` 형태로 기술.

변경 포인트: Pipeline Overview 직전에 "Permission Mode Detection" 섹션을 삽입하고, 각 Agent() 예시의 mode를 조건부로 변경.

### templates/claude/commands/auto-router.md.tmpl

하드코딩된 mode 위치 (총 12곳):

| 줄 번호 | 에이전트 | 현재 mode | bypass 시 |
|---------|---------|-----------|-----------|
| 229 | explorer | plan | bypassPermissions |
| 904 | planner | plan | bypassPermissions |
| 937 | tester (scaffold) | bypassPermissions | bypassPermissions (변경 없음) |
| 967 | executor | bypassPermissions | bypassPermissions (변경 없음) |
| 974 | executor (worktree) | bypassPermissions | bypassPermissions (변경 없음) |
| 1043 | validator | plan | bypassPermissions |
| 1086 | annotator | bypassPermissions | bypassPermissions (변경 없음) |
| 1102 | tester | bypassPermissions | bypassPermissions (변경 없음) |
| 1133 | reviewer | plan | bypassPermissions |
| 1143 | security-auditor | plan | bypassPermissions |

실질적 변경 대상: `mode = "plan"` 인 5곳 (229, 904, 1043, 1133, 1143). 이미 `bypassPermissions`인 곳은 변경 불필요.

## 설계 결정

### D1: 프로세스 트리 순회 방식

**결정**: macOS `ps -o args= -p {PID}` + PPID 체인 순회

**근거**:
- Claude Code는 macOS/Linux에서 실행됨
- `ps`는 POSIX 표준이므로 양쪽 OS에서 동작
- `/proc/{pid}/cmdline` (Linux)은 macOS에서 불가
- `ps -o ppid= -p {PID}`로 부모 PID를 얻어 체인 순회

**대안 검토**:
- `/proc` 파일시스템 직접 읽기: Linux 전용, macOS 비호환 → 기각
- `sysctl` (macOS): 크로스플랫폼 불가 → 기각
- 환경변수만 사용: 프로세스에서 자동 감지 불가, 사용자 수동 설정 필요 → P2로 격하

### D2: fail-safe 기본값

**결정**: 감지 실패 시 "safe" 반환

**근거**:
- 보안 원칙 — 의심스러우면 제한적 모드 유지
- 사용자가 명시적으로 `--dangerously-skip-permissions`를 사용한 경우에만 bypass
- 프로세스 검사 실패는 드문 케이스이며, safe 모드는 기능에 문제없음 (프롬프트가 더 나올 뿐)

### D3: 환경변수 오버라이드 우선순위

**결정**: `AUTOPUS_PERMISSION_MODE` > 프로세스 트리 검사

**근거**:
- CI/CD 환경에서 프로세스 트리가 다를 수 있음
- 테스트 시 환경변수로 모드를 강제할 수 있어 편리
- 유효 값만 수용: "bypass", "safe" 외 값은 무시하고 프로세스 검사로 폴백

### D4: 별도 파일 vs detect.go 내 추가

**결정**: `pkg/detect/permission.go`로 별도 파일 생성

**근거**:
- `detect.go`는 이미 191줄, 함수 추가 시 250줄 이상 예상
- 파일 크기 제한 (300줄 하드 리밋) 준수
- 관심사 분리: 플랫폼/의존성 감지 vs 권한 모드 감지

### D5: 서브커맨드 구조

**결정**: `auto permission detect` (2단계 서브커맨드)

**근거**:
- 향후 `auto permission list`, `auto permission set` 등 확장 가능
- 기존 `auto platform` 패턴과 유사하되, permission 도메인으로 분리
- 단일 커맨드 `auto detect-permission`보다 체계적

## 크로스플랫폼 고려사항

### macOS

```bash
# 부모 PID 획득
ps -o ppid= -p $PID

# 프로세스 args 획득
ps -o args= -p $PID
```

### Linux

```bash
# 동일한 ps 명령어 사용 가능
ps -o ppid= -p $PID
ps -o args= -p $PID

# 대안: /proc 직접 읽기 (더 빠르지만 macOS 불가)
cat /proc/$PID/cmdline
```

현재 구현은 `ps` 기반으로 양쪽 호환. Linux 최적화(`/proc`)는 향후 P2로.

## 테스트 전략

### 단위 테스트 (permission_test.go)

- 환경변수 오버라이드 테스트: `t.Setenv` 사용
- 프로세스 트리 순회 모킹: `walkFunc` 함수 타입 주입
- fail-safe 테스트: 에러 반환 mock으로 "safe" 확인

### 통합 테스트

- 실제 `auto permission detect` 바이너리 실행
- JSON 출력 파싱 검증
- 환경변수 설정 후 실행 결과 확인
