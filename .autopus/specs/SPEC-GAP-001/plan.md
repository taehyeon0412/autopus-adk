# SPEC-GAP-001 구현 계획

## 로드맵 개요

3개 Tier를 순차적으로 진행하되, Tier 내 태스크는 병렬 가능한 것끼리 묶어 실행한다. 각 태스크는 독립 SPEC을 생성/갱신하고 구현하는 단위이다.

## Tier 1: Must Have (시장 생존) — 예상 4-6주

### T1: Plugin Distribution 패키징
- SPEC-PUB-001 갱신 또는 신규 SPEC 생성
- Claude Plugin manifest 구조 설계 (.claude/ 기반 배포)
- Go binary와 Plugin 형식 이중 배포 파이프라인 구축
- `auto publish` 명령어에 Plugin 타겟 추가
- 의존: 없음

### T2: Multi-Language SigMap 확장
- SPEC-SIGMAP-001 확장 또는 SPEC-SIGMAP-002 신규 생성
- Language Detector 모듈 구현 (go.mod, package.json, Cargo.toml, pom.xml 등)
- Tree-sitter 기반 범용 AST 파서 통합 (Go의 go/ast 대체)
- 언어별 시그니처 추출 어댑터: TypeScript, Python, Rust, Java
- SigMap 출력 포맷 언어 무관하게 통일
- 의존: 없음 (T1과 병렬 가능)

### T3: Multi-Language Testing Strategy
- 테스팅 스킬(tdd.md, testing-strategy.md)의 언어 확장
- 언어별 테스트 러너 매핑 (jest, pytest, cargo test, gradle test 등)
- 커버리지 도구 통합 (istanbul, coverage.py, tarpaulin, jacoco)
- 의존: T2 (Language Detector 활용)

### T4: Pipeline State Persistence
- 신규 SPEC 생성 (SPEC-PERSIST-001 등)
- `.autopus/pipeline-state/` 디렉토리 구조 설계
- Phase별 체크포인트 직렬화 (YAML 또는 JSON)
- `auto go --resume` 플래그로 중단 지점 재개
- RALF loop 상태 포함 (재시도 횟수, 실패 기록)
- 의존: 없음 (T1, T2와 병렬 가능)

## Tier 2: Should Have (경쟁 균형) — 예상 3-4주

### T5: Hard Gate Enforcement
- SPEC-GUARD-001 갱신 또는 신규 SPEC 생성
- 기존 advisory Gate를 mandatory로 전환하는 설정 추가
- Hook 기반 강제 차단 메커니즘 (pre-phase hook)
- 게이트 실패 시 rollback 또는 재시도 정책
- autopus.yaml에 `gates.mode: mandatory | advisory` 설정
- 의존: T4 (상태 persistence 필요)

### T6: Community Skill Marketplace
- SPEC-MARKET-001 갱신 (기존 draft 활용)
- GitHub 기반 레지스트리 인덱스 설계
- `auto skill search/install/publish` CLI 확장
- 스킬 버저닝 및 호환성 검사
- 의존: T1 (Plugin Distribution 인프라 활용)

### T7: Extended Provider Support
- 신규 SPEC 생성 (SPEC-PROVIDER-001 등)
- orchestra 엔진 프로바이더 어댑터 인터페이스 확장
- Ollama 어댑터 (로컬 모델, 오프라인 지원)
- OpenRouter 어댑터 (모델 선택 UI)
- Perplexity 어댑터 (검색 강화 리서치)
- 의존: 없음 (독립 실행 가능)

## Tier 3: Could Have (경쟁 우위) — 예상 4-6주

### T8: Meta-Agent System
- 신규 SPEC 생성
- 에이전트 실행 패턴 로깅 및 분석
- 스킬/에이전트 템플릿 자동 생성
- MoAI의 Builder tier 참고, 안전 장치 설계 (생성된 에이전트 검증)
- 의존: T6 (Marketplace에 게시 가능)

### T9: Reaction Engine 구현
- SPEC-REACT-001 활용 (기존 draft)
- CI 실패 감지 → debugger 에이전트 자동 호출
- PR 코멘트 분석 → 코드 변경 자동 시도
- 안전 장치: 브랜치 전용, PR당 최대 3회
- 의존: T5 (Gate Enforcement와 연동)

### T10: Context Window Monitor
- 신규 SPEC 생성
- 토큰 사용량 추적 모듈 (telemetry 확장)
- 70% 경고, 85% 자동 압축 트리거
- 컨텍스트 압축 전략 (요약, 오래된 대화 제거)
- 의존: 없음

### T11: Deep Worker Mode
- 신규 SPEC 생성
- 장시간 실행 에이전트 아키텍처 (탐색-실행-검증 루프)
- 목표 분해 및 진행률 추적
- experiment loop 확장
- 의존: T4 (상태 persistence), T10 (컨텍스트 관리)

## 구현 전략

### 기존 코드 활용

| 갭 항목 | 활용 가능한 기존 코드 |
|--------|-------------------|
| Plugin Distribution | `pkg/plugin/`, `internal/cli/setup.go`, `pkg/adapter/` |
| Multi-Language | `pkg/sigmap/`, `pkg/detect/` (Language Detector) |
| State Persistence | `pkg/experiment/` (experiment loop 상태 관리 패턴) |
| Gate Enforcement | `internal/cli/check.go`, `internal/cli/check_rules.go` |
| Provider Support | `pkg/orchestra/` (기존 프로바이더 어댑터) |
| Reaction Engine | `pkg/issue/`, Hook 시스템 |
| Context Monitor | `pkg/telemetry/`, `pkg/cost/` |

### 변경 범위 추정

- Tier 1: 15-20개 신규/변경 파일, 2000-3000 LOC
- Tier 2: 10-15개 신규/변경 파일, 1500-2000 LOC
- Tier 3: 15-20개 신규/변경 파일, 2000-3000 LOC
- 총계: 40-55개 파일, 5500-8000 LOC

### 위험 요소

1. **Tree-sitter 통합 복잡도** — Go에서 Tree-sitter 바인딩이 CGO 의존성 추가. 대안: Language Server Protocol 활용
2. **Plugin manifest 규격 변동** — Claude Plugin 규격이 아직 안정화되지 않았을 가능성
3. **상태 직렬화 호환성** — 파이프라인 버전 업그레이드 시 기존 체크포인트 호환 필요
