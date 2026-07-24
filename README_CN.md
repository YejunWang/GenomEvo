# 🧬 GenomEvo 
**细菌比较与演化基因组学自动化分析集成系统**

[![Python](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/)
[![Go](https://img.shields.io/badge/go-1.18+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

[English](README.md) | **中文**

---

## 📋 概述

**GenomEvo** 是一个以祖先直系同源基因组（AOG）为核心的自动化细菌比较与演化基因组学分析管线。它将七个核心模块整合为一个统一的、易于使用的平台，同时提供**命令行界面（CLI）**和**基于Web的图形用户界面（GUI）**。

### 核心模块一览

| 模块 | 功能 | 输入 | 输出 |
|--------|----------|-------|--------|
| **BactCG** | 核心基因组鉴定 + 系统发育树推断 | 蛋白质FASTA文件（`.fasta`/`.faa`/`.fa`） | 核心基因簇 + MEG比对 + SNP矩阵 + 系统发育树 |
| **BactAG** | 祖先基因组重建 | 系统发育树（Newick格式）+ 基因组FASTA文件 | 祖先基因组（AOG）+ `bactID.txt` |
| **BactPG** | 泛基因组分析（单文件统一二进制） | 序列FASTA文件目录 | `PG.txt` 泛基因组存在/缺失矩阵 |
| **BactPGA** | 泛基因组注释（整合BactAG + BactPG） | GenBank文件 + PG矩阵 + 相互最佳比对 | 含PG簇UID注释的基因表格 |
| **Bact1DGR** | 一维基因组表示 | AOG + 补丁菌株 + `bactID.txt` | `*.1dgr.txt` 片段坐标图谱 |
| **BactEvolTraj** | 演化轨迹分析 | FASTA + GenBank + 进化树边（三向Mauve） | SV事件表 + 覆盖度矩阵 + 图表 |
| **BactFragAnn** | 片段注释与可视化（与上游联动） | 1DGR输出 / SV事件 + GenBank文件 | 交互式HTML仪表板（Plotly） |

### 分析流程

```
蛋白质FASTA + 基因组FASTA + GenBank注释
    │
    └─ [1] BactCG ──→ 核心基因组 + 系统发育树
         │
         ├─ [2] BactAG ──→ 祖先基因组 (AOG) + bactID.txt
         │    (可并行)
         ├─ [3] BactPG ──→ 泛基因组矩阵 (PG.txt)
         │
         └─ [4] BactPGA ──→ 注释后的基因表 (整合AG + PG)
              │
              ├─ [5] Bact1DGR ──→ 一维基因组表示
              │    └─→ [7] BactFragAnn (1DGR模式) ──→ 交互式马赛克图
              │
              └─ [6] BactEvolTraj ──→ 结构变异分析
                   └─→ [7] BactFragAnn (EvolTraj模式) ──→ 交互式环形图
```

**要点：** 从BactCG开始获取系统发育树。BactAG与BactPG可并行运行。BactPGA整合两者输出。下游模块消费BactAG/BactPGA结果。

---

## 🚀 快速开始

### 环境要求

- **Python** ≥ 3.8
- **Go** ≥ 1.18（编译二进制文件所需；已提供预编译版本）
- **外部工具**: `progressiveMauve`, `blastn`, `blastp`, `cd-hit`

Ubuntu/Debian 安装外部工具：
```bash
# Mauve比对器
sudo apt install mauve-aligner

# NCBI BLAST+
sudo apt install ncbi-blast+

# CD-HIT
sudo apt install cd-hit
```

### 安装（使用 Conda/Mamba）

```bash
# 创建并激活环境（也可使用 mamba）
conda create -n genomevo python=3.10 -y
conda activate genomevo

# 安装依赖
pip install biopython pandas numpy matplotlib plotly flask

# 安装 GenomEvo
cd /path/to/Genomevo
pip install -e .
```

> **💡 获取注释文件：** GFF和GBK注释文件可通过NCBI的[PGAP](https://github.com/ncbi/pgap)（原核基因组注释流程）生成。

---

## 💻 命令行使用

### 检查系统依赖

```bash
genomevo check
```

### 各模块使用示例

#### 1. BactAG – 祖先基因组重建

```bash
genomevo bactag \
    --tree ./input/tree_dir \
    --gene ./input/genome_dir \
    --threads 20 \
    --output BactAG_Results
```

**输入**：包含单个Newick格式系统发育树文件的目录，以及基因组FASTA文件目录（`.fasta`）。

**输出**：`BactAG_Results/` 包含重建的祖先基因组、`bactID.txt`（菌株谱系映射表）和处理日志。

#### 2. BactCG – 核心基因组分析（四步完整流程）

```bash
genomevo bactcg \
    --input ./input/proteins \
    --output BactCG_Results \
    --ref MG1655 \
    --cd-c 0.7 \
    --cg1 0.8
```

**输入**：蛋白质FASTA文件目录（每个菌株一个文件，接受 `.fasta`/`.faa`/`.fa`）。

**输出**：完整的四步流程：
- `CG_ALL.txt` – 核心基因存在/缺失表
- `all-strain-together/2.result/` – 每个家族的FASTA文件
- `all-strain-together/3.mega/` – MEG格式多序列比对
- `all-strain-together/4.SNP_mega/all_core_gene.meg` – SNP拼接比对

**步骤**：(1) QC过滤（可选）→ (2) CD-HIT聚类 → (3) BLAST直系同源分析 → (4) getfa提取 → (5) clustalw2多序列比对 → (6) SNP位点拼接

#### 3. BactPG – 泛基因组分析（统一单文件二进制）

```bash
genomevo bactpg \
    --seq ./input/proteins \
    --output BactPG_Results \
    --similarity 0.7 \
    --threads 30
```

**输入**：蛋白质FASTA文件目录（需要 `.fasta` 扩展名）。

**输出**：`BactPG_Results/result/PG.txt` – 含蛋白质ID的泛基因组存在/缺失矩阵。

**流程**：CD-HIT → 批次划分 → 全对全BLASTP → 直系同源过滤 → 迭代合并 → 最终PG.txt。使用自包含的单文件Go二进制，通过内部子命令调度实现。

#### 4. BactPGA – 泛基因组注释

```bash
# 全流程模式（推荐）
genomevo bactpga \
    --mode pipeline \
    --gbk ./input/genome.gbk \
    --pg ./BactPG_Results/PG.txt \
    --seq ./input/sequences

# 仅解析模式
genomevo bactpga --mode parse --gbk ./input/genome.gbk --output-file genes.tab.txt

# 注释模式
genomevo bactpga \
    --mode annotate \
    --gbk ./input/genome.gbk \
    --pg ./BactPG_Results/PG.txt \
    --strain MG1655 \
    --mutbest-dir ./BactCG_Results/mutbest
```

**输入**：GenBank文件、泛基因组矩阵和相互最佳比对数据。

**输出**：带泛基因组UID映射的注释基因表格。

#### 5. Bact1DGR – 一维基因组表示

```bash
genomevo bact1dgr \
    --base AncestralStrain \
    --bactid ./BactAG_Results/bactID.txt \
    --fasta-dir ./BactAG_Results/genomes \
    --workers 8
```

**输入**：基础菌株名称、BactAG生成的 `bactID.txt`、FASTA基因组目录。

**输出**：`OneDGR_Output/Final_Results/*.1DGR.txt` – 一维片段表示图谱。

#### 6. BactEvolTraj – 演化轨迹分析

```bash
# 使用JSON配置（推荐）
genomevo bactevoltraj --config evolt_config.json

# 或直接内联参数
genomevo bactevoltraj \
    --root-node S.enterica_AOG \
    --fasta-dir ./genomes \
    --gbk-dir ./annotations \
    --tree '[["root","mA"],["mA","mB"],["mB","mC"]]'
```

**输入**：根节点名称、FASTA和GenBank目录、进化树边定义。

**输出**：`Final_Large_SV_Analysis/` 包含SV事件表、覆盖度矩阵和Matplotlib图表。

#### 7. BactFragAnn – 片段注释与可视化

```bash
genomevo bactfragann \
    --mode 1dgr \
    --base-dir ./working_dir \
    --output Mosaic_Charts
```

**输入**：包含1DGR文本文件和GenBank文件的工作目录。

**输出**：`Mosaic_Charts/` 包含交互式Plotly HTML仪表板。

### 完整管线（自动化）

创建JSON配置文件（`pipeline.json`）按推荐顺序定义所有步骤：

```json
{
    "steps": [
        {"module": "bactcg",  "params": {"input_dir": "./proteins", "output_dir": "./cg_out",  "ref_strain": "MG1655"}},
        {"module": "bactag",  "params": {"tree_dir": "./tree", "gene_dir": "./genomes", "threads": 20, "output_dir": "./ag_out"}},
        {"module": "bactpg",  "params": {"seq_dir": "./proteins", "output_dir": "./pg_out", "similarity": 0.7}},
        {"module": "bactpga", "params": {"mode": "pipeline", "gbk_file": "./ag_out/AG_root.gbk", "pg_file": "./pg_out/PG.txt", "seq_dir": "./proteins"}},
        {"module": "bact1dgr","params": {"base_strain": "AG_root", "bactid_file": "./bactID.txt", "fasta_dir": "./genomes"}}
    ]
}
```

运行完整管线：
```bash
genomevo pipeline --config pipeline.json
```

---

## 🌐 Web界面使用

启动交互式Web界面：

```bash
genomevo web --port 5000
```

然后在浏览器打开 `http://localhost:5000`。

Web界面提供：
- **模块卡片**，清晰标注输入输出说明
- **各模块的分步表单**
- **管线流程可视化**
- **依赖检查器**
- **任务提交与实时状态追踪**

---

## 📦 项目结构

```
Genomevo/
├── README.md                          # 英文README
├── README_CN.md                       # 中文README（本文件）
├── setup.py                           # Python包安装器
├── run_genomevo.py                    # 快速启动入口脚本
├── genomevo/                          # 主包
│   ├── __init__.py                    # 包元数据
│   ├── cli.py                         # CLI入口（所有子命令）
│   ├── config.py                      # 全局配置与工具路径
│   ├── bin/                           # 预编译Go二进制文件（单文件）
│   │   ├── BactAG                     # 祖先基因组二进制
│   │   ├── bactcg                     # 核心基因组二进制（统一子命令）
│   │   ├── BactPG                     # 泛基因组二进制（统一单文件）
│   │   ├── bactpga                    # PG注释二进制
│   │   └── clustalw2                  # ClustalW2比对器
│   ├── modules/                       # 分析模块Python包装器
│   │   ├── __init__.py                # 模块导出
│   │   ├── bactag.py                  # BactAG包装器（工作目录隔离）
│   │   ├── bactcg.py                  # BactCG包装器（四步流程）
│   │   ├── bactpg.py                  # BactPG包装器（--yes 非交互）
│   │   ├── bactpga.py                 # BactPGA包装器
│   │   ├── bact1dgr.py                # Bact1DGR包装器
│   │   ├── bactevoltraj.py            # BactEvolTraj包装器
│   │   ├── bactfragann.py             # BactFragAnn包装器（可调用函数）
│   │   ├── BactFragAnn_for_1DGR.py    # 1DGR马赛克图生成器
│   │   ├── BactFragAnn_for_BactEvolTraj.py  # SV环形图生成器
│   │   ├── bactag_src/                # BactAG Go源码
│   │   ├── bactcg_src/                # BactCG Go源码（cobra CLI统一）
│   │   ├── bactpg_src/                # BactPG Go源码（仅main.go）
│   │   ├── bactpga_src/               # BactPGA Go源码
│   │   ├── onedgr/                    # OneDGR Python包
│   │   └── onedgr_src/                # OneDGR Go源码
│   ├── web/                           # Web界面
│   │   ├── app.py                     # Flask应用（10菌株限制）
│   │   ├── templates/                 # HTML模板
│   │   │   ├── index.html             # 首页（菌株限制提示）
│   │   │   ├── document.html          # 完整文档
│   │   │   └── module.html            # 模块配置页
│   │   └── static/
│   │       ├── style.css              # 样式表
│   │       └── workflow.svg           # 管线流程图
│   └── data/
│       └── pipeline_example.json      # 示例管线配置
└── workflow.svg                       # 主流程图
```

> **注意：** `GenomEvo_web/` 是一个独立的静态Web部署目录（位于包外部），包含可直接在浏览器中访问的静态HTML文档页面。
```

---

## 📖 输入输出规范

### BactAG

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：进化树** | Newick文本文件 | 树目录中的单个文件，如 `(A,B),(C,D);` |
| **输入：基因组** | FASTA文件（`.fasta`） | 每个菌株一个多FASTA文件 |
| **输出：AOG** | FASTA + GenBank | 重建的祖先基因组 |
| **输出：bactID.txt** | 表格文本 | 谱系记录：`父+子 Outside 兄弟 = AG_ID` |

### BactCG

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：蛋白质** | FASTA文件（`.fasta`） | 每菌株一个文件，蛋白质序列 |
| **输出：核心基因** | FASTA + MEG | 比对后的核心基因序列 |
| **输出：SNP** | 文本矩阵 | 跨菌株的SNP位置 |

### BactPG

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：序列** | FASTA文件 | 蛋白质或核酸序列 |
| **输出：PG.txt** | Tab分隔矩阵 | 跨菌株的基因存在/缺失模式 |

### BactPGA

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：GenBank** | `.gbk` | NCBI格式GenBank文件 |
| **输入：PG矩阵** | `PG.txt` | 来自BactPG |
| **输出：注释表** | `.tab.txt` | 含PG簇UID的基因表 |

### Bact1DGR

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：基础菌株** | FASTA `.fasta` | 祖先基因组序列 |
| **输入：补丁菌株** | FASTA `.fasta` | 后代基因组 |
| **输入：bactID.txt** | 文本 | 来自BactAG |
| **输出：1DGR** | `*.1dgr.txt` | Tab分隔的片段坐标 |

### BactEvolTraj

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：FASTA** | `.fasta` | 所有节点的基因组序列 |
| **输入：GenBank** | `.gbk` | 所有节点的基因注释 |
| **输入：进化树边** | Python列表 | `[("parent","child"), ...]` |
| **输出：SV表** | CSV | 每个分支的插入/缺失事件 |
| **输出：图表** | PNG/PDF | Matplotlib可视化 |

### BactFragAnn

| 项目 | 格式 | 说明 |
|------|--------|-------------|
| **输入：1DGR文件** | `.txt` | 来自Bact1DGR |
| **输入：GenBank文件** | `.gbk` | 基因注释 |
| **输出：HTML** | `.html` | 交互式Plotly仪表板 |

---

## 🔧 依赖项

### 外部工具（必须在$PATH中）

| 工具 | 被哪些模块使用 | 安装方式 |
|------|-------------|-------------|
| `progressiveMauve` | BactAG, Bact1DGR, BactEvolTraj | `sudo apt install mauve-aligner` |
| `blastn` / `blastp` | BactCG, BactPG | `sudo apt install ncbi-blast+` |
| `cd-hit` | BactCG, BactPG | `sudo apt install cd-hit` |
| `clustalw2` | BactCG | 已打包在 `genomevo/bin/` |

### Python包（pip自动安装）

`biopython`, `pandas`, `numpy`, `matplotlib`, `plotly`, `flask`

---

## 🧪 测试

安装后验证：

```bash
# 检查所有依赖
genomevo check

# 验证二进制文件可访问
genomevo/bin/BactAG --help

# 测试Web界面（本地启动）
genomevo web --port 5000
```

---

## 📚 引用

如果您在研究中使用了GenomEvo，请引用：

> Wang Y, Chen P, Zheng M, et al. GenomEvo: an efficient system delineating and annotating the evolutionary trajectories of bacterial genomes automatically. *(In preparation)*

---

## 📄 许可证

MIT License

---

## 🌍 相关数据库

- **先前工作（ESG工具）**: [https://resources.szu-bioinf.org/ESG/tools](https://resources.szu-bioinf.org/ESG/tools)
- **EEG数据库（大肠杆菌）**: [https://resources.szu-bioinf.org/EEG](https://resources.szu-bioinf.org/EEG) — *E. coli* 演化网络
- **ESEEG数据库（沙门氏菌）**: [https://resources.szu-bioinf.org/ESEEG](https://resources.szu-bioinf.org/ESEEG) — *Salmonella enterica* subsp. *enterica* 演化网络

## 🔗 链接

- **GitHub**: [https://github.com/YejunWang/GenomEvo](https://github.com/YejunWang/GenomEvo)
