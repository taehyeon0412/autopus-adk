# PRD: --multi Hook 기반 멀티프로바이더 오케스트레이션

> Product Requirements Document — Standard (10-section format).

- **SPEC-ID**: SPEC-ORCH-007
- **Author**: Autopus Planner Agent
- **Status**: Draft
- **Date**: 2026-03-26
- **Extends**: SPEC-ORCH-006 (인터랙티브 pane 모드)
- **Origin**: BS-002 (Hook 기반 멀티프로바이더 오케스트레이션 브레인스토밍)

---

## 1. Problem & Context

**현재 상황**

SPEC-ORCH-006은 cmux/tmux pane에서 프로바이더 CLI를 인터랙티브 세션으로 실행하고, `ReadScreen` 폴링으로 완료를 감지하며 결과를 수집하는 구조를 구현했다. 현재 `interactive.go`의 `waitForCompletion()`은 500ms 간격으로 `ReadScreen`을 호출하여 화면 텍스트에서 프롬프트 패턴을 감지하고, `isOutputIdle()`로 10초간 출력이 없으면 완료로 판정한다.

**문제**

ReadScreen 화면 스크래핑 방식에는 근본적 한계가 있다:

1. **결과 파싱 불안정**: ANSI 이스케이프, 프롬프트 장식, 스크롤 위치에 따라 결과 수집 품질이 가변적이다. `cleanScreenOutput()`이 `stripANSI()` + `filterPromptLines()`을 적용하지만, 프로바이더마다 출력 포맷이 달라 완전한 정제가 어렵다.
2. **완료 감지 불확실**: 프롬프트 패턴 매칭(`isPromptVisible()`)과 idle 감지의 이중 전략이 false positive(중간 일시정지를 완료로 오판)와 false negative(프롬프트 패턴 불일치로 타임아웃)를 모두 발생시킨다.
3. **프로바이더 통합 비용**: 새 프로바이더 추가 시마다 `DefaultCompletionPatterns()`에 정규식을 추가하고, `defaultPromptPatterns`을 갱신해야 하며, 각 프로바이더의 CLI 출력 형식을 수동으로 분석해야 한다.
4. **Codex CLI 대체 필요**: 현재 `codex` 프로바이더는 OpenAI의 Codex CLI인데, opencode(sst/anomalyco)가 동일한 API를 더 확장 가능한 plugin 시스템으로 지원한다.

**영향**

- `--multi` 실행 시 결과 수집 실패율이 높아 사용자가 수동으로 각 pane 결과를 확인해야 하는 경우가 빈번하다.
- Consensus/debate 전략의 정확도가 원시 결과의 품질에 직접 의존하므로, 오케스트레이션 가치가 반감된다.
- 프로바이더 확장이 "프롬프트 패턴 추가 + 테스트"의 수작업 사이클을 요구한다.

**변경 동기**

3개 주요 프로바이더(Claude Code, Gemini CLI, opencode)가 모두 hook/plugin 기반 결과 시그널을 제공한다:
- Claude Code: `Stop` hook (`last_assistant_message` 필드)
- Gemini CLI: `AfterAgent` hook (`prompt_response` 필드)
- opencode: `experimental.text.complete` plugin (`text` 필드)

이 hook/plugin 시스템을 활용하면 ReadScreen 스크래핑 없이 구조화된 JSON 결과를 파일로 받을 수 있다. BS-002에서 이 접근의 타당성을 검증했으며, 지금이 전환 적기이다.

---

## 2. Goals & Success Metrics

| 목표 | 성공 지표 | 목표값 | 일정 |
|------|----------|--------|------|
| Hook 기반 결과 수집으로 전환 | ReadScreen 호출 제거율 | 100% (결과 수집 경로) | Phase 1 |
| 결과 수집 신뢰성 향상 | 결과 수집 성공률 | >= 95% (현재 추정 ~70%) | Phase 1 완료 후 |
| Codex -> opencode 프로바이더 전환 | opencode plugin 연동 성공 | opencode 결과 수집 정상 동작 | Phase 2 |
| 오케스트레이션 전략 hook 연동 | 4개 전략(consensus/debate/relay/fastest) hook 결과 활용 | 전 전략 정상 동작 | Phase 3 |
| 프로바이더 확장 비용 감소 | 새 프로바이더 추가 시 필요 코드 | hook 스크립트 1개 + config 1줄 | Phase 1 이후 |

