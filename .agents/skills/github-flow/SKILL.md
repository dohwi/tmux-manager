---
name: github-flow
description: 기능별 브랜치 생성 → 커밋 → squash merge 병합 워크플로우
metadata:
  workflow: github-flow
---

## GitHub Flow란

main 브랜치는 항상 배포 가능한 상태를 유지하고, 모든 변경은 기능별 브랜치에서 작업 후 PR로 병합하는 브랜치 전략이다.

## 규칙

- **main 직접 커밋 금지** — 반드시 기능 브랜치에서 작업
- **커밋은 목적별로 상세하게 분리** — 하나의 거대한 커밋 대신, 논리적 단위마다 개별 커밋
- **기능에 맞는 브랜치에만 커밋** — 현재 작업 중인 기능 브랜치에만 커밋, 다른 기능은 별도 브랜치에서 진행
- **squash merge만 사용** — `gh pr merge --squash --delete-branch`
- force push 금지, CI 실패 시 merge 금지

## 브랜치 네이밍

접두사(영어) + `/` + kebab-case: `feat/`, `fix/`, `hotfix/`, `refactor/` (예: `feat/add-session-restore`)

## 커밋 분리 원칙

- **한 커밋 = 하나의 논리적 변경** — 관심사 분리 원칙 적용
- **커밋 전 lint·typecheck·build·test 통과 필수**
- **기능별 브랜치 원칙** — 각 브랜치는 하나의 기능/수정만 담당, 여러 기능을 한 브랜치에서 섞지 않음

## 커밋 메시지

`<영어 접두사>: <한글 설명>` 형식 (예: `feat: 세션 복구 기능 추가`, `fix: 세션명 중복 처리 수정`)

## 워크플로우

1. **브랜치 생성**: `git checkout main && git pull origin main && git checkout -b <type>/<name>`
2. **개발 & 상세 분리 커밋**: 논리적 단위마다 개별 커밋, lint·typecheck·build·test 통과 후 커밋
3. **PR 생성**: `git push -u origin <branch> && gh pr create`
4. **리뷰 & 수정**: 피드백 반영, CI 통과 확인
5. **squash merge**: `gh pr merge --squash --delete-branch`
6. **배포**: main 머지 후 CI/CD 자동 배포 또는 수동 배포

## 핫픽스

`hotfix/<name>` 브랜치 → 최소 변경 커밋 → 즉시 PR → squash merge → 배포 후 회고

## 예시: 프로젝트 설정 기능 추가

```
# 1. main에서 기능 브랜치 생성
git checkout main && git pull origin main
git checkout -b feat/add-project-config

# 2. 개발 & 상세 분리 커밋 (각 논리적 단위별로 개별 커밋)
# 설정 구조체 추가
git add internal/config/config.go
git commit -m "feat: 프로젝트 설정 구조체 추가"

# TUI UI 구현
git add internal/tui/
git commit -m "feat: 설정 화면 UI 추가"

# 버그 수정
git commit -m "fix: 설정 파일 파싱 오류 수정"

# 헬퍼 유틸 추가
git add internal/util/
git commit -m "refactor: 경로 처리 유틸 분리"

# 3. PR 생성
git push -u origin feat/add-project-config
gh pr create --title "feat: 프로젝트별 설정 기능 추가" --body "YAML 설정 파일 기반 프로젝트 세션 관리"

# 4. 리뷰 & 수정 (피드백 반영 후 추가 커밋)
git commit -m "refactor: 설정 로드 로직 간소화"
git push

# 5. squash merge
gh pr merge --squash --delete-branch

# 6. 배포 (CI/CD 자동 또는 수동)
```