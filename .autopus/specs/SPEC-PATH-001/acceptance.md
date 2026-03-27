# SPEC-PATH-001 수락 기준

## 시나리오

### S1: 서브모듈 SPEC 자동 resolve
- Given: `autopus-adk/.autopus/specs/SPEC-ORCH-003/spec.md`가 존재한다
- When: `/auto go SPEC-ORCH-003`을 실행한다
- Then: resolution이 `autopus-adk/.autopus/specs/SPEC-ORCH-003/`을 찾고, TARGET_MODULE=`autopus-adk`, WORKING_DIR=`autopus-adk`로 설정된다

### S2: 최상단 레거시 SPEC resolve
- Given: `.autopus/specs/SPEC-ORCH-001/spec.md`가 최상단에 존재한다
- When: `/auto go SPEC-ORCH-001`을 실행한다
- Then: resolution이 `.autopus/specs/SPEC-ORCH-001/`을 찾고, TARGET_MODULE=`.`, WORKING_DIR=`.`로 설정된다

### S3: SPEC 미발견 에러
- Given: SPEC-NOEXIST-999가 어디에도 존재하지 않는다
- When: `/auto go SPEC-NOEXIST-999`를 실행한다
- Then: "SPEC-NOEXIST-999 not found" 에러 메시지가 표시된다

### S4: 중복 SPEC 에러
- Given: 동일 SPEC-ID가 `.autopus/specs/`와 `autopus-adk/.autopus/specs/` 모두에 존재한다
- When: 해당 SPEC-ID로 resolve를 시도한다
- Then: "Duplicate SPEC-{ID}" 에러와 함께 두 경로가 나열된다

### S5: plan에서 --target 플래그로 모듈 지정
- Given: 사용자가 `/auto plan "orchestra 기능" --target autopus-adk`를 실행한다
- When: spec-writer가 SPEC을 생성한다
- Then: `autopus-adk/.autopus/specs/SPEC-{DOMAIN}-{NUMBER}/`에 4개 파일이 생성된다

### S6: plan에서 자동 모듈 감지
- Given: 사용자가 `/auto plan "orchestra 프로바이더 추가"`를 실행한다 (--target 없음)
- When: 코드베이스 검색으로 orchestra 관련 코드가 autopus-adk에 집중되어 있음을 감지한다
- Then: autopus-adk를 target module로 선택하고 해당 경로에 SPEC을 생성한다

### S7: status에서 모듈별 그룹화
- Given: 최상단에 5개 SPEC, autopus-adk에 3개 SPEC이 존재한다
- When: `/auto status`를 실행한다
- Then: 모듈별로 그룹화된 대시보드가 표시된다 (예: "autopus-co (root): 5개", "autopus-adk: 3개")

### S8: go에서 WORKING_DIR이 executor에 전달됨
- Given: SPEC-ORCH-003이 autopus-adk에 resolve되었다
- When: Phase 2에서 executor가 스폰된다
- Then: executor의 프롬프트에 WORKING_DIR=autopus-adk가 포함되고, 빌드/테스트가 autopus-adk 내에서 실행된다

### S9: sync에서 TARGET_MODULE 기반 git 작업
- Given: SPEC-ORCH-003이 autopus-adk에 resolve되었다
- When: `/auto sync SPEC-ORCH-003`을 실행한다
- Then: git 작업이 `git -C autopus-adk` 또는 `cd autopus-adk && git` 형태로 서브모듈 내에서 수행된다

### S10: idea에서 BS 파일 모듈별 저장
- Given: 사용자가 `/auto idea "autopus-adk 관련 아이디어" --target autopus-adk`를 실행한다
- When: BS 파일을 저장한다
- Then: `autopus-adk/.autopus/brainstorms/BS-{ID}.md`에 저장된다