**Anti-Goals** (성공이 아닌 것)

- ReadScreen/PipePane 인프라 자체를 제거하는 것은 아니다. 이들은 hook 미설정 시 fallback 및 디버깅 용도로 유지한다.
- OAuth나 API 키 관리를 이 SPEC에서 다루지 않는다.
- 프로바이더 CLI 바이너리의 설치/업데이트를 자동화하지 않는다.

---

## 3. Target Users

| 사용자 그룹 | 역할 | 사용 빈도 | 핵심 기대 |
|------------|------|----------|----------|
| 로컬 개발자 | `auto orchestra --multi` 사용자 | 일일 | 안정적인 멀티프로바이더 결과 수집 및 자동 병합 |
| Autopus-ADK 기여자 | 프로바이더 확장 개발자 | 월간 | 새 프로바이더를 hook 스크립트 하나로 추가 가능 |
| CI/자동화 시스템 | 비대화식 환경 | 일일 | Graceful degradation으로 hook 미설정 시에도 동작 |

**Primary User**: 로컬 개발자 (여러 AI 프로바이더를 pane에서 동시 활용하여 consensus/debate/relay 실행)

---

## 4. User Stories

### Story 1: Hook 기반 자동 결과 수집

**As a** 로컬 개발자,
**I want** `auto orchestra --multi`가 각 프로바이더의 hook/plugin을 통해 결과를 자동 수집하도록,
**so that** ReadScreen 스크래핑의 불안정성 없이 깨끗한 결과를 받을 수 있다.

**Acceptance Criteria**

- Given hook이 자동 주입된 상태에서, when `auto orchestra --multi` 실행 시, then 각 프로바이더 완료 후 `/tmp/autopus/{session-id}/result.json`에 구조화된 결과가 저장된다.
- Given 결과 파일이 생성되면, when 파일 감시가 `done` 시그널을 감지하면, then ReadScreen 없이 결과를 수집하여 merge 로직에 전달한다.

---

### Story 2: Hook 자동 주입

**As a** 로컬 개발자,
**I want** `auto init` 또는 `auto orchestra` 실행 시 hook 스크립트가 자동으로 각 프로바이더 설정에 주입되도록,
**so that** 수동으로 hook을 설정할 필요 없이 바로 사용할 수 있다.

**Acceptance Criteria**

- Given Claude Code가 설치된 상태에서, when `auto init` 실행 시, then `.claude/settings.json`의 hooks.Stop에 결과 수집 스크립트가 등록된다.
- Given Gemini CLI가 설치된 상태에서, when `auto init` 실행 시, then `.gemini/settings.json`의 hooks.AfterAgent에 결과 수집 스크립트가 등록된다.
- Given opencode가 설치된 상태에서, when `auto init` 실행 시, then `opencode.json`의 experimental.text.complete plugin이 등록된다.

---

### Story 3: Graceful Degradation

**As a** 개발자,
**I want** hook이 설정되지 않은 프로바이더에 대해서도 오케스트레이션이 동작하도록,
**so that** 부분적 hook 설정 상태에서도 `--multi`를 사용할 수 있다.

**Acceptance Criteria**

- Given claude hook만 설정되고 gemini hook이 미설정일 때, when `auto orchestra --multi` 실행 시, then claude는 hook 결과를 사용하고 gemini는 기존 ReadScreen fallback으로 결과를 수집한다.
- Given 어떤 프로바이더도 hook이 없을 때, when 실행 시, then 전체가 기존 ReadScreen 모드(SPEC-ORCH-006)로 fallback한다.

---

### Story 4: opencode 프로바이더 전환

**As a** 개발자,
**I want** Codex CLI 대신 opencode를 멀티프로바이더 중 하나로 사용하도록,
**so that** opencode의 plugin 시스템을 통해 일관된 hook 기반 결과 수집이 가능하다.

**Acceptance Criteria**

- Given opencode 바이너리가 설치된 상태에서, when `auto orchestra --multi` 실행 시, then opencode가 codex 대신 프로바이더로 사용된다.
- Given autopus.yaml에 `codex` 프로바이더가 설정된 경우, when 마이그레이션 시, then `opencode`로 자동 전환되고 경고 메시지가 출력된다.

---

### Story 5: Consensus 전략 hook 연동

**As a** 개발자,
**I want** consensus 전략이 hook으로 수집된 구조화 결과를 활용하도록,
**so that** 더 정확한 합의 판정이 가능하다.

