# SPEC-ORCH-001 리서치

## 기존 코드 분석

### pkg/terminal/ — 터미널 멀티플렉서 어댑터 (SPEC-TERM-001에서 구현 완료)

**terminal.go** — `Terminal` 인터페이스 정의
- `Name() string`
- `CreateWorkspace(ctx, name) error`
- `SplitPane(ctx, direction) (PaneID, error)`
- `SendCommand(ctx, paneID, cmd) error`
- `Notify(ctx, message) error`
- `Close(ctx, name) error`

**cmux.go** — `CmuxAdapter` 구현
- `SplitPane()`: `cmux pane split --direction h|v` 실행, PaneID 반환
- `SendCommand()`: `cmux send-keys <paneID> <command>` 실행
- `Close()`: `cmux workspace remove <name>` 실행
- 입력 검증: `validateWorkspaceName()`, `validatePaneID()` 사용

**detect.go** — `DetectTerminal()` 함수
- 우선순위: cmux > tmux > plain
- `detect.IsInstalled()` 래핑하여 바이너리 존재 여부 확인
- 테스트에서 `isInstalled` 변수를 mock 가능

**plain.go** — `PlainAdapter` (no-op fallback)
- 모든 메서드가 no-op
- `CreateWorkspace()`에서 경고 로그 출력

### pkg/orchestra/ — 오케스트레이션 엔진

**types.go** — 핵심 타입 정의
- `OrchestraConfig`: Providers, Strategy, Prompt, TimeoutSeconds, JudgeProvider, DebateRounds
- `ProviderConfig`: Name, Binary, Args, PromptViaArgs
- `ProviderResponse`: Provider, Output, Error, Duration, ExitCode, TimedOut, EmptyOutput
- `FailedProvider`: Name, Error
- `OrchestraResult`: Strategy, Responses, Merged, Duration, Summary, FailedProviders

**runner.go** — 메인 실행 로직
- `RunOrchestra()`: 전략별 분기 (consensus, pipeline, debate, fastest)
- `runParallel()`: goroutine으로 병렬 실행, graceful degradation (빈 출력 → FailedProvider)
- `runProvider()`: 단일 프로바이더 실행, stdin/stdout 처리
  - `PromptViaArgs=true`: prompt를 args에 추가, stdin nil
  - `PromptViaArgs=false`: stdin pipe로 prompt 전송 후 닫기

**command.go** — exec 래퍼
- `command` 인터페이스: StdinPipe, SetStdin, SetStdout, SetStderr, Start, Wait, ExitCode
- `newCommand` 변수: 테스트에서 mock 가능

**debate.go** — debate 전략 실행
- Phase1: runParallel() → Phase2: runRebuttalRound() → Phase3: judge
- 각 phase에서 runProvider() 호출

### internal/cli/ — CLI 커맨드

**orchestra.go** — 서브커맨드 등록 및 `runOrchestraCommand()`
- `runOrchestraCommand()`: config 로딩 → 전략/프로바이더 resolve → `orchestra.RunOrchestra()` 호출
- `buildProviderConfigs()`: 하드코딩 기본값 (claude: -p, codex: -q, gemini: -p)

**orchestra_brainstorm.go** — brainstorm 서브커맨드
- `newOrchestraBrainstormCmd()`: SCAMPER/HMW 프롬프트 생성

**orchestra_config.go** — config resolve 로직
- `resolveStrategy()`, `resolveProviders()`, `resolveJudge()`: CLI 플래그 > 커맨드별 설정 > 글로벌 기본값

## 설계 결정

### 1. Terminal 인터페이스를 OrchestraConfig에 주입 (선택됨)

**이유**: pkg/orchestra 패키지가 pkg/terminal에 직접 의존하면 순환 참조나 결합도가 높아진다. 대신 CLI 레이어에서 DetectTerminal()을 호출하고 결과를 OrchestraConfig에 주입하면, orchestra 패키지는 Terminal 인터페이스만 알면 되고 테스트에서도 mock이 용이하다.

**대안 검토**:
- (A) orchestra 내부에서 DetectTerminal() 직접 호출 → 패키지 간 결합도 증가, 테스트 어려움
- (B) 전략 패턴으로 PaneRunner를 별도 전략으로 등록 → 과도한 추상화, 현재 전략 구조(consensus/pipeline/debate/fastest)와 직교하는 관심사

### 2. 임시 파일 기반 출력 캡처

**이유**: cmux send-keys로 인터랙티브 CLI를 실행하면 stdout을 직접 캡처할 수 없다. 각 프로바이더 명령을 `{cmd} | tee /tmp/output.txt; echo __DONE__ >> /tmp/output.txt` 형태로 실행하고, sentinel 문자열을 감시하여 완료를 감지한다.

**대안 검토**:
- (A) cmux capture-pane API → cmux에 해당 API가 없을 수 있음, 출력이 불완전할 수 있음
- (B) script 명령 사용 → 플랫폼 호환성 문제 (macOS vs Linux script 옵션 차이)
- (C) 파일 리디렉션 + sentinel → 가장 단순하고 신뢰성 있음 (선택됨)

### 3. 비인터랙티브 플래그 제거 로직

**이유**: 기존 ProviderConfig.Args에 `-p`, `-q` 등이 포함되어 있으므로, pane 모드에서는 이 플래그를 제거해야 한다. 알려진 비인터랙티브 플래그 목록을 관리하여 필터링한다.

**제거 대상 플래그**: `-p`, `--print`, `-q`, `--quiet`, `--non-interactive`

### 4. 기존 runProvider()와의 관계

pane 모드에서는 runProvider()를 사용하지 않는다. 대신 pane_runner가 Terminal.SendCommand()로 명령을 전송하고 파일에서 결과를 읽어 ProviderResponse를 직접 구성한다. 이렇게 하면 runProvider()의 stdin/stdout 파이프 로직과 충돌하지 않는다.

### 5. debate 전략에서의 pane 재활용

debate Phase1과 Phase2에서 같은 프로바이더가 재실행된다. Phase2에서는 기존 pane을 재활용하여 새 프롬프트를 전송한다(pane을 다시 생성하지 않음). judge는 단일 실행이므로 기존 비인터랙티브 모드로 충분하다(별도 pane 불필요).

## 리스크

1. **cmux send-keys 타이밍**: 프롬프트가 길면 send-keys가 잘리거나 지연될 수 있다 → 프롬프트를 파일에 쓰고 `cat prompt.txt | {binary}` 형태로 전송
2. **출력 파싱**: 인터랙티브 모드 출력에 ANSI escape code, 프롬프트 텍스트 등이 포함될 수 있다 → ANSI strip 필터 적용
3. **프로바이더별 완료 감지**: 인터랙티브 모드에서 명시적 종료 시그널이 없을 수 있다 → sentinel + 프로세스 종료 감시 병행
