# SPEC-GAP-001: Autopus-ADK 경쟁력 강화 갭 분석 및 개선 로드맵

**Status**: completed
**Created**: 2026-03-24
**Domain**: GAP
**Priority**: Strategic (전체 로드맵)
**Revision**: 2 (Revision 1 리뷰 피드백 반영: CGO 회피, 복원 수준 명확화, 게이트 메커니즘 구체화)

## 목적

Autopus-ADK는 5-Phase 파이프라인, RALF loop, @AX annotation, Lore commit 등 고유한 강점을 보유하지만, 4개 주요 경쟁 프로젝트(Superpowers, MoAI-ADK, OhMyOpenAgent, Claude Octopus) 대비 채택 장벽과 기능 격차가 존재한다.

이 SPEC은 경쟁력 분석 결과를 체계적으로 구조화하고, 3개 Tier(Must Have / Should Have / Could Have)로 나눈 개선 로드맵을 정의한다. 각 갭 항목은 독립 SPEC으로 분해되어 구현된다.

## 아키텍처 경계 (Architecture Boundary)

Autopus-ADK는 **하네스(harness)**이다 — rules, agents, skills, hooks를 `.claude/`에 배포하는 CLI 도구이지, LLM 호출을 직접 수행하거나 런타임 프록시 역할을 하는 시스템이 아니다. 모든 요구사항은 이 경계 내에서 정의된다.

## 경쟁 환경 요약

| 프로젝트 | 핵심 차별점 | Stars/규모 |
|---------|-----------|-----------|
| Superpowers | Git clone 배포, mandatory gates, community skills | 89K stars |
| MoAI-ADK | 27 agents (4 tiers), 52 skills, session persistence, 18 languages | Go binary |
| OhMyOpenAgent | Hashline Edit, Boulder resume, LSP+AST-Grep 25+ langs, tmux | 11 agents |
| Claude Octopus | 8 providers, Double Diamond, Reaction Engine, 75% consensus | 32 personas |

## Autopus-ADK 현재 강점 (유지/강화 대상)

- S1: 5-Phase multi-agent pipeline (Phase 2.5 Annotation, Phase 3.5 UX Verify)
- S2: RALF loop with circuit breaker
- S3: @AX annotation system
- S4: Lore commit (9-trailer protocol)
- S5: Adaptive quality (opus/sonnet/haiku routing)
- S6: Worktree isolation with file ownership conflict detection
- S7: Cost estimation and telemetry
- S8: Experiment loop (autonomous iteration)
- S9: SigMap (AST-based API signature tracking)

## 요구사항

### Tier 1: Must Have — Critical

#### ~~R1: 경량 배포 (Lite Distribution)~~ → **삭제됨**
> 설계 결정: Autopus-ADK는 Go binary 풀 배포만 지원. lite/plugin-only 모드는 채택하지 않음. Go binary가 제공하는 orchestra, arch, spec, telemetry 등의 통합 기능이 핵심 가치이며, 파일 배포만으로는 이 가치를 전달할 수 없음.

#### R2: Multi-Language Support (Phase 1: Go + TypeScript)
WHEN 사용자가 TypeScript 프로젝트에서 Autopus를 사용할 때 THE SYSTEM SHALL TypeScript용 SigMap, 테스팅 전략 (jest/vitest 감지), 코드 분석을 제공한다. Phase 1은 Go + TypeScript만 지원하며, Phase 2에서 Python, Rust 순으로 확장한다.

**Acceptance Criteria**:
- AC2.1: TypeScript 프로젝트에서 `auto setup`이 TypeScript 시그니처를 추출한다
- AC2.2: 테스팅 전략이 jest/vitest를 자동 감지하고 TDD 스킬에 반영한다
- AC2.3: `auto arch generate`가 TypeScript import 그래프를 분석한다

**기술 전략 (Architecture Decision)**:
- **CGO 회피**: 순수 Go 빌드를 유지한다. Tree-sitter C 바인딩(CGO)은 도입하지 않는다.
- **Phase 1 접근**: 정규식 기반 TypeScript 시그니처 추출 (`export function`, `export class`, `export interface` 패턴 매칭). `pkg/sigmap/`에 `Extractor` 인터페이스를 추가하고, Go용 `GoExtractor`와 TypeScript용 `TSExtractor`를 분리한다.
- **Phase 2 고려**: 정규식 한계 도달 시 `tree-sitter-cli`를 subprocess로 호출하는 어댑터 검토. CGO 없이 `exec.Command("tree-sitter", "parse", ...)` 방식.
- **빌드 영향**: goreleaser 설정 변경 없음. 순수 Go 빌드 유지.