**Acceptance Criteria**

- Given 3개 프로바이더 hook 결과가 모두 수집된 상태에서, when consensus 전략 실행 시, then 기존 `MergeConsensus()` 대신 구조화된 JSON 결과를 직접 비교하여 66% 합의를 판정한다.

---

## 5. Functional Requirements

### P0 -- Must Have

| ID | 요구사항 | 비고 |
|----|---------|------|
| FR-01 | **파일 시그널 프로토콜 정의**: `/tmp/autopus/{session-id}/result.json` 형식 (`{ session_id, provider, response, timestamp }`)과 `/tmp/autopus/{session-id}/done` 시그널 파일 | BS-002 프로토콜 확정 |
| FR-02 | **Claude Code Stop hook 스크립트**: `last_assistant_message` 필드에서 응답 추출, result.json 저장, done 시그널 생성 | `.claude/settings.json` hooks.Stop 등록 |
| FR-03 | **Gemini CLI AfterAgent hook 스크립트**: `prompt_response` 필드에서 응답 추출, result.json 저장, done 시그널 생성 | `.gemini/settings.json` hooks.AfterAgent 등록 |
| FR-04 | **opencode plugin 스크립트**: `experimental.text.complete` 이벤트에서 `text` 필드 추출, result.json 저장, done 시그널 생성 | `opencode.json` plugin 등록 |
| FR-05 | **interactive.go 파일 감시 전환**: `waitForCompletion()`을 `done` 파일 감시 기반으로 변경. ReadScreen 폴링 제거 (결과 수집 경로) | `fsnotify` 또는 폴링 기반 |
| FR-06 | **결과 수집 전환**: `collectResults()`를 `result.json` 파싱 기반으로 변경. `cleanScreenOutput()` 불필요 | JSON 직접 파싱 |
| FR-07 | **Graceful degradation**: hook 미설정 프로바이더는 기존 ReadScreen fallback 사용. 프로바이더별 hook 존재 여부 감지 | 혼합 모드 지원 |
| FR-08 | **Hook 자동 주입**: `auto init` 시 각 프로바이더 설정 파일에 hook/plugin 엔트리 자동 등록. 기존 사용자 hook 보존 (merge) | 어댑터별 구현 |

### P1 -- Should Have

| ID | 요구사항 | 비고 |
|----|---------|------|
| FR-10 | **Codex -> opencode 프로바이더 전환**: `ProviderConfig`에서 codex를 opencode로 교체. autopus.yaml 마이그레이션 | `pkg/config/migrate.go` 확장 |
| FR-11 | **세션 ID 관리**: 오케스트레이션 실행마다 고유 세션 ID 생성, `/tmp/autopus/{session-id}/` 디렉토리 생성/정리 | 기존 `randomHex()` 활용 |
| FR-12 | **오케스트레이션 전략 hook 연동**: consensus/debate/relay/fastest 각 전략이 hook 결과 JSON을 직접 활용 | 기존 merge 로직 확장 |
| FR-13 | **Hook 스크립트 템플릿화**: hook 스크립트를 `templates/` 아래 템플릿으로 관리하여 버전 업데이트 용이 | `content/hooks/` 확장 |
| FR-14 | **completion 패턴 정리**: hook 모드에서 불필요해진 `DefaultCompletionPatterns()`, `defaultPromptPatterns` 코드 경로를 fallback 전용으로 격리 | 코드 정리 |

### P2 -- Could Have

| ID | 요구사항 | 비고 |
|----|---------|------|
| FR-20 | **Hook 상태 진단**: `auto doctor`에서 각 프로바이더의 hook/plugin 설정 상태를 검증하고 미설정 시 가이드 출력 | doctor.go 확장 |
| FR-21 | **Slash command 지원**: hook 내에서 `/auto` slash command를 프로바이더 세션에 주입 | 향후 확장 |
| FR-22 | **result.json 스키마 검증**: 결과 파일의 JSON 스키마 유효성 검증 | 방어적 프로그래밍 |

---

## 6. Non-Functional Requirements

