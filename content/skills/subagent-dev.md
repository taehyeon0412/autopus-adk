---
name: subagent-dev
description: 서브에이전트 개발 및 오케스트레이션 스킬
triggers:
  - subagent
  - 서브에이전트
  - agent development
  - 에이전트 개발
  - orchestration
category: agentic
level1_metadata: "서브에이전트 설계, 오케스트레이션 패턴, 병렬 실행"
---

# Subagent Development Skill

효과적인 서브에이전트를 설계하고 오케스트레이션하는 스킬입니다.

## 서브에이전트 설계 원칙

### 단일 책임 원칙
각 에이전트는 하나의 명확한 역할을 가집니다:
- ✅ `expert-backend`: 백엔드 API 구현만 담당
- ✅ `manager-tdd`: TDD 워크플로우 조율만 담당
- ❌ 여러 역할을 하나의 에이전트에 혼재

### 격리 원칙
에이전트는 독립적으로 실행됩니다:
- 이전 대화 히스토리에 접근 불가
- 필요한 모든 컨텍스트를 spawn prompt에 포함
- 결과는 구조화된 형식으로 반환

### 병렬 실행 원칙
독립적인 작업은 병렬로 실행합니다:
```
독립적 → 병렬 실행 (단일 메시지에 여러 Agent() 호출)
의존적 → 순차 실행 (이전 결과를 다음 입력으로)
```

## 오케스트레이션 패턴

### Fan-Out / Fan-In
```
조율자 → [에이전트 A, 에이전트 B, 에이전트 C] (병렬)
         → 결과 통합 → 조율자
```

### Pipeline
```
에이전트 A → 결과 → 에이전트 B → 결과 → 에이전트 C
```

### Supervisor
```
감독자 에이전트 → 실행자 에이전트 모니터링
               → 실패 시 재시도 또는 대안 전략
```

## 서브에이전트 정의 형식

```markdown
---
name: expert-[domain]
description: [한 줄 역할 설명]
tools:
  - Read, Write, Edit, Bash, Glob, Grep
  - TaskCreate, TaskUpdate, TaskList
---

역할 지침:
1. 무엇을 해야 하는가
2. 어떤 파일을 수정할 수 있는가
3. 완료 기준은 무엇인가
4. 결과를 어떻게 보고하는가
```

## 완료 기준

- [ ] 에이전트 정의에 명확한 역할 설명
- [ ] 입출력 계약 정의
- [ ] 병렬 실행 기회 파악
- [ ] 에러 처리 전략 포함
- [ ] 최대 3회 재시도 후 사용자 개입 요청
