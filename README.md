<div align="center">

# tmux-manager

**tmux 세션 관리를 위한 Dracula 테마 기반 TUI 도구**

`tm` 하나로 세션 생성·접속·삭제, YAML 기반 워크스페이스 복구까지.

</div>

---

## ✨ 기능

- **TUI 세션 관리** — `↑↓` 탐색, `Enter` 접속, `Ctrl+N/D/R` 조작
- **YAML 워크스페이스** — 세션·윈도우·pane 레이아웃을 파일로 정의
- **자동 복구** — 재부팅 후 `tm restore` 한 번이면 워크스페이스 복구
- **cfg 태그** — YAML로 관리되는 세션에 `cfg` 표시로 구분
- **Dracula 테마** — 개발자 친화적 컬러 스킴

---

## 📦 설치

```bash
git clone https://github.com/dohwi/tmux-manager.git
cd tmux-manager
go build -o tmux-manager .
./tmux-manager setup
```

`setup`이 자동으로 처리합니다:

| 항목 | 내용 |
|------|------|
| `~/.local/bin/tm` | 심볼릭 링크 생성 |
| `~/.zshrc` | `nocorrect tm` alias 등록 |
| `~/.profile` | `tm restore` 부팅 시 자동 실행 등록 |
| `~/.config/tmux-manager/sessions/` | 설정 디렉터리 생성 |

업데이트는 빌드만 하면 됩니다 — 심볼릭 링크가 자동으로 최신 바이너리를 가리킵니다.

```bash
go build -o tmux-manager .
```

---

## 🚀 사용법

```bash
tm          # TUI 실행
tm restore  # 설정파일 기반 세션 복구
```

### 키바인딩

| 키 | 동작 |
|:---|:-----|
| `↑` `↓` `j` `k` | 세션 선택 |
| `Enter` | 세션 접속 |
| `Ctrl+N` | 새 세션 생성 |
| `Ctrl+R` | 세션 이름 변경 |
| `Ctrl+D` | 세션 삭제 |
| `Ctrl+C` | 종료 |

---

## ⚙️ 설정

`~/.config/tmux-manager/sessions/*.yaml`

### 최소형 — 명령어 하나

```yaml
sessions:
  - name: monitoring
    command: htop
```

### 좌우 분할

```yaml
sessions:
  - name: dev
    directory: ~/projects/myapp
    panes:
      - command: nvim
      - command: lazygit
        direction: right
```

```
┌────────────┬──────────┐
│    nvim    │ lazygit  │
└────────────┴──────────┘
```

### 복합 레이아웃

```yaml
sessions:
  - name: dev
    directory: ~/projects/myapp
    panes:
      - command: nvim
      - command: lazygit
        direction: right
      - command: npm run dev
        direction: down
```

```
┌────────────┬──────────┐
│            │ lazygit  │
│    nvim    ├──────────┤
│            │ npm run  │
│            │   dev    │
└────────────┴──────────┘
```

### 멀티 윈도우

```yaml
sessions:
  - name: myapp
    directory: ~/projects/myapp
    windows:
      - name: code
        panes:
          - command: nvim
          - command: lazygit
            direction: right
      - name: infra
        panes:
          - command: docker-compose up
          - command: docker logs -f
            direction: down
```

### 한 파일에 여러 세션

```yaml
sessions:
  - name: dev
    directory: ~/projects/myapp
    panes:
      - command: opencode
      - command: lazygit
        direction: right

  - name: db
    directory: ~/projects/myapp
    command: psql myapp
```

---

## 📋 필드 참조

| 필드 | 필수 | 설명 |
|:-----|:----:|:-----|
| `sessions` | ✅ | 세션 정의 배열 |
| `sessions[].name` | ✅ | tmux 세션명 |
| `sessions[].directory` | | 시작 디렉터리 (`~/` 지원) |
| `sessions[].command` | | 첫 pane에서 실행할 명령어 |
| `sessions[].windows` | | 윈도우 목록 |
| `sessions[].windows[].name` | | 윈도우 탭 이름 (미지정 시 기본값) |
| `sessions[].windows[].panes` | | 윈도우 내 pane 구성 |
| `sessions[].panes` | | 기본 윈도우의 pane 구성 |
| `panes[].command` | | 해당 pane에서 실행할 명령어 |
| `panes[].direction` | | 분할 방향: `right` · `down` (첫 pane은 생략) |

---

## 🔄 자동 복구

`tm setup`이 `~/.profile`에 등록합니다:

```bash
tm restore 2>/dev/null
```

재부팅 후 SSH 접속 시 YAML에 정의된 세션이 자동으로 생성됩니다. 이미 존재하는 세션은 건너뜁니다.

---

## 🎨 cfg 태그

YAML로 관리되는 세션은 TUI에서 `cfg` 태그로 구분됩니다:

```
◉  good-giraffe-drawing    attached    cfg
○  myapp-dev               1 windows   cfg
●  temp-session             attached
```

---

<div align="center">

**tmux-manager** — manage sessions, not commands.

</div>