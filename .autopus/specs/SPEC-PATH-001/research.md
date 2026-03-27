# SPEC-PATH-001 리서치

## 기존 코드 분석

### 1. auto-router.md.tmpl (라우터)

**파일**: `templates/claude/commands/auto-router.md.tmpl`
**현재 동작**: 모든 서브커맨드에서 `.autopus/specs/SPEC-{ID}/` 를 상대 경로로 하드코딩.

- **go 서브커맨드** (라인 641): `Load .autopus/specs/SPEC-{SPEC_ID}/spec.md and check status.`
  - SPEC 로드 경로가 CWD 기준 상대 경로로 고정
  - executor 프롬프트에 WORKING_DIR 개념 없음

- **plan 서브커맨드** (라인 470-596): spec-writer를 스폰하며 `Project directory: {current directory}` 전달
  - --target 플래그 미정의
  - 모듈 결정 로직 없음

- **Context Load 섹션** (라인 15-31): 세션 시작 시 프로젝트 컨텍스트 로드
  - SPEC Path Resolution 섹션을 이 직후에 삽입하는 것이 자연스러움

### 2. spec-writer.md (에이전트)

**파일**: `content/agents/spec-writer.md`
**현재 동작** (라인 27-38): "SPEC 저장 위치 규칙" 섹션

```
1. `git rev-parse --show-toplevel`로 git root를 감지
2. `{git-root}/.autopus/specs/`에 SPEC 디렉토리 생성
```

- git root 감지 → CWD가 autopus-co/이면 항상 autopus-co가 됨
- 서브모듈 구분 불가

**변경 방향**: git root 감지를 제거하고, 호출자(auto-router의 plan 서브커맨드)로부터 전달받은 target module 경로를 사용.

### 3. planner.md (에이전트)

**파일**: `content/agents/planner.md`
**현재 동작** (라인 51): `.autopus/specs/SPEC-XXX/spec.md` 상대 경로 참조
**영향**: go 서브커맨드의 Phase 1에서 planner가 SPEC을 읽을 때 경로가 올바르게 전달되어야 함

### 4. prd.md (스킬)

**파일**: `content/skills/prd.md`
**현재 동작** (라인 36-43): `ls .autopus/specs/`, `cat .autopus/specs/SPEC-*/prd.md` 하드코딩
**영향**: 기존 SPEC 탐색 시 서브모듈의 SPEC을 발견하지 못함. Resolution 기반 글로빙으로 교체 필요.

### 5. idea.md (스킬)

**파일**: `content/skills/idea.md`
**현재 동작** (라인 29-36): "저장 위치 규칙" 섹션

```
1. `git rev-parse --show-toplevel`로 git root를 감지
2. `{git-root}/.autopus/brainstorms/`에 BS 파일 생성
```

- spec-writer.md와 동일한 문제 패턴
- plan의 선행 산출물이므로 동일 모듈 기준 적용 필요

## 설계 결정

### 결정 1: Resolution을 auto-router 내 공통 섹션으로 정의

**이유**: 모든 서브커맨드(go, plan, sync, status, review)가 동일한 resolution 로직을 사용해야 하므로, auto-router.md.tmpl에 한 번 정의하고 각 서브커맨드가 참조하는 구조가 중복을 방지.

**대안 검토**:
- Go 함수로 구현: 하네스 콘텐츠(Markdown 프롬프트)가 아니라 바이너리 변경이 필요. 이번 SPEC의 범위를 넘어감.
- 별도 스킬 파일로 분리: 스킬은 트리거 기반이라 공통 절차 정의에 부적합.

### 결정 2: --target 플래그 도입 (plan, idea)

**이유**: 자동 감지가 실패하거나 모호한 경우 사용자가 명시적으로 모듈을 지정할 수 있어야 함. 기존 글로벌 플래그 패턴(--auto, --multi 등)과 일관됨.

**대안 검토**:
- 항상 자동 감지만 사용: 크로스 모듈 기능이나 새 서브모듈 추가 시 감지 불가.
- CWD 기반: 사용자가 항상 autopus-co/ 루트에서 실행하므로 의미 없음.

### 결정 3: 레거시 호환을 위해 최상단을 resolution 순서 1번으로

**이유**: 기존 5개 SPEC(SPEC-AI-001, SPEC-ORCH-001 등)이 `.autopus/specs/`에 있으므로, 이들을 마이그레이션 없이 그대로 사용 가능해야 함. 새 SPEC부터 서브모듈별 저장을 적용.

### 결정 4: worktree 이슈는 별도 SPEC으로 분리

**이유**: 서브모듈이 독립 git repo인 경우, worktree isolation이 서브모듈 단위로 동작해야 하는 문제는 별도의 설계 검토가 필요. 이번 SPEC의 범위를 "경로 resolution"으로 한정하여 복잡도를 관리.

## 모노레포 구조 현황

```
autopus-co/                          # 최상단 (git root)
├── .autopus/specs/                  # 레거시 SPEC 5개
│   ├── SPEC-AI-001/
│   ├── SPEC-ORCH-001/
│   ├── SPEC-ORCH-002/
│   ├── SPEC-SETUP-002/
│   └── SPEC-TERM-001/
├── autopus-adk/                     # 서브모듈
│   └── .autopus/specs/              # 서브모듈 SPEC 3개
│       ├── SPEC-GAP-001/
│       ├── SPEC-INITUX-001/
│       └── SPEC-ORCH-003/
├── Autopus/                         # 서브모듈 (프론트엔드)
│   └── .autopus/specs/              # (확인 필요)
└── autopus-bridge/                  # 서브모듈
```
