# JSON Diff

一个高性能、递归的 JSON 对比工具，支持多种输出格式（控制台彩色文本、HTML 报告）。

## 🌟 核心特性

- **递归对比**: 深入对比嵌套的 JSON 结构（对象、数组、基本类型）。
- **智能识别**: 自动识别 `新增`、`删除`、`修改` 和 `类型变更`。
- **多种输出模式**:
  - **控制台**: ANSI 彩色高亮显示差异。
  - **HTML 报告**: 生成交互式的 HTML 可视化报告，支持深色模式。
- **差异过滤**: 支持 `-diff-only` 模式，只关注发生变化的部分。
- **零外部依赖**: 仅使用 Go 标准库实现（`encoding/json`, `html/template` 等）。

## 🚀 安装方法

确保您的系统中已安装 Go (建议 1.25+)。

```bash
# 克隆仓库或下载源码
git clone <repository-url>
cd json-diff

# 构建可执行文件
go build -o json-diff
```

## 📖 使用指南

### 基本用法

对比两个 JSON 文件并直接在终端输出结果：

```bash
./json-diff -file1 1.json -file2 2.json
```

### 仅显示差异

如果您只想看到变化的内容，可以启用 `diff-only` 标志：

```bash
./json-diff -file1 1.json -file2 2.json -diff-only
```

### 生成 HTML 报告

生成一个美观的、可交互的 HTML 报告：

```bash
./json-diff -file1 1.json -file2 2.json -html report.html
```

### 保存文本报告

将带颜色的文本报告保存到文件中：

```bash
./json-diff -file1 1.json -file2 2.json -out diff.txt
```

## 🛠️ 命令行参数说明

| 参数 | 说明 |
| :--- | :--- |
| `-file1` | 第一个 JSON 文件的路径（旧值） |
| `-file2` | 第二个 JSON 文件的路径（新值） |
| `-diff-only` | 布尔值，设为 true 时仅显示发生变化的节点 |
| `-html` | 指定生成的 HTML 报告的文件路径 |
| `-out` | 指定生成的文本报告的文件路径 |

## 📊 报告效果展示

### HTML 报告
HTML 报告提供了一个现代化的 UI，包含：
- **深色主题**: 减少视觉疲劳。
- **状态切换**: 通过页面上的开关随时切换“全量显示”或“仅显示差异”。
- **语法高亮**: 对不同类型的值（字符串、数字、布尔值等）进行颜色区分。

### 控制台输出
控制台使用 ANSI 转义序列：
- <span style="color:green">+ 绿色</span> 表示新增
- <span style="color:red">- 红色</span> 表示删除或修改前的值
- <span style="color:cyan">青色</span> 表示键名

---

Made with ❤️ using Go
