# review - Go Code Review CLI 设计文档

## 概述

将 vim-code-review 插件移植为独立的 Go CLI 工具，使用 Bubble Tea 框架实现 TUI 界面，专门用于命令行环境下的 code review。

## 目标

- 开源发布，供其他开发者使用
- 保持与 vim 插件一致的布局和操作体验
- 零配置，开箱即用
- 支持代码高亮

## 技术栈

- **语言**: Go 1.21+
- **TUI 框架**: Bubble Tea (Charm 生态)
- **代码高亮**: Chroma
- **CLI 解析**: urfave/cli/v2

## 功能规格

### 核心功能

| 功能 | 描述 |
|------|------|
| 本地 Git Diff | 比较工作区与 HEAD，或比较任意分支 |
| 分栏布局 | 左侧文件列表，右侧 diff 内容 |
| 布局切换 | 支持左右布局和上下布局切换 |
| 代码高亮 | 使用 Chroma 实现多语言语法高亮 |
| 行号显示 | 显示旧/新文件行号（与 vim 插件一致） |
| 外部编辑 | 按 `e` 使用 `$EDITOR` 打开当前文件 |

### 键位设计

| 按键 | 功能 |
|------|------|
| `j` / `↓` | 下一个文件（列表焦点）/ 向下滚动（diff 焦点） |
| `k` / `↑` | 上一个文件（列表焦点）/ 向上滚动（diff 焦点） |
| `h` / `←` | 聚焦文件列表 |
| `l` / `→` | 聚焦 diff 视图 / 切换布局 |
| `Enter` | 查看选中文件的 diff |
| `e` | 用 `$EDITOR` 打开当前文件 |
| `r` | 刷新（重新执行 git diff） |
| `Tab` | 在文件列表和 diff 视图间切换焦点 |
| `q` / `Ctrl+C` | 退出 |

### 命令行接口

```bash
review              # 查看本地变更（vs HEAD）
review main         # 对比指定分支
review --staged     # 查看暂存区变更
review -U 10 main   # 指定上下文行数
```

## 界面布局

### 左右布局（默认）

```
┌─────────────────┬─────────────────────────────────────┐
│ M  src/main.go  │  1   1   package main               │
│ A  src/util.go  │  2   2                                │
│    src/helper/  │  3      -import "fmt"               │
│ D  README.md    │  4      +import (                     │
│                 │  5   3       "fmt"                    │
│                 │  6   4       "strings"                │
│                 │  7   5   )                            │
│                 │                                     │
├─────────────────┴─────────────────────────────────────┤
│ main > HEAD | Files: 4 | Current: 1/4 | [h]elp [q]uit │
└───────────────────────────────────────────────────────┘
```

### 上下布局（按 `l` 切换）

```
┌───────────────────────────────────────────────────────┐
│ M  src/main.go                                        │
│ A  src/util.go                                        │
│ D  README.md                                          │
├───────────────────────────────────────────────────────┤
│  1   1   package main                                 │
│  3      -import "fmt"                                 │
│  4      +import (                                     │
├───────────────────────────────────────────────────────┤
│ main > HEAD | Files: 3 | 1/3 | [h]elp [l]ayout [q]    │
└───────────────────────────────────────────────────────┘
```

## 架构设计

```
┌─────────────────────────────────────┐
│           main.go                   │  CLI 入口，参数解析
│         (urfave/cli)                │
├─────────────────────────────────────┤
│         cmd/review/                 │
│  ┌─────────────────────────────┐    │
│  │         model.go            │    │  Bubble Tea Model
│  │  - Model (state)            │    │  - Update (msg handler)
│  │  - View (render)            │    │  - Init (tea.Cmd)
│  └─────────────────────────────┘    │
├─────────────────────────────────────┤
│        internal/                    │
│  ┌──────────┐ ┌──────────┐ ┌─────┐ │
│  │   git    │ │  diff    │ │ ui  │ │
│  │  diff    │ │  parser  │ │     │ │
│  │  stats   │ │  hunk    │ │     │ │
│  └──────────┘ └──────────┘ └─────┘ │
│  ┌──────────┐ ┌──────────┐         │
│  │  chroma  │ │ config   │         │
│  │ highlighter│  (env)   │         │
│  └──────────┘ └──────────┘         │
└─────────────────────────────────────┘
```

## 数据结构

### Model 状态

```go
type Model struct {
    // 布局
    layout LayoutType // LayoutHorizontal | LayoutVertical

    // 文件列表
    files []FileInfo
    cursor int

    // Diff 内容
    currentFile string
    diffContent string
    diffLines []DiffLine

    // 焦点
    focus FocusType // FocusList | FocusDiff

    // 滚动
    diffViewport viewport.Model

    // 状态
    err error
    loading bool
}
```

### Diff 行结构

```go
type DiffLine struct {
    Type      LineType    // Added | Removed | Context | HunkHeader
    OldLineNo int         // 0 表示新增行
    NewLineNo int         // 0 表示删除行
    Content   string
    Highlight []token     // Chroma 高亮结果
}
```

## 开发原则

1. **TDD 开发**: 每个功能先写测试，再写实现
2. **小步提交**: 每个可工作的功能点单独提交
3. **保持简单**: 不添加未指定的功能
4. **测试覆盖**: 核心逻辑必须有单元测试

## 验收标准

- [ ] 可以编译为单个二进制文件
- [ ] 支持 `review` 查看本地变更
- [ ] 支持 `review <branch>` 对比分支
- [ ] 文件列表可以上下导航
- [ ] 选中文件显示 diff 内容
- [ ] diff 内容支持语法高亮
- [ ] diff 显示旧/新行号
- [ ] 支持布局切换（左右/上下）
- [ ] 支持外部编辑器打开文件
- [ ] 所有键位功能正常工作
