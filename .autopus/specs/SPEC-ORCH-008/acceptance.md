# SPEC-ORCH-008 수락 기준

---

## P0 시나리오

### S1: 3라운드 Interactive Debate 정상 실행 (R1)

- **Given**: claude, gemini, opencode 3개 프로바이더가 설정되고, interactive + hook 모드가 활성화된 상태
- **When**: `auto orchestra --multi --strategy debate --rounds 3` 실행
- **Then**: Round 1(독립 응답) → Round 2(교차 반박) → Round 3(재반박) 순서로 3라운드가 실행되고, 최종 결과가 출력된다

### S2: 라운드 스코프 Hook 시그널 분리 (R2)

- **Given**: Round 1이 진행 중인 상태
- **When**: claude가 응답을 완료하면
- **Then**: `/tmp/autopus/{session-id}/claude-round1-result.json`과 `claude-round1-done`이 생성된다 (기존 `claude-done`이 아닌)

### S3: 라운드 간 시그널 충돌 방지 (R2, R6)

- **Given**: Round 1이 완료되어 `claude-round1-done`이 존재하는 상태
- **When**: Round 2가 시작되면
- **Then**: `claude-round1-done`이 삭제되고, `AUTOPUS_ROUND=2`가 설정되며, Round 1의 done 시그널이 Round 2 완료를 오판하지 않는다

### S4: Rebuttal 프롬프트 Interactive 주입 (R3)

- **Given**: Round 1이 완료되어 각 프로바이더의 응답이 수집된 상태
- **When**: Round 2가 시작되면
- **Then**: 각 pane에 `## Other debaters' arguments:` 섹션이 포함된 rebuttal 프롬프트가 전송되고, 다른 프로바이더들의 Round 1 응답이 포함된다

### S5: Pane 입력 대기 확인 후 프롬프트 전송 (R3)

- **Given**: Round 1의 응답이 아직 진행 중인 프로바이더 pane이 있는 상태
- **When**: Round 2 프롬프트 전송을 시도하면
- **Then**: `pollUntilPrompt()`로 pane이 입력 대기 상태가 될 때까지 기다린 후 프롬프트를 전송한다

### S6: --rounds 플래그 유효성 검증 (R4)

- **Given**: CLI에서 `--strategy consensus --rounds 3`을 지정한 상태
- **When**: 실행을 시도하면
- **Then**: `--rounds requires --strategy debate` 오류 메시지를 출력하고 종료한다

### S7: --rounds 범위 검증 (R4)

- **Given**: CLI에서 `--strategy debate --rounds 11`을 지정한 상태
- **When**: 검증 시
- **Then**: 유효 범위(1-10) 오류를 출력하고 종료한다

### S8: --rounds 미지정 시 기본값 (R4)

- **Given**: `--strategy debate`만 지정하고 `--rounds`를 지정하지 않은 상태
- **When**: debate 실행 시
- **Then**: 기본값 2(Round 1 독립 응답 + Round 2 rebuttal)로 동작한다

### S9: Hook 스크립트 라운드 인식 (R5)

- **Given**: `AUTOPUS_ROUND=2` 환경변수가 설정된 상태
- **When**: claude hook 스크립트가 실행되면
- **Then**: `claude-round2-result.json`과 `claude-round2-done` 파일이 생성된다

### S10: Hook 스크립트 하위 호환 (R5)

- **Given**: `AUTOPUS_ROUND` 환경변수가 미설정인 상태 (단일 실행 또는 기존 흐름)
- **When**: claude hook 스크립트가 실행되면
- **Then**: 기존 `claude-result.json`과 `claude-done` 파일이 생성된다

### S11: 시그널 정리 및 결과 보존 (R6)

- **Given**: Round 1이 완료되어 result와 done 파일이 모두 존재하는 상태
- **When**: `CleanRoundSignals(round=1)` 실행 시
- **Then**: `{provider}-round1-done`은 삭제되고, `{provider}-round1-result.json`은 보존된다