**Dependency**: SigMap 확장 → SPEC-SIGMAP-001 (completed, Go 전용) 위에 구축

**Phase 완료 기준**:
- Phase 1 완료: Go + TypeScript SigMap 동작, jest/vitest 감지 → 독립 SPEC(SPEC-SIGMAP-002) 구현 완료
- Phase 2 트리거: Phase 1 완료 후 Python/Rust 요청 발생 시

#### R3: Pipeline State Persistence
WHEN 파이프라인 실행 중 세션이 중단되면 THE SYSTEM SHALL 현재 Phase, 완료된 태스크, 에이전트 출력 요약을 `.autopus/pipeline-state/{SPEC-ID}.yaml`에 체크포인트하여, 세션 재시작 시 `--continue` 플래그로 중단 지점부터 재개할 수 있다.

**Acceptance Criteria**:
- AC3.1: 각 Phase 완료 시 체크포인트 파일이 자동 생성된다
- AC3.2: `--continue`가 체크포인트를 읽고 다음 Phase부터 시작한다
- AC3.3: 체크포인트에 Phase 번호, 태스크 상태, 마지막 에이전트 결과 요약이 포함된다
- AC3.4: 코드베이스 변경 감지 — 체크포인트의 git commit hash와 현재 HEAD가 다르면 "stale checkpoint" 경고를 표시한다

**복원 수준 (Restoration Level)**:
- **Phase 단위 재시작**: `--continue`는 마지막으로 완료된 Phase 다음부터 실행한다. 예: Phase 2에서 중단 시 Phase 2를 처음부터 재실행.
- **에이전트 출력 비복원**: Claude Code 세션 종료 시 LLM 컨텍스트 윈도우가 소실되므로, 이전 세션의 에이전트 대화 내용은 복원하지 않는다. 대신 체크포인트에 저장된 결과 요약(태스크 완료 목록, 수정된 파일, 커버리지 수치)을 새 세션의 에이전트에 입력으로 전달한다.
- **태스크 단위 재개 불가**: 개별 태스크 중간 지점으로의 복원은 지원하지 않는다. Phase 내 태스크는 처음부터 재실행된다.

**Stale Checkpoint 감지**:
- 체크포인트 YAML에 `git_commit_hash` 필드를 포함한다
- `--continue` 실행 시 현재 HEAD와 비교하여, 불일치 시 경고를 표시하고 사용자에게 계속 진행 여부를 확인한다
- `--auto` 모드에서는 stale checkpoint를 무시하고 Phase부터 재실행한다

**Dependency**: 파이프라인 실행은 현재 스킬 기반(agent-pipeline.md)으로 동작. 체크포인트는 스킬 내에서 파일 I/O로 구현 가능 — 별도 엔진 불필요.

**Implementation Note**: 파이프라인 엔진은 Go 코드가 아닌 스킬/에이전트 수준에서 동작한다. 체크포인트는 에이전트가 각 Phase 완료 시 YAML 파일을 Write하는 방식으로 구현한다. Go binary 측에서는 `auto go --continue`가 체크포인트 파일의 존재를 확인하고 SPEC-ID와 마지막 Phase 정보를 파싱하여 파이프라인 스킬에 전달하는 얇은 래퍼 역할만 수행한다.

### Tier 2: Should Have — High

#### ~~R4: Community Skill Registry~~ → **보류 (Autopus 플랫폼 연동 예정)**
> 설계 결정: 커뮤니티 스킬 레지스트리는 향후 Autopus 플랫폼과 연결하여 구현. ADK 단독으로 GitHub 기반 레지스트리를 만들지 않음. Ref: SPEC-MARKET-001 (draft, 보류)

