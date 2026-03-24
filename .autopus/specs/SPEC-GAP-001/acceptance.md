# SPEC-GAP-001 수락 기준

## Tier 1: Must Have (시장 생존)

### S1: Plugin Distribution으로 설치 가능
- Given: 사용자가 Claude Code 환경에서 Autopus-ADK를 처음 사용하려 함
- When: `plugin install autopus` 또는 이에 준하는 Plugin marketplace 설치 명령 실행
- Then: .claude/ 디렉토리에 rules, skills, commands, agents가 배포되고, `/auto` 명령이 즉시 사용 가능하다
- Then: Go binary 설치 없이도 핵심 기능(plan, go, review)이 동작한다

### S2: 5개 주요 언어에서 SigMap 동작
- Given: TypeScript(package.json), Python(pyproject.toml), Rust(Cargo.toml), Java(pom.xml), Go(go.mod) 프로젝트 각각이 존재
- When: `auto setup`을 실행
- Then: 각 언어의 public API 시그니처가 `.autopus/context/signatures.md`에 추출된다
- Then: 시그니처 형식이 언어에 무관하게 통일된 포맷이다

### S3: 5개 주요 언어에서 테스트 실행
- Given: 각 언어 프로젝트에서 테스트 파일이 존재
- When: 파이프라인 Phase 3(Testing)이 실행
- Then: 언어에 맞는 테스트 러너(jest, pytest, cargo test, gradle test, go test)가 자동 선택되어 실행된다
- Then: 커버리지 결과가 통합 형식으로 보고된다

### S4: 파이프라인 중단 후 재개
- Given: 5-Phase 파이프라인이 Phase 2(Implementation) 진행 중
- When: 세션이 중단(네트워크 끊김, 사용자 종료 등)되고 재시작
- Then: `auto go --resume SPEC-ID` 실행 시 Phase 2부터 재개된다
- Then: Phase 1의 결과(계획, 테스트 스캐폴드)가 보존되어 재실행 불필요
- Then: 중단 전 완료된 executor 워크트리의 변경사항이 보존된다

## Tier 2: Should Have (경쟁 균형)

### S5: Mandatory Gate가 다음 Phase 진입을 차단
- Given: autopus.yaml에 `gates.mode: mandatory` 설정
- When: Gate 1(Approval)에서 테스트 미작성 감지
- Then: Phase 2(Implementation) 진입이 차단되고, 사용자에게 필수 조건 안내
- Then: advisory 모드 전환 시 경고만 출력하고 진행 허용

### S6: 커뮤니티 스킬 검색 및 설치
- Given: GitHub 기반 스킬 레지스트리가 운영 중
- When: `auto skill search "testing"` 실행
- Then: 커뮤니티 공유 스킬 목록이 표시된다
- When: `auto skill install @community/advanced-tdd` 실행
- Then: 스킬이 .claude/skills/autopus/에 설치되고 즉시 사용 가능하다

### S7: Ollama 로컬 모델로 오프라인 실행
- Given: Ollama가 로컬에 설치되고 모델이 다운로드된 상태
- When: autopus.yaml에 `providers.ollama.enabled: true` 설정 후 파이프라인 실행
- Then: 인터넷 연결 없이 파이프라인이 실행된다 (품질은 모델 성능에 의존)

## Tier 3: Could Have (경쟁 우위)

### S8: Meta-Agent가 반복 패턴에서 스킬 자동 생성
- Given: 동일 유형의 에이전트 태스크가 5회 이상 반복 감지
- When: Meta-Agent 분석이 트리거
- Then: 해당 패턴을 자동화하는 스킬 초안이 생성되고 사용자 승인을 요청한다
- Then: 승인 시 .claude/skills/autopus/에 배치된다

### S9: CI 실패 자동 분석 및 수정 제안
- Given: GitHub Actions CI가 테스트 실패로 빨간불
- When: post-push Hook이 CI 상태를 감지
- Then: 실패 로그를 분석하고 수정 커밋을 별도 브랜치에 생성한다
- Then: 자동 수정은 PR당 최대 3회로 제한된다
- Then: main/master에 직접 푸시하지 않는다

### S10: 컨텍스트 윈도우 70% 경고
- Given: 에이전트가 대규모 코드베이스를 분석 중
- When: 컨텍스트 윈도우 사용률이 70%에 도달
- Then: 사용자에게 경고가 표시된다
- When: 85%에 도달
- Then: 자동 컨텍스트 압축이 실행되어 오래된 대화/중복 정보가 제거된다

### S11: Deep Worker가 장시간 목표를 자율 수행
- Given: 사용자가 "이 모듈 전체를 리팩터링해줘" 같은 장시간 목표를 제시
- When: Deep Worker 모드가 활성화
- Then: 에이전트가 독립적으로 탐색-계획-실행-검증을 반복하며 진행률을 보고한다
- Then: 중간 체크포인트로 상태가 보존된다

## 전체 로드맵 수락 기준

### S12: 경쟁사 대비 기능 달성률
- Given: Tier 1 완료 시
- Then: Superpowers와 동등한 배포 접근성, MoAI 대비 5개 이상 언어 지원, 세션 중단 복구 기능 확보
- Given: Tier 2 완료 시
- Then: 4개 경쟁사 대비 기능 격차 없음 (parity), Autopus 고유 강점(파이프라인, RALF, @AX, Lore)은 유지
- Given: Tier 3 완료 시
- Then: 경쟁사 대비 차별화된 고유 기능 3개 이상 보유 (Meta-Agent, Reaction Engine + 기존 강점)
