# Review: SPEC-HARN-DOCFETCH-001

**Verdict**: REVISE
**Revision**: 0
**Date**: 2026-04-05 09:59:48

## Findings

| Provider | Severity | Description |
|----------|----------|-------------|
| gemini | critical | 파이프라인 엔진 모델과 Phase 1.8 불일치 (REQ-DF-020) |
| gemini | major | pkg/setup 스캐너의 의존성 전체 추출 기능 부재 (REQ-DF-010, REQ-DF-011) |
| gemini | major | 정확한 토큰 예산 산정을 위한 토크나이저 누락 (REQ-DF-006) |
| gemini | minor | HTML 웹 스크래핑의 취약성과 대안 부재 (REQ-DF-002) |
| gemini | minor | Claude Code MCP 환경 감지 메커니즘 불명확 (REQ-DF-021) |

## Provider Responses

### Response 1

MCP issues detected. Run /mcp list for status.1. VERDICT: REVISE

2. FINDINGS:
   - FINDING: [critical] 파이프라인 엔진 모델과 Phase 1.8 불일치 (REQ-DF-020)
     현재 `pkg/pipeline/phase.go`의 `DefaultPhases()`는 `plan`, `test_scaffold`, `implement`, `validate`, `review` 5단계만 관리하며 실행합니다. `Phase 1.8 Doc Fetch`는 프롬프트 템플릿 마크다운 내부에 지시사항으로만 존재할 뿐, Go 파이프라인 엔진이 인식하는 독립적인 Phase가 아닙니다. 엔진 내에서 어떻게 `auto docs fetch`를 가로채어 실행할지에 대한 아키텍처 및 `pkg/pipeline/*` 파일 수정 계획이 스펙과 생성 파일 목록에서 완전히 누락되었습니다.
   - FINDING: [major] pkg/setup 스캐너의 의존성 전체 추출 기능 부재 (REQ-DF-010, REQ-DF-011)
     SPEC은 `pkg/setup` 스캐너의 논리를 재사용하여 의존성을 감지할 것을 요구하지만, 현재 `pkg/setup/scanner.go`는 프로젝트의 메인 언어와 일부 프론트엔드 프레임워크(React, Vue 등)만 식별할 뿐 `go.mod`나 `package.json` 등의 "전체 패키지/라이브러리 목록"을 추출하고 반환하는 로직이 없습니다. 단순 재사용으로는 불가능하며 스캐너의 대대적인 확장이 스펙에 명시되어야 합니다.
   - FINDING: [major] 정확한 토큰 예산 산정을 위한 토크나이저 누락 (REQ-DF-006)
     "1 lib ~5000 tokens"와 같이 엄격한 토큰 예산을 강제하고 있으나, 현재 프로젝트(`pkg/cost` 등)에는 문자열로부터 실제 토큰 수를 계산할 수 있는 라이브러리(tiktoken 등)가 구현되어 있지 않습니다. 신규 라이브러리 추가 계획인지, 혹은 단순 글자 수 비례(문자 수 / 4) 근사치 알고리즘을 사용할 것인지에 대한 정의가 누락되어 구현 타당성이 떨어집니다.
   - FINDING: [minor] HTML 웹 스크래핑의 취약성과 대안 부재 (REQ-DF-002)
     `npmjs.com`, `pypi.org` 등 공식 문서 웹페이지를 직접 스크래핑하는 것은 Cloudflare 봇 방어 및 DOM 변경에 의해 CLI 환경에서 쉽게 붕괴될 수 있습니다. HTML 스크래핑보다는 공식 레지스트리의 JSON API(`registry.npmjs.org`)나 로컬 명령어(`go doc`) 등을 활용하는 보다 안정적인 방식이 권장됩니다.
   - FINDING: [minor] Claude Code MCP 환경 감지 메커니즘 불명확 (REQ-DF-021)
     "Claude Code 환경에서 MCP가 사용 가능한 경우"를 시스템이 감지하도록 요구하지만, 서브프로세스로 AI CLI를 호출하는 오케스트레이터 입장에서 이 상태를 어떻게 판단할지(`claude-code-manifest.json` 파일 존재 유무 등) 구체적인 감지 기준이 명시되어 있지 않습니다.

3. REASONING:
이 SPEC은 기능 추가의 의도와 요구사항은 명확하지만, 현재 `autopus-adk`의 파이프라인 엔진 모델 및 스캐너 코드베이스의 실제 한계를 간과하고 작성되었습니다. 특히 파이프라인 엔진(`pkg/pipeline/engine.go`) 내에는 존재하지 않는 "Phase 1.8"을 내부적으로 어떻게 스케줄링하고 주입할 것인지에 대한 아키텍처 설계가 누락된 점은 치명적입니다. 또한, 없는 기능(`pkg/setup`의 전체 패키지 파싱)을 재사용하라는 지시와 토큰 예산/웹 스크래핑 구현의 기술적 한계 등 Feasibility(타당성) 이슈가 다수 존재하므로 코드베이스 현실에 맞게 전체적으로 문서를 수정(REVISE)해야 합니다.


