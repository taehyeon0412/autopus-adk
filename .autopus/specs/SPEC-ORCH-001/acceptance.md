# SPEC-ORCH-001 수락 기준

## 시나리오

### S1: cmux 감지 시 pane 분할 모드 진입
- Given: cmux가 설치되어 있고 DetectTerminal()이 CmuxAdapter를 반환
- When: `auto orchestra brainstorm "test" --strategy debate` 실행
- Then: Terminal.SplitPane()이 프로바이더 수만큼 호출되고, 각 pane에 인터랙티브 CLI 명령이 전송된다

### S2: cmux 미감지 시 기존 모드 fallback
- Given: cmux/tmux가 설치되지 않아 PlainAdapter가 반환
- When: `auto orchestra brainstorm "test"` 실행
- Then: 기존 `-p` 비인터랙티브 모드로 실행되며, pane 분할이 발생하지 않는다

### S3: 프로바이더 결과 수집 및 merge
- Given: cmux pane 모드에서 claude, codex, gemini 3개 프로바이더 실행
- When: 모든 프로바이더가 정상 완료
- Then: 각 프로바이더의 출력이 ProviderResponse로 변환되고, 기존 merge 로직(consensus, debate 등)으로 통합 결과가 생성된다

### S4: pane 정리
- Given: orchestra 실행이 완료 (성공 또는 실패)
- When: 정리 단계 진입
- Then: 생성된 모든 pane이 닫히고, 임시 출력 파일이 삭제된다

### S5: pane 생성 실패 시 fallback
- Given: cmux가 감지되었으나 SplitPane()이 에러 반환
- When: pane 생성 실패
- Then: 에러를 로깅하고 기존 비인터랙티브 모드로 자동 fallback한다

### S6: 프로바이더 타임아웃 처리
- Given: cmux pane 모드에서 프로바이더 실행 중
- When: 프로바이더가 설정된 타임아웃 내에 완료되지 않음
- Then: 해당 pane이 강제 종료되고, FailedProvider로 기록되며, 나머지 프로바이더 결과로 merge가 진행된다

### S7: OrchestraConfig.Terminal이 nil일 때 기존 동작 유지
- Given: OrchestraConfig.Terminal이 nil (CLI에서 주입하지 않은 경우)
- When: RunOrchestra() 호출
- Then: 기존 runParallel() 또는 전략별 실행 함수가 호출되어 동작이 변경되지 않는다

### S8: 인터랙티브 Args 변환
- Given: ProviderConfig.Args가 ["-p"]인 claude 프로바이더
- When: pane 모드로 실행
- Then: "-p" 플래그가 제거되어 인터랙티브 모드로 실행된다

### S9: debate 전략에서 pane 모드 동작
- Given: cmux 감지, strategy=debate, 3개 프로바이더 + judge 설정
- When: `auto orchestra brainstorm "test" --strategy debate --judge gemini` 실행
- Then: Phase1(initial)에서 3개 pane 생성, Phase2(rebuttal)에서 pane 재활용 또는 재생성, Phase3(judge)에서 별도 pane 또는 기존 모드로 judge 실행
