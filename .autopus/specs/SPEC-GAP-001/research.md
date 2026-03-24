# SPEC-GAP-001 리서치

## 기존 코드 분석

### 현재 아키텍처 구조

```
cmd/auto/              — CLI 엔트리포인트
internal/cli/          — 명령어 핸들러 (root.go, setup.go, spec.go, orchestra.go 등)
pkg/
  adapter/             — Claude Code 어댑터 (.claude/ 파일 관리)
  arch/                — 아키텍처 분석
  config/              — autopus.yaml 설정
  constraint/          — 제약 조건 검증
  content/             — 콘텐츠 처리
  cost/                — 비용 추정
  detect/              — 언어/프레임워크 감지
  e2e/                 — E2E 테스트
  experiment/          — 실험 루프
  issue/               — 이슈 관리
  lore/                — Lore 커밋
  lsp/                 — LSP 통합
  orchestra/           — 멀티프로바이더 오케스트레이션
  plugin/              — 플러그인 시스템
  search/              — 코드 검색
  selfupdate/          — 자동 업데이트
  setup/               — 프로젝트 설정 (SigMap 통합 포함)
  sigmap/              — AST 기반 시그니처 맵
  spec/                — SPEC 관리
  telemetry/           — 텔레메트리
  template/            — 템플릿 엔진
  version/             — 버전 관리
```

### 갭별 관련 코드

#### Plugin Distribution
- `pkg/plugin/` — 기존 플러그인 인프라 존재
- `pkg/adapter/` — .claude/ 파일 배포 로직 (rules, skills, agents, commands)
- `internal/cli/setup.go` — `auto setup` 명령, adapter 호출로 .claude/ 배포
- 현재 Go binary(`cmd/auto/`) + `auto setup`으로 .claude/ 배포하는 2단계 설치
- **갭**: Claude Plugin manifest 형식 미지원, marketplace 등록 프로세스 부재

#### Multi-Language Support
- `pkg/sigmap/` — 현재 Go 전용 (`go/ast`, `go/parser` 사용)
- `pkg/setup/sigmap_integration.go` — SigMap을 setup 단계에서 호출
- `pkg/detect/` — 언어 감지 모듈 (go.mod, package.json 등 파일로 감지)
- `.claude/skills/autopus/tdd.md`, `testing-strategy.md` — Go 중심 테스트 전략
- **갭**: SigMap이 Go AST에 하드코딩, 언어별 어댑터 패턴 부재

#### Pipeline State Persistence
- `.claude/skills/autopus/agent-pipeline.md` — 5-Phase 파이프라인 정의
- `pkg/experiment/` — experiment loop에서 상태 관리 패턴 존재 (참고 가능)
- `internal/cli/experiment.go`, `experiment_helpers.go` — 실험 상태 추적
- **갭**: 파이프라인 Phase별 체크포인트 메커니즘 부재, 중단 시 전체 재시작 필요

#### Gate Enforcement
- `internal/cli/check.go`, `check_rules.go` — 규칙 검사 로직
- `internal/cli/verify.go`, `verify_types.go` — 검증 로직
- **갭**: Gate가 advisory (경고만), mandatory 차단 미구현

#### Provider Support
- `pkg/orchestra/` — 멀티프로바이더 오케스트레이션 엔진
- `internal/cli/orchestra.go`, `orchestra_config.go` — 프로바이더 설정
- **갭**: Ollama, OpenRouter, Perplexity 어댑터 미구현

## 경쟁사 심층 분석

### Superpowers (89K stars)
- **배포**: Claude Plugin marketplace 네이티브. `plugin install superpowers`로 원클릭 설치
- **Mandatory Gates**: 테스트 미작성 시 구현 코드 삭제. 코드 작성 전 테스트 강제
- **Community Skills**: GitHub repo에서 community-editable skills. PR로 스킬 기여
- **약점**: 단일 에이전트, 파이프라인 없음, 의사결정 추적 없음

### MoAI-ADK
- **규모**: 27 agents (4 tiers: Starter/Pro/Expert/Builder), 52 skills
- **Session Persistence**: `progress.md`에 실행 상태 자동 저장, 재개 가능
- **Multi-Language**: 18개 언어 지원 (LSP 기반)
- **Meta-Agent**: Builder tier에서 새 에이전트/스킬 자동 생성
- **약점**: Go binary 전용 배포, 커뮤니티 생태계 약함

