# BactAG

**BactAG** is an integrated software package for bacterial ancestral genome inference and reconstruction. It completely rewrites the previous multi-process Python architectural scripts, Perl data-processing scripts, and `Ocode` Go components into a single statically-compiled Go binary. This significantly reduces environment dependency issues and increases parallel performance.

**BactAG** 是一个用于细菌祖先基因组推断和重建的集成软件。它将之前的多进程 Python 架构脚本、Perl 数据处理脚本以及 `Ocode` 相关的 Go 组件完全重写并打包成了一个单一的 Go 语言静态编译二进制可执行文件。这不仅大大减少了环境依赖问题，还显著提升了并行处理的性能。

---

## 🚀 Features | 特性

* **Single Binary (单一文件)**: No more configuring `multiprocessing` or resolving Perl/Python interpreter path issues. (不再需要配置多进程库或解决 Perl/Python 解释器路径问题)
* **High Performance (高性能并发)**: Utilizes Go's lightweight `goroutines` for task scheduling, which accelerates parsing and pipeline patching operations. (利用 Go 语言轻量级的协程进行任务调度，加速解析和管道修补操作)
* **Busybox-Style Tools (内建工具箱)**: Contains 12+ embedded CLI sub-tools out of the box (`homBB`, `progBackbonePrep`, `patching`, etc.), eliminating chaotic `$PATH` pollution. (开箱即用超过 12 个内置子工具，消除了对系统 `$PATH` 的混乱污染)

---

## 🛠️ Build & Installation | 编译与安装

If you have the Go compiler installed, you can build this software simply by:
如果您已经安装了 Go 编译器，只需通过以下命令即可编译本软件：

```bash
cd BactAG_Go
go mod tidy
go build -o BactAG ./cmd/bactag
```

After building, a single executable file named `BactAG` will be generated in the root directory.
编译完成后，根目录下会生成一个名为 `BactAG` 的独立可执行文件。

---

## 📖 How to Run | 运行方法

### Main Pipeline | 主流程

To run the main Ancestral Genome reconstruction pipeline, you can use the following command:
要运行主要的祖先基因组重建流程，您可以使用以下命令：

```bash
./BactAG -t 20 -tree input/tree_file -gene input/gene -id my_id.txt

```

**Parameters (参数解释):**
* `-t <int>`: Number of threads/goroutines to use. Default is `4`. (并发使用的线程数，默认是 4)
* `-tree <string>`: Path to the directory containing the phylogenetic tree file. Default is `input/tree_file`. (包含系统发育树文件的目录路径。默认是 `input/tree_file`)
* `-gene <string>`: Path to the directory containing the gene files. Default is `input/gene`. (包含基因文件的目录路径。默认是 `input/gene`)
* `-id <string>`: Path to the Strain ID file. Default is `codes/bactID.txt`. (菌株 ID 文件的路径。默认是 `codes/bactID.txt`)

*Note: Make sure your `input/gene/` and other required datasets exist in the current working directory before running.*
*注：在运行之前，请确保当前工作目录下存在 `input/gene/` 及其他所需的数据集。*

### Running Sub-components directly | 运行独立子组件

BactAG acts as a dynamic router (like busybox). You can invoke any internal legacy Go/Perl programs by passing the tool name as the first argument:
BactAG 充当了动态路由器的角色（类似 busybox）。您可以将工具名称作为第一个参数传递，以调用任何内部的传统程序：

```bash
./BactAG homBB arg1 arg2 > output.txt
./BactAG progBackbonePrep input.backbone > output.backbone.txt
```

**Available internal tools (可用的内置工具)：**
`homBB`, `homBlkReorder`, `orthBB`, `orthJoin1`, `orthJoin2`, `orthJoin3`, `orthoCombine`, `orthoParsing`, `patching`, `patching1`, `progBackbonePrep`, `revcomp`

---

## 📂 Directory Structure | 目录结构

* `cmd/bactag/`: Main entry point and CLI router. (主入口与 CLI 路由器)
* `internal/pipeline/`: Core orchestration logic, handles task trees and parallel scheduling. (核心编排逻辑，处理任务树和并行调度)
* `internal/tools/`: Parsers and manipulators replacing legacy Perl & Python scripts (`buchong`, `tiaozhen`, `seqext`). (替代传统 Perl/Python 脚本的解析和操作工具)
* `internal/ocode/`: Migrated Go components for backbone preparation and patching. (迁移而来的 Go 主干组装组件)