#### R5: Hard Gate Enforcement
WHEN `autopus.yaml`에서 `gates.mode: mandatory`로 설정될 때 THE SYSTEM SHALL 파이프라인 스킬 내 Gate 로직에서 `auto check` CLI 명령을 호출하여 게이트 조건 미충족 시 파이프라인 진행을 차단한다.

**Acceptance Criteria**:
- AC5.1: `autopus.yaml`에 `gates.mode: mandatory | advisory` 설정이 존재한다
- AC5.2: mandatory 모드에서 테스트 미작성 시 구현 Phase가 차단된다
- AC5.3: advisory 모드(기본값)에서는 기존 동작 유지 — 경고만 표시
- AC5.4: `auto init`이 mandatory 모드 선택 시 해당 설정을 `autopus.yaml`에 기록한다

**Gate 트리거 메커니즘 (Architecture Decision)**:
- **스킬 레벨 강제**: Gate 차단은 Claude Code hooks가 아닌, 파이프라인 스킬(`agent-pipeline.md`)의 Gate 정의에서 구현한다. 각 Gate 전에 `auto check --gate {gate-name}` CLI 명령을 Bash tool로 실행하여 exit code를 확인한다.
- **`auto check` 확장**: 기존 `internal/cli/check.go`에 `--gate` 플래그를 추가한다. `--gate phase2`는 Phase 2 진입 조건(테스트 존재 여부, lint 통과)을 검사하고, `mandatory` 모드에서 조건 미충족 시 exit 1을 반환한다.
- **Phase 경계 정의**: Gate 1(Phase 1→2 전환), Gate 2(Phase 2→3 전환), Gate 3(Phase 3→4 전환)에서 각각 검사한다.
- **Claude Code hooks 미사용**: `PreToolUse`/`PostToolUse`는 도구 호출 단위이지 Phase 전환 단위가 아니므로, Gate 강제에 사용하지 않는다. 파이프라인 스킬이 직접 `auto check` 결과를 읽고 분기한다.

#### ~~R6: Extended Provider Support~~ → **삭제됨**
> 리뷰 피드백: Autopus-ADK는 하네스이지 LLM 라우터가 아니다. 모델 라우팅은 Claude Code 또는 사용자 환경의 책임. 대신 기존 orchestra 기능에서 MCP 서버 연동 가이드를 제공하는 것으로 대체.

### Tier 3: Could Have — Medium

#### R7: Meta-Agent System
WHEN 사용자가 `auto skill create` 또는 `auto agent create`를 실행할 때 THE SYSTEM SHALL 기존 스킬/에이전트 패턴을 분석하여 새로운 스킬/에이전트 스켈레톤을 생성하는 빌더 에이전트를 제공한다.

**Acceptance Criteria**:
- AC7.1: 생성된 에이전트/스킬이 기존 규칙(file-size-limit, lore-commit 등)을 준수한다
- AC7.2: 생성 후 자동 검증 (frontmatter 유효성, 트리거 중복 확인)을 수행한다
- AC7.3: 사용자 승인 없이 자동 배포하지 않는다 (dry-run 기본)

**Safety**: 자동 생성된 콘텐츠는 반드시 사용자 확인 후 적용. `--auto` 모드에서도 검증 게이트 통과 필수.

#### R8: Reaction Engine
WHEN 사용자가 `auto react check` 명령을 실행하면 THE SYSTEM SHALL `gh` CLI를 통해 최근 CI 실행 상태를 조회하고, 실패한 워크플로우의 로그를 분석하여 수정 보고서를 생성한다.

**Acceptance Criteria**:
- AC8.1: `auto react check`가 `gh run list --status failure`로 실패한 CI를 탐지한다
- AC8.2: 실패 로그 분석 후 수정 보고서를 `.autopus/react/{run-id}.md`에 저장한다
- AC8.3: 사용자 명시적 승인(`auto react apply {run-id}`) 후에만 수정 실행
- AC8.4: 수정 범위는 실패한 파일에 한정 (blast radius 제한)
- AC8.5: 롤백 전략: 수정 전 자동 git stash

