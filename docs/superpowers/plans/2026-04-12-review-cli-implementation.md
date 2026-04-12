# review CLI - 实现计划

基于设计文档: `docs/superpowers/specs/2026-04-12-review-cli-design.md`

## 阶段 1: 项目初始化

**目标**: 创建 Go 项目结构，设置依赖

**任务**:
1. 初始化 Go module (`go mod init github.com/kbliu/review`)
2. 创建目录结构:
   ```
   cmd/review/
   internal/git/
   internal/diff/
   internal/ui/
   internal/highlight/
   ```
3. 添加依赖:
   - `github.com/charmbracelet/bubbletea`
   - `github.com/charmbracelet/bubbles`
   - `github.com/charmbracelet/lipgloss`
   - `github.com/alecthomas/chroma/v2`
   - `github.com/urfave/cli/v2`

**验收**: `go build ./cmd/review` 成功编译

---

## 阶段 2: Git 操作模块

**目标**: 实现 git diff 和文件列表获取

**任务**:
1. 创建 `internal/git/diff.go`
   - `GetDiffFiles(target string)` - 获取变更文件列表
   - `GetFileDiff(target, file string)` - 获取单个文件 diff
2. 创建 `internal/git/diff_test.go`
   - 使用 mock git 输出测试解析逻辑
3. 定义数据结构:
   - `FileInfo` (Status, Name)
   - `DiffResult` (Content)

**验收**: 单元测试通过，能从真实 git 仓库获取 diff

---

## 阶段 3: Diff 解析模块

**目标**: 解析 git diff 输出，提取行信息

**任务**:
1. 创建 `internal/diff/parser.go`
   - `ParseDiff(content string) []DiffLine`
   - 解析 hunk header (`@@ -old +new @@`)
   - 分类行类型 (Added/Removed/Context/HunkHeader)
   - 计算每行的旧/新行号
2. 创建 `internal/diff/parser_test.go`
   - 测试各种 diff 格式
   - 测试行号计算

**验收**: 单元测试覆盖所有行类型，行号计算正确

---

## 阶段 4: 代码高亮模块

**目标**: 使用 Chroma 高亮 diff 内容

**任务**:
1. 创建 `internal/highlight/chroma.go`
   - `HighlightDiff(lines []DiffLine, filename string) []HighlightedLine`
   - 根据文件扩展名选择 lexer
   - 对 context/added/removed 行分别高亮
2. 创建 `internal/highlight/chroma_test.go`
   - 测试常见语言高亮

**验收**: Go/Python/JavaScript 等文件能正确高亮

---

## 阶段 5: UI 组件

**目标**: 实现 Bubble Tea Model 和基础视图

**任务**:
1. 创建 `cmd/review/main.go`
   - CLI 参数解析 (target branch, --staged, -U)
   - 初始化 Bubble Tea program
2. 创建 `internal/ui/model.go`
   - 定义 Model 结构体
   - 实现 `tea.Model` 接口
3. 创建 `internal/ui/view.go`
   - 文件列表视图
   - diff 内容视图
   - 状态栏视图
4. 创建 `internal/ui/update.go`
   - 键盘事件处理
   - 窗口大小变化处理

**验收**: 能运行并显示界面，q 键可退出

---

## 阶段 6: 文件列表功能

**目标**: 实现文件列表显示和导航

**任务**:
1. 文件列表渲染 (状态 + 文件名)
2. j/k 导航
3. 当前选中项高亮
4. Enter 打开 diff

**验收**: 能在文件列表中上下导航，选中项高亮

---

## 阶段 7: Diff 显示功能

**目标**: 实现 diff 内容显示和滚动

**任务**:
1. 使用 `viewport` bubble 显示 diff
2. 显示行号 (旧行号 + 新行号)
3. 不同行类型不同颜色 (添加=绿，删除=红)
4. j/k 滚动 diff

**验收**: diff 正确显示，颜色区分，可滚动

---

## 阶段 8: 布局切换

**目标**: 支持左右/上下布局切换

**任务**:
1. 实现水平布局 (文件列表 | diff)
2. 实现垂直布局 (文件列表 / diff)
3. l 键切换布局
4. 响应式调整比例

**验收**: l 键可切换布局，内容正确重排

---

## 阶段 9: 外部编辑

**目标**: 支持用外部编辑器打开文件

**任务**:
1. e 键触发外部编辑
2. 获取 $EDITOR 环境变量
3. 调用编辑器并等待返回
4. 返回后刷新 diff

**验收**: 按 e 能用 vim/code 等编辑器打开文件

---

## 阶段 10: 刷新和错误处理

**目标**: 完善用户体验

**任务**:
1. r 键刷新 diff
2. 错误显示 (非 git 仓库、git 命令失败等)
3. 加载状态提示
4. 空状态处理 (无变更)

**验收**: 错误友好提示，刷新正常工作

---

## 阶段 11: 集成测试和优化

**目标**: 确保整体可用

**任务**:
1. 在真实项目测试
2. 性能优化 (大 diff 处理)
3. 修复边界情况
4. 添加 README 文档

**验收**: 在大型仓库（如 kubernetes）能正常使用

---

## 开发顺序

按阶段顺序开发，每个阶段完成后提交。
遵循 TDD：先写测试 -> 运行失败 -> 实现 -> 测试通过。