| 카테고리 | 요구사항 | 목표 |
|---------|---------|------|
| 성능 | 파일 시그널 감지 지연 | done 파일 생성 후 500ms 이내 감지 |
| 성능 | result.json 파싱 시간 | < 10ms per provider |
| 안정성 | Hook 실패 시 시스템 영향 | Hook 실패가 프로바이더 CLI 동작에 영향 없음 (fire-and-forget) |
| 보안 | result.json 파일 권한 | 0o600 (소유자만 읽기/쓰기) |
| 보안 | 세션 디렉토리 경로 | `/tmp/autopus/{session-id}/` 형식, session-id는 랜덤 hex |
| 보안 | Hook 스크립트 경로 순회 방지 | 프로바이더 이름 sanitize (기존 `sanitizeProviderName()` 재사용) |
| 호환성 | 기존 SPEC-ORCH-006 코드 경로 | Fallback으로 완전 보존, 기능 회귀 없음 |
| 정리 | 세션 임시 파일 | 오케스트레이션 완료 후 `/tmp/autopus/{session-id}/` 자동 삭제 |

---

## 7. Technical Constraints

**기술 스택 제약**

- Go 1.26+, 외부 의존성 최소화 (stdlib 우선)
- 파일 감시: `os.Stat()` 폴링 우선 (fsnotify는 외부 의존성). 폴링 간격 200ms.
- Hook 스크립트: POSIX shell (bash/zsh) 호환. `jq` 의존성 없이 기본 shell 도구만 사용.
- 파일 크기 제한: 소스 파일 300줄 하드 리밋, 200줄 목표

**외부 의존성**

| 의존성 | 버전/SLA | 미가용 시 위험 |
|--------|---------|--------------|
| Claude Code CLI | >= 1.0 | Stop hook API 필요. 미설치 시 해당 프로바이더 skip |
| Gemini CLI | >= 2.0 | AfterAgent hook API 필요. 미설치 시 해당 프로바이더 skip |
| opencode (sst/anomalyco) | >= 0.1 | experimental.text.complete plugin 필요. 미설치 시 해당 프로바이더 skip |

**호환성 요구사항**

- macOS (darwin) 및 Linux 지원
- cmux 및 tmux 터미널 모두 지원
- 기존 `autopus.yaml` orchestra 설정과 하위 호환

**인프라 제약**

- 모든 데이터는 로컬 파일시스템 (`/tmp/autopus/`)에 저장
- 네트워크 통신 없음 (프로바이더 CLI가 자체적으로 API 호출)

---

## 8. Out of Scope

이 릴리즈에서 다루지 않는 항목:

- **원격 오케스트레이션**: 다른 머신의 프로바이더를 네트워크로 연결하는 기능
- **프로바이더 CLI 자동 설치/업데이트**: 바이너리 설치는 사용자 책임
- **OAuth/API 키 관리**: 인증 설정은 각 프로바이더 CLI가 자체 관리
- **Windows 지원**: POSIX shell hook 스크립트 기반이므로 Windows는 미지원
- **웹 UI 대시보드**: 결과 시각화는 CLI 출력으로만 제공
- **ReadScreen/PipePane 인프라 제거**: fallback 및 디버깅용으로 유지

**향후 반복으로 연기**

- Admin dashboard를 통한 오케스트레이션 모니터링 (UX 연구 후)
- 4개 이상 프로바이더 동시 오케스트레이션 (성능 검증 후)
- Hook 기반 양방향 통신 (slash command 주입 등)

---

## 9. Risks & Open Questions

### Risks

| 위험 | 심각도 | 확률 | 완화 전략 |
|------|--------|------|----------|
| 프로바이더 hook API 변경 | High | Low | Hook 스크립트를 템플릿화하여 버전별 분기 가능하게 설계. `auto doctor`에서 hook 호환성 검증 |
| opencode plugin API가 experimental | Medium | Medium | opencode 버전 핀 + fallback to ReadScreen. plugin API가 안정화될 때까지 실험적 태그 유지 |
| 동시 3 프로바이더의 `/tmp` 파일 충돌 | Medium | Low | 세션 ID 기반 디렉토리 격리로 충돌 방지. `randomHex()` 16자 세션 ID |
| Hook 스크립트 권한 문제 (실행 권한 없음) | Low | Medium | `auto init`에서 `chmod +x` 자동 적용. `auto doctor`에서 권한 검증 |
| 기존 ReadScreen 사용자의 행동 변경 | Low | Low | Hook 모드를 opt-in 기본값으로, `--legacy-screen` 플래그로 강제 fallback 제공 |

### Open Questions