**감지 메커니즘 (Architecture Decision)**:
- **수동 트리거 방식**: `auto react check` CLI 명령을 사용자가 실행. 자동 폴링이나 webhook은 이 SPEC 범위 외.
- **`gh` CLI 의존**: GitHub Actions 상태 조회에 `gh run list`, `gh run view --log`를 사용한다. `gh` CLI가 설치되지 않은 경우 에러와 설치 가이드를 표시한다.
- **향후 자동화**: scheduled trigger(`/auto schedule`)로 주기적 `auto react check`를 설정하는 것은 별도 SPEC에서 다룬다.
- **Ref**: SPEC-REACT-001 (기존 draft)

#### ~~R9: Context Window Monitor~~ → **삭제됨 (향후 별도 SPEC)**
> 리뷰 피드백: Claude Code가 이미 자동 컨텍스트 압축을 수행. 하네스 레벨에서 LLM 런타임 상태에 접근할 API가 없음. MCP 프록시 기반 아키텍처로 전환 시 별도 SPEC에서 재검토.

#### R10: Deep Worker Agent Definition
WHEN 복잡한 탐색+구현 작업이 필요할 때 THE SYSTEM SHALL 장시간 자율 작업에 최적화된 `deep-worker` 에이전트 정의(.md 파일)를 제공한다.

**Acceptance Criteria**:
- AC10.1: `.claude/agents/autopus/deep-worker.md` 에이전트 정의가 존재한다
- AC10.2: 에이전트 프롬프트에 태스크 분해 지시, 체크포인트 파일 저장 지시, 검증 루프 지시가 포함된다
- AC10.3: 기존 `general-purpose` 대비 장시간(30분+) 작업에 최적화된 프롬프트 구조

**구현 범위 (Architecture Decision)**:
- **프롬프트 전용**: `deep-worker`는 `.md` 파일 기반 에이전트 정의이다. Go binary 변경은 없다.
- **상속 미사용**: Claude Code 에이전트는 상속 개념이 없으므로, `general-purpose`를 "확장"하는 것이 아니라 독립적인 에이전트 정의를 작성한다. `general-purpose`의 유용한 패턴(도구 사용 가이드라인, 에러 처리 등)은 프롬프트 텍스트로 포함한다.
- **체크포인트 연동**: R3(Pipeline State Persistence)의 체크포인트 형식을 재사용한다. 에이전트 프롬프트에서 `Write` 도구로 `.autopus/pipeline-state/` 디렉토리에 YAML 파일을 저장하도록 지시한다.
- **검증 루프**: 에이전트 프롬프트에 "구현 후 `go test`를 실행하고, 실패 시 수정하는 루프를 최대 N회 반복하라"는 지시를 포함한다. 이는 LLM의 reasoning 능력에 의존하며, 하드웨어적 루프가 아닌 프롬프트 기반 자기 조절이다.

## Out of Scope

- 자체 LLM 모델 학습/파인튜닝
- IDE 플러그인 (VS Code, JetBrains 등) — 별도 SPEC으로 분리
- LLM 호출 라우팅 / 프록시 — 하네스 아키텍처 경계 밖
- 컨텍스트 윈도우 런타임 모니터링 — LLM API 접근 불가
- 유료 SaaS 기능 (클라우드 호스팅, 팀 관리 등)

## 기존 SPEC 관계

| 갭 항목 | 기존 SPEC | 상태 |
|--------|----------|------|
| ~~Lite Distribution~~ | — | 삭제 (설계 결정: Go binary only) |
| Multi-Language SigMap | SPEC-SIGMAP-001 | completed (Go), 확장 필요 |
| Pipeline State | — | 신규 |
| ~~Community Skill Registry~~ | SPEC-MARKET-001 | 보류 (Autopus 플랫폼 연동 예정) |
| Hard Gate Enforcement | — | 신규 |
| Meta-Agent | — | 신규 |
| Reaction Engine | SPEC-REACT-001 | draft |
| Deep Worker | — | 신규 |

## 구현 순서

이 SPEC은 로드맵 SPEC으로, 직접적인 코드 변경 없이 하위 SPEC들의 우선순위와 구현 순서를 정의한다. 각 R(요구사항)은 독립 SPEC으로 분해되어 구현된다.

```
Phase 1 (Critical): R3 → R2
Phase 2 (High):     R5
Phase 3 (Medium):   R7 → R8 → R10
```

R3(Pipeline State Persistence)이 최우선 — 세션 중단 시 작업 손실 방지가 가장 시급.