### OhMyOpenAgent (OMO)
- **Hashline Edit**: 콘텐츠 해시 앵커 기반 편집 (줄 번호 불안정성 해결)
- **Boulder System**: zero-context-loss 재개. 세션 종료 후 완전 복구
- **LSP+AST-Grep**: 25+ 언어에서 구조적 코드 검색/변환
- **tmux 통합**: 병렬 에이전트를 tmux pane에서 시각화
- **약점**: 설치 복잡, 학습 곡선 높음

### Claude Octopus
- **8 AI Providers**: Claude, GPT-4, Gemini, DeepSeek 등 멀티 프로바이더
- **Double Diamond**: 발산-수렴 2회 반복 방법론
- **Reaction Engine**: CI 실패 → 자동 분석 → 수정 PR. PR 코멘트 → 자동 대응
- **75% Consensus Gate**: 멀티 모델 합의 기반 품질 게이트
- **약점**: 복잡한 설정, 비용 높음

## 설계 결정

### D1: Multi-Language SigMap — Tree-sitter vs LSP

| 옵션 | 장점 | 단점 |
|------|------|------|
| **Tree-sitter** | 빠름, 오프라인, 일관된 AST | CGO 의존성, 언어별 grammar 관리 |
| **LSP** | 언어 서버 재사용, 정확한 타입 정보 | 서버 기동 시간, 메모리, 설치 의존성 |
| **regex 기반** | 의존성 없음, 빠른 구현 | 부정확, 복잡한 구문 처리 불가 |

**결정**: Tree-sitter를 1차 선택. Go에서 `go-tree-sitter` 바인딩 사용. CGO 의존성은 빌드 시 static link로 해결. 대안으로 `tree-sitter-cli`를 subprocess로 호출하는 방식도 검토.

### D2: Pipeline State Persistence — 저장 형식

| 옵션 | 장점 | 단점 |
|------|------|------|
| **YAML** | 사람이 읽기 쉬움, 기존 config와 일관 | 대용량 데이터에 부적합 |
| **JSON** | 파싱 빠름, 구조화 | 사람이 읽기 어려움 |
| **SQLite** | 쿼리 가능, 대용량 | 바이너리, 복잡도 증가 |

**결정**: YAML. 파이프라인 상태는 소량(Phase 정보, 태스크 목록, 에이전트 출력 경로)이므로 YAML로 충분. 에이전트 출력 자체는 별도 .md 파일로 저장.

### D3: Plugin Distribution — 배포 전략

현재 Autopus-ADK는 Go binary + `auto setup`으로 .claude/ 배포하는 2단계 설치. Claude Plugin 형식은 .claude/ 디렉토리 자체가 Plugin이므로, Go binary 없이도 rules/skills/agents/commands만으로 핵심 기능 제공 가능.

**결정**: 2-track 배포.
1. **Plugin-only**: .claude/ 파일만 배포. Go binary 없이 동작하는 "lite" 모드. SigMap, telemetry 등 Go 의존 기능 비활성화.
2. **Full**: 기존 Go binary + Plugin. 모든 기능 사용 가능.

### D4: Gate Enforcement — 구현 방식

기존 `check.go`의 규칙 검사 결과를 파이프라인 게이트에서 강제하려면, 파이프라인 스킬(`agent-pipeline.md`)의 Gate 정의를 확장해야 한다.

**결정**: autopus.yaml에 `gates.mode` 설정 추가. mandatory 모드에서는 Gate 실패 시 파이프라인 상태를 "blocked"로 전환하고, 사용자가 조건 충족 후 `auto go --resume`으로 재개.

### D5: Extended Providers — 어댑터 인터페이스

`pkg/orchestra/`의 기존 프로바이더 인터페이스를 확장하여 Ollama, OpenRouter, Perplexity를 추가.

**결정**: 기존 orchestra 엔진의 프로바이더 어댑터 패턴을 그대로 활용. 각 프로바이더는 `Provider` 인터페이스를 구현하는 별도 파일로 분리.

## 미결 사항

1. **Claude Plugin manifest 최신 규격** — Anthropic 공식 문서에서 Plugin manifest 형식 확인 필요
2. **Tree-sitter Go 바인딩 안정성** — `go-tree-sitter`의 최신 버전 호환성 검증 필요
3. **Boulder vs Checkpoint** — OMO의 Boulder 시스템이 정확히 어떤 수준의 컨텍스트를 보존하는지 추가 조사 필요
4. **Ollama 모델 품질** — 로컬 모델로 에이전트 파이프라인을 실행할 때 최소 품질 기준 정의 필요
5. **기존 draft SPEC(MARKET-001, REACT-001) 갱신 범위** — 기존 SPEC을 수정할지, 새 SPEC으로 대체할지 결정 필요