| # | 질문 | 담당 | 기한 | 상태 |
|---|------|------|------|------|
| Q1 | opencode의 `experimental.text.complete` plugin API가 정식 출시 전에 변경될 가능성은? | planner | 2026-04-15 | Open |
| Q2 | Claude Code Stop hook에서 `last_assistant_message`의 정확한 JSON 스키마는? | executor | 2026-04-01 | Open |
| Q3 | Gemini CLI AfterAgent hook의 `prompt_response` 필드가 streaming 응답의 전체를 포함하는가, 마지막 chunk만인가? | executor | 2026-04-01 | Open |
| Q4 | Hook 모드를 기본값으로 할지, opt-in으로 할지? | planner | 2026-04-01 | Open — BS-002에서는 기본값 권장 |
| Q5 | `/tmp` 대신 XDG 캐시 디렉토리(`$XDG_CACHE_HOME/autopus/`)를 사용해야 하는가? | planner | 2026-04-15 | Open |

---

## 10. Practitioner Q&A

**Q1: Hook 스크립트의 정확한 실행 흐름은?**
A: 프로바이더 CLI가 응답 완료 시 hook을 자동 호출한다. Hook 스크립트는 stdin 또는 환경변수로 전달된 JSON에서 응답 필드를 추출하여 `/tmp/autopus/{session-id}/result.json`에 저장하고, `/tmp/autopus/{session-id}/done` 빈 파일을 생성한다. 세션 ID는 환경변수 `AUTOPUS_SESSION_ID`로 전달한다.

**Q2: `interactive.go`의 어느 부분이 변경되는가?**
A: `waitForCompletion()`의 ReadScreen 폴링 + idle 감지 이중 전략이 파일 감시(`done` 파일 존재 확인)로 대체된다. `collectResults()`의 `ReadScreen` + `cleanScreenOutput()`이 `result.json` 파싱으로 대체된다. `startPipeCapture()`와 `waitForSessionReady()`는 hook 모드에서 선택적이다.

**Q3: 기존 sentinel 모드(SPEC-ORCH-001)와의 관계는?**
A: sentinel 모드는 비대화식(`buildPaneCommand()`) 전용이다. Hook 모드는 인터랙티브 모드(SPEC-ORCH-006)의 결과 수집 방식을 대체한다. 두 모드는 독립적으로 공존하며, `OrchestraConfig`의 플래그로 분기한다.

**Q4: 프로바이더별 hook 설정 파일 위치는?**
A: Claude Code: `.claude/settings.json` (hooks.Stop), Gemini CLI: `.gemini/settings.json` (hooks.AfterAgent), opencode: `opencode.json` (experimental.text.complete plugin). 각 어댑터(`pkg/adapter/{platform}/`)에서 hook 주입 로직을 담당한다.

**Q5: `AUTOPUS_SESSION_ID` 환경변수는 어떻게 전달되는가?**
A: `launchInteractiveSessions()`에서 각 pane에 `export AUTOPUS_SESSION_ID={id}` 명령을 먼저 전송한 후 프로바이더 바이너리를 실행한다. 또는 프로바이더 실행 시 환경변수로 직접 설정한다.

**Q6: 3개 프로바이더가 동시에 같은 세션 디렉토리에 쓸 때 충돌은?**
A: 각 프로바이더는 자체 result.json과 done 파일을 `{provider}` 접두사로 구분한다: `/tmp/autopus/{session-id}/claude-result.json`, `/tmp/autopus/{session-id}/claude-done`. 파일명에 프로바이더명을 포함하여 동시 쓰기 충돌을 방지한다.

**Q7: Codex에서 opencode로의 마이그레이션 경로는?**
A: `pkg/config/migrate.go`에 마이그레이션 함수를 추가한다. `autopus.yaml`에서 `codex` 프로바이더를 `opencode`로 자동 변환하고, 바이너리 경로를 업데이트한다. 마이그레이션 시 경고 메시지를 출력하고, `auto doctor`에서 잔여 codex 설정을 감지한다.

**Q8: 실행 로드맵의 Phase 순서와 각 Phase의 산출물은?**
A:
- **Phase 1**: Hook 파일 시그널 구현 — hook 스크립트 3종, 파일 시그널 프로토콜, `interactive.go` 파일 감시 전환, hook 자동 주입
- **Phase 2**: Codex -> opencode 전환 — opencode 어댑터, config 마이그레이션, opencode plugin 스크립트
- **Phase 3**: 오케스트레이션 전략 고도화 — 4개 전략의 hook 결과 직접 활용, 구조화 합의/판정

