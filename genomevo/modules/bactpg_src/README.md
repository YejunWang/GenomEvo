# BactPG 2.0 (Optimized Single-Binary Version / 优化单文件版)

[English](#english) | [中文](#chinese)

---

<a id="english"></a>
## English

BactPG is a Bacterial Pan-Genome (PG) analysis pipeline. Previously, it consisted of a main execution script (`PG.go`) along with a dozen discrete auxiliary binaries placed in a `bin` directory. This optimized version entirely eliminates the need for separate sub-binaries by seamlessly integrating all auxiliary pipelines (such as sequence counting, map formatting, blast filtering, and fasta generation) into a **single, self-contained Go application**.

### Prerequisites

Before running BactPG, ensure the following command-line tools and software are installed and accessible in your system's `PATH`:
- **CD-HIT**: specifically the `cd-hit` command.
- **BLAST+**: specifically the `makeblastdb` and `blastp` commands.
- **Go**: Required if you wish to re-compile the source manually from `main.go`.

### Installation / Building

To compile the application simply navigate to the directory containing `main.go` and run:

```bash
go mod tidy
go build -o BactPG main.go
```

This will produce the standalone executable `BactPG`.

### Usage

```bash
./BactPG <SEQ_DIR> <SIMILARITY> <Number_of_threads> [--yes|-y]
```

#### Arguments
1. **SEQ_DIR**: The path to your input directory containing the genome datasets (e.g., protein `.fasta` files). This tool scans for `.fasta` files in the directory.
2. **SIMILARITY**: The threshold used by `cd-hit` and subsequent analysis steps (e.g., `0.7` for 70% identity).
3. **Number_of_threads**: Used by `cd-hit` and `blastp` to speed up alignments.
4. **`--yes` or `-y`** (optional): Non-interactive mode. If a previous `result/` folder exists, it will be automatically deleted without prompting. Essential for automated/web-based runs.

#### Example Argument

```bash
./BactPG test_data 0.7 30
./BactPG test_data 0.7 30 --yes   # non-interactive mode (for scripts/web)
```

> **Note:** The `result/` folder is created in the current working directory. When running via scripts or the web interface, always use the `--yes` flag to avoid interactive prompts. For manual interactive use, you will be asked whether to delete an existing `result/` folder.

### Workflow Details

BactPG operates in several systematic phases, automatically distributing computational loads over internal batches and iterations:
1. **CD-HIT clustering**: Eliminates redundant sequence files per bacterial strain and outputs independent maps linking specific genome sequences to clustered references.
2. **Batch Generation**: Strains are grouped randomly into distinct sub-batches (`batch1`, `batch2`, etc.).
3. **BLAST Search**: Reciprocal, all-to-all blast matching happens within batches and evaluates homologous proteins.
4. **Filtering and Parsing**: Highly optimized internal parsing (`blastFilterApp` internally overrides legacy `blastFilter`) checks similarity and combined structural coverage thresholds.
5. **Pan-genome Table Export**: Output tables (`PG.tmp1.txt` progressively updated across batches/iterations) map strain protein equivalencies.
6. **Iteration Consolidation**: Successive loops compile the respective output sequence sets into the ultimate cross-group pan-genome summary map.

### Output

All pipeline artifacts, intermediate database files, and the final results are generated in the `result/` folder:
- `cd_hit_out/` and `nr_seq/`: Sequences after local redundancy eliminations.
- `str_famap/`: Individual mappings of the representative sequences.
- `batch/` and `iter/`: Temporary blast databases, tabular matches, and unaligned FASTA data loops.
- `PG.txt`: The definitive consolidated pan-genome output matrix.

---

<a id="chinese"></a>
## 中文 (Chinese)

BactPG 是一个细菌泛基因组 (Pan-Genome, PG) 分析流程工具。在之前的版本中，它由一个主执行脚本（`PG.go`）和放置在 `bin` 目录中的十几个独立辅助二进制程序组成。这个优化版本通过将所有辅助流程（如序列计数、映射格式化、BLAST 过滤以及 FASTA 生成等）无缝集成到一个**独立的、完整的 Go 应用程序**中，完全消除了对单独子程序的依赖。

### 环境要求

在运行 BactPG 之前，请确保系统中已安装以下命令行工具和软件，并已加入到 `PATH` 环境变量中：
- **CD-HIT**：具体来说是 `cd-hit` 命令。
- **BLAST+**：具体来说是 `makeblastdb` 和 `blastp` 命令。
- **Go**：如果您想手动从 `main.go` 重新编译源码，则需要安装。

### 安装与构建

要编译应用程序，只需导航到包含 `main.go` 的目录并运行：

```bash
go mod tidy
go build -o BactPG main.go
```

这将会生成一个独立的可执行文件 `BactPG`。

### 使用方法

```bash
./BactPG <序列目录> <相似度阈值> <线程数> [--yes|-y]
```

#### 参数说明
1. **序列目录 (SEQ_DIR)**: 包含基因组数据集（例如 `.fasta` 蛋白质文件）的输入目录路径。本工具会扫描以 `.fasta` 结尾的文件。
2. **相似度阈值 (SIMILARITY)**: `cd-hit` 以及后续分析步骤使用的阈值（例如：`0.7` 代表 70% 的一致性）。
3. **线程数 (Number_of_threads)**: 用于加速 `cd-hit` 和 `blastp` 比对的线程数量。
4. **`--yes` 或 `-y`**（可选）：非交互模式。如果已存在 `result/` 文件夹，将自动删除而不提示。对于自动化脚本或Web界面运行至关重要。

#### 运行示例

```bash
./BactPG test_data 0.7 30
./BactPG test_data 0.7 30 --yes   # 非交互模式（适用于脚本/Web）
```

> **注意：** `result/` 文件夹会在当前工作目录中创建。通过脚本或Web界面运行时，请务必使用 `--yes` 标志以避免交互式提示。手动交互运行时，系统会询问是否删除已存在的 `result/` 文件夹。

### 工作流程细节

BactPG 在多个系统阶段中运行，自动分配计算负载并在内部划分批次和迭代：
1. **CD-HIT 聚类**: 消除每个细菌菌株的冗余序列，并输出将特定基因组序列与聚类参考序列相关联的独立映射。
2. **批次生成**: 菌株被随机分组到不同的子批次中（`batch1`, `batch2` 等）。
3. **BLAST 搜索**: 在批次内进行交互式（互为 Query 和 Subject）BLAST 匹配，评估同源蛋白质。
4. **过滤与解析**: 高度优化的内部解析器（`blastFilterApp` 内部替代了旧版的 `blastFilter`）用于检查相似度和结构覆盖率阈值。
5. **泛基因组表格导出**: 输出表格（在批次/迭代中逐步更新的 `PG.tmp1.txt`）用于映射菌株间的蛋白质同源等价关系。
6. **迭代合并**: 连续的循环迭代，逐步将各个输出的序列集编译成最终涵盖所有组别的泛基因组汇总表。

### 输出结果

所有流程产生的文件、中间数据库文件以及最终结果均生成在 `result/` 文件夹中：
- `cd_hit_out/` 与 `nr_seq/`: 去除局部冗余后的序列。
- `str_famap/`: 代表序列的个体映射关系。
- `batch/` 与 `iter/`: 临时的 BLAST 数据库、匹配表格及未比对的 FASTA 数据循环文件。
- `PG.txt`: 最终、确定性的综合泛基因组输出矩阵。
