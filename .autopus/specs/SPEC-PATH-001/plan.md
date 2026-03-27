# SPEC-PATH-001 구현 계획

## 태스크 목록

- [ ] T1: auto-router.md.tmpl에 SPEC Path Resolution 공통 섹션 삽입
- [ ] T2: auto-router.md.tmpl의 go 서브커맨드에 WORKING_DIR 주입 로직 추가
- [ ] T3: auto-router.md.tmpl의 plan 서브커맨드에 --target 플래그 및 모듈 결정 로직 추가
- [ ] T4: auto-router.md.tmpl의 sync 서브커맨드에 TARGET_MODULE 기반 git 작업 추가
- [ ] T5: auto-router.md.tmpl의 status 서브커맨드에 전체 글로빙 및 모듈별 그룹화 추가
- [ ] T6: spec-writer.md의 저장 위치 규칙을 target module 기반으로 변경
- [ ] T7: planner.md의 SPEC 경로 참조를 resolution 기반으로 변경
- [ ] T8: prd.md의 SPEC 탐색/저장 경로를 resolution 기반으로 변경
- [ ] T9: idea.md의 BS 파일 저장을 모듈 인식 로직으로 변경

## 구현 전략

### 접근 방법

1. **공통 Resolution 절차를 auto-router에 정의** (T1): 모든 서브커맨드가 참조하는 단일 source of truth. `resolve_spec(SPEC_ID)` 의사코드를 Markdown 섹션으로 정의하고, 서브커맨드별로 이를 참조하도록 한다.

2. **서브커맨드별 적용** (T2-T5): go, plan, sync, status 각각에서 resolution 결과를 활용하는 방식을 기술. 특히 go는 WORKING_DIR을 executor에 전달하고, sync는 `git -C {TARGET_MODULE}` 패턴을 사용.

3. **에이전트/스킬 수정** (T6-T9): spec-writer, planner, prd, idea의 하드코딩된 `.autopus/specs/` 경로를 resolution 기반 또는 target module 기반으로 변경.

### 기존 코드 활용

- auto-router.md.tmpl의 기존 "Context Load" 섹션 패턴을 따라 "SPEC Path Resolution" 섹션을 배치
- spec-writer.md의 "SPEC 저장 위치 규칙" 섹션을 교체
- idea.md의 "저장 위치 규칙" 섹션을 교체

### 변경 범위

- 5개 파일 수정 (신규 파일 없음)
- 모든 변경은 Markdown 콘텐츠 (프롬프트 엔지니어링) 수정
- Go 소스 코드 변경 없음
- 레거시 SPEC과의 하위 호환성 보장 (resolution 순서 1번)

### 주의사항

- auto-router.md.tmpl은 ~1100줄의 대형 템플릿이므로, Resolution 섹션은 Context Load 직후에 삽입하여 모든 서브커맨드보다 앞에 위치시킴
- worktree 이슈(서브모듈 독립 git root)는 이번 SPEC 범위 외 — 별도 SPEC으로 분리