### S12: Interactive Debate 결과 병합 (R7)

- **Given**: 3라운드 debate가 모든 라운드를 완료한 상태
- **When**: 결과를 병합하면
- **Then**: 최종 라운드(Round 3)의 응답으로 `mergeByStrategy(StrategyDebate, ...)`가 호출되고, `OrchestraResult.RoundHistory`에 3개 라운드의 히스토리가 저장된다

---

## P1 시나리오

### S13: Judge 라운드 Interactive 실행 (R8)

- **Given**: `--judge claude`로 judge가 지정되고 3라운드 debate가 완료된 상태
- **When**: judge 라운드가 실행되면
- **Then**: judge pane에 전체 토론 기록이 포함된 판정 프롬프트가 전송되고, 결과가 `(judge)` 태그와 함께 최종 결과에 포함된다

### S14: 라운드 진행 상태 표시 (R9)

- **Given**: 3라운드 debate가 진행 중인 상태
- **When**: 각 라운드가 시작/완료되면
- **Then**: `[Round 1/3] 시작...`, `[Round 1/3] claude 완료 (12.3s)`, `[Debate 완료] 3라운드, 45.2s` 형태의 메시지가 stdout에 출력된다

### S15: 조기 합의 감지 (R10)

- **Given**: Round 2에서 모든 프로바이더의 응답이 66% 이상 동일한 상태
- **When**: 합의 체크가 실행되면
- **Then**: `[Early Consensus] Round 2에서 합의 도달, 남은 1라운드 건너뜀` 메시지 출력 후 Round 3를 skip한다

### S16: RoundHistory 구조 (R11)

- **Given**: 3라운드 debate가 완료된 상태
- **When**: `OrchestraResult`를 확인하면
- **Then**: `RoundHistory[0]`에 Round 1 응답, `RoundHistory[1]`에 Round 2 응답, `RoundHistory[2]`에 Round 3 응답이 저장되고, `Responses`는 최종 라운드 응답과 동일하다

---

## P2 시나리오

### S17: 라운드별 비교 뷰 (R12)

- **Given**: 멀티라운드 debate가 완료된 상태
- **When**: 비교 뷰가 출력되면
- **Then**: 각 프로바이더의 라운드별 응답 변화가 요약되어 표시된다

### S18: 라운드별 개별 타임아웃 (R13)

- **Given**: `PerRoundTimeout=30s`가 설정된 3라운드 debate 상태
- **When**: Round 2에서 특정 프로바이더가 30초를 초과하면
- **Then**: 해당 프로바이더만 타임아웃 처리되고, 부분 결과로 라운드가 완료된다

### S19: 토론 기록 저장 (R14)

- **Given**: 3라운드 debate가 완료된 상태
- **When**: 결과가 저장되면
- **Then**: `.autopus/debate-history/{session-id}.json`에 전체 라운드별 프롬프트와 응답이 JSON으로 저장된다

---

## 비기능 요구사항 검증

### S20: 라운드 간 전환 성능

- **Given**: 라운드 N이 완료된 직후
- **When**: 시그널 정리 + env 업데이트 + 프롬프트 주입이 실행되면
- **Then**: 전환 시간이 3초 이내이다

### S21: 하위 호환 (rounds=1)

- **Given**: `--strategy debate --rounds 1` 또는 `--rounds` 미지정 + rounds=1 상태
- **When**: debate가 실행되면
- **Then**: SPEC-ORCH-007의 단일 수집 흐름과 동일하게 동작하고, 라운드 스코프 시그널 없이 기존 `{provider}-done` 형식을 사용한다

### S22: 프로바이더 1개 실패 시 나머지 계속

- **Given**: 3프로바이더 중 1개가 Round 2에서 타임아웃된 상태
- **When**: 나머지 2개 프로바이더의 결과를 수집하면
- **Then**: 타임아웃된 프로바이더만 skip되고, 나머지 2개의 결과로 라운드가 정상 완료된다
