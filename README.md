# 🧬 GenomEvo

**An Integrated Automated System for Bacterial Comparative and Evolutionary Genomics**

[![Python](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/)
[![Go](https://img.shields.io/badge/go-1.18+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**English** | [中文](README_CN.md)

---

## 📋 Overview

**GenomEvo** is an automated bacterial comparative and evolutionary genomics pipeline centered on Ancestral Orthologous Genomes (AOG). It integrates seven core modules into a unified, user-friendly platform, offering both a **command-line interface (CLI)** and a **web-based graphical user interface (GUI)**.

### Core Modules at a Glance

| Module | Function | Input | Output |
|--------|----------|-------|--------|
| **BactCG** | Core genome identification + phylogenetic tree inference | Protein FASTA files (`.fasta`/`.faa`/`.fa`) | Core gene clusters + MEG alignments + SNP matrix + phylogenetic tree |
| **BactAG** | Ancestral genome reconstruction | Phylogenetic tree (Newick format) + genome FASTA files | Ancestral genomes (AOG) + `bactID.txt` |
| **BactPG** | Pan-genome analysis (single-file unified binary) | Directory of sequence FASTA files | `PG.txt` pan-genome presence/absence matrix |
| **BactPGA** | Pan-genome annotation (integrates BactAG + BactPG) | GenBank files + PG matrix + reciprocal best hits | Gene table annotated with PG cluster UIDs |
| **Bact1DGR** | One-dimensional genome representation | AOG + patch strains + `bactID.txt` | `*.1dgr.txt` fragment coordinate map |
| **BactEvolTraj** | Evolutionary trajectory analysis | FASTA + GenBank + tree edges (three-way Mauve) | SV event tables + coverage matrices + charts |
| **BactFragAnn** | Fragment annotation & visualization (linked with upstream) | 1DGR output / SV events + GenBank files | Interactive HTML dashboard (Plotly) |

### Analysis Workflow

```
Protein FASTA + Genome FASTA + GenBank annotations
    │
    └─ [1] BactCG ──→ Core genome + phylogenetic tree
         │
         ├─ [2] BactAG ──→ Ancestral genomes (AOG) + bactID.txt
         │    (runnable in parallel)
         ├─ [3] BactPG ──→ Pan-genome matrix (PG.txt)
         │
         └─ [4] BactPGA ──→ Annotated gene table (integrating AG + PG)
              │
              ├─ [5] Bact1DGR ──→ One-dimensional genome representation
              │    └─→ [7] BactFragAnn (1DGR mode) ──→ Interactive mosaic charts
              │
              └─ [6] BactEvolTraj ──→ Structural variation analysis
                   └─→ [7] BactFragAnn (EvolTraj mode) ──→ Interactive circular diagrams
```

**Key point:** Begin with BactCG to obtain the phylogenetic tree. BactAG and BactPG can run in parallel. BactPGA integrates the outputs of both. Downstream modules consume BactAG / BactPGA results.

---

## 🚀 Quick Start

### Requirements

- **Python** ≥ 3.8
- **Go** ≥ 1.18 (required for compiling binaries; pre-built binaries are provided)
- **External tools**: `progressiveMauve`, `blastn`, `blastp`, `cd-hit`

Installing external tools on Ubuntu/Debian:
```bash
# Mauve aligner
sudo apt install mauve-aligner

# NCBI BLAST+
sudo apt install ncbi-blast+

# CD-HIT
sudo apt install cd-hit
```

### Installation (using Conda / Mamba)

```bash
# Create and activate environment
mamba create -n genomevo python=3.10 -y
mamba activate genomevo

# Install dependencies
pip install biopython pandas numpy matplotlib plotly flask

# Install GenomEvo
cd /path/to/Genomevo
pip install -e .
```

### Using the Pre-configured Mamba Environment

```bash
# The project includes a pre-built "bactet" environment with all dependencies installed
mamba run -n bactet genomevo --help

# Launch the web interface
mamba run -n bactet genomevo web --port 8070
```

> **💡 Obtaining annotation files:** GFF and GBK annotation files can be generated via NCBI's [PGAP](https://github.com/ncbi/pgap) (Prokaryotic Genome Annotation Pipeline).

---

## 💻 Command-Line Usage

### Check System Dependencies

```bash
genomevo check
```

### Module Usage Examples

#### 1. BactAG – Ancestral Genome Reconstruction

```bash
genomevo bactag \
    --tree ./input/tree_dir \
    --gene ./input/genome_dir \
    --threads 20 \
    --output BactAG_Results
```

**Input**: A directory containing a single phylogenetic tree file in Newick format, and a directory of genome FASTA files (`.fasta`).

**Output**: `BactAG_Results/` contains reconstructed ancestral genomes, `bactID.txt` (strain-lineage mapping table), and processing logs.

#### 2. BactCG – Core Genome Analysis (four-step complete pipeline)

```bash
genomevo bactcg \
    --input ./input/proteins \
    --output BactCG_Results \
    --ref MG1655 \
    --cd-c 0.7 \
    --cg1 0.8
```

**Input**: Directory of protein FASTA files (one file per strain; accepts `.fasta`/`.faa`/`.fa`).

**Output**: Complete four-step pipeline:
- `CG_ALL.txt` – core gene presence/absence table
- `all-strain-together/2.result/` – FASTA files per family
- `all-strain-together/3.mega/` – MEG-format multiple sequence alignments
- `all-strain-together/4.SNP_mega/all_core_gene.meg` – concatenated SNP alignment

**Steps**: (1) QC filtering (optional) → (2) CD-HIT clustering → (3) BLAST orthology analysis → (4) getfa extraction → (5) clustalw2 multiple sequence alignment → (6) SNP site concatenation

#### 3. BactPG – Pan-genome Analysis (unified single-file binary)

```bash
genomevo bactpg \
    --seq ./input/proteins \
    --output BactPG_Results \
    --similarity 0.7 \
    --threads 30
```

**Input**: Directory of protein FASTA files (`.fasta` extension required).

**Output**: `BactPG_Results/result/PG.txt` – pan-genome presence/absence matrix with protein IDs.

**Pipeline**: CD-HIT → batch partitioning → all-vs-all BLASTP → orthology filtering → iterative merging → final PG.txt. Powered by a self-contained, single-file Go binary with internal subcommand dispatch.

#### 4. BactPGA – Pan-genome Annotation

```bash
# Pipeline mode (recommended)
genomevo bactpga \
    --mode pipeline \
    --gbk ./input/genome.gbk \
    --pg ./BactPG_Results/PG.txt \
    --seq ./input/sequences

# Parse-only mode
genomevo bactpga --mode parse --gbk ./input/genome.gbk --output-file genes.tab.txt

# Annotation mode
genomevo bactpga \
    --mode annotate \
    --gbk ./input/genome.gbk \
    --pg ./BactPG_Results/PG.txt \
    --strain MG1655 \
    --mutbest-dir ./BactCG_Results/mutbest
```

**Input**: GenBank files, pan-genome matrix, and reciprocal best-hit data.

**Output**: Annotated gene table with pan-genome UID mappings.

#### 5. Bact1DGR – One-dimensional Genome Representation

```bash
genomevo bact1dgr \
    --base AncestralStrain \
    --bactid ./BactAG_Results/bactID.txt \
    --fasta-dir ./BactAG_Results/genomes \
    --workers 8
```

**Input**: Base strain name, `bactID.txt` from BactAG, FASTA genome directory.

**Output**: `OneDGR_Output/Final_Results/*.1DGR.txt` – one-dimensional fragment representation maps.

#### 6. BactEvolTraj – Evolutionary Trajectory Analysis

```bash
# Using a JSON config (recommended)
genomevo bactevoltraj --config evolt_config.json

# Or inline parameters directly
genomevo bactevoltraj \
    --root-node S.enterica_AOG \
    --fasta-dir ./genomes \
    --gbk-dir ./annotations \
    --tree '[["root","mA"],["mA","mB"],["mB","mC"]]'
```

**Input**: Root node name, FASTA and GenBank directories, tree edge definitions.

**Output**: `Final_Large_SV_Analysis/` contains SV event tables, coverage matrices, and Matplotlib charts.

#### 7. BactFragAnn – Fragment Annotation & Visualization

```bash
genomevo bactfragann \
    --mode 1dgr \
    --base-dir ./working_dir \
    --output Mosaic_Charts
```

**Input**: Working directory containing 1DGR text files and GenBank files.

**Output**: `Mosaic_Charts/` containing interactive Plotly HTML dashboards.

### Full Pipeline (Automated)

Create a JSON configuration file (`pipeline.json`) defining all steps in the recommended order:

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

Run the full pipeline:
```bash
genomevo pipeline --config pipeline.json
```

---

## 🌐 Web Interface

Launch the interactive web interface:

```bash
genomevo web --port 8080
```

Then open `http://localhost:8080` in your browser.

The web interface provides:
- **Module cards** with clear input/output descriptions
- **Step-by-step forms** for each module
- **Pipeline workflow visualization**
- **Dependency checker**
- **Task submission with real-time status tracking**

---

## 📦 Project Structure

```
Genomevo/
├── README.md                          # English README
├── README_CN.md                       # Chinese README
├── setup.py                           # Python package installer
├── run_genomevo.py                    # Quick-start entry script
├── genomevo/                          # Main package
│   ├── __init__.py                    # Package metadata
│   ├── cli.py                         # CLI entry point (all subcommands)
│   ├── config.py                      # Global config & tool paths
│   ├── bin/                           # Pre-compiled Go binaries (single-file)
│   │   ├── BactAG                     # Ancestral genome binary
│   │   ├── bactcg                     # Core genome binary (unified subcommands)
│   │   ├── BactPG                     # Pan-genome binary (unified single-file)
│   │   ├── bactpga                    # PG annotation binary
│   │   └── clustalw2                  # ClustalW2 aligner
│   ├── modules/                       # Analysis module Python wrappers
│   │   ├── __init__.py                # Module exports
│   │   ├── bactag.py                  # BactAG wrapper (working-directory isolation)
│   │   ├── bactcg.py                  # BactCG wrapper (four-step pipeline)
│   │   ├── bactpg.py                  # BactPG wrapper (--yes non-interactive)
│   │   ├── bactpga.py                 # BactPGA wrapper
│   │   ├── bact1dgr.py                # Bact1DGR wrapper
│   │   ├── bactevoltraj.py            # BactEvolTraj wrapper
│   │   ├── bactfragann.py             # BactFragAnn wrapper (callable functions)
│   │   ├── BactFragAnn_for_1DGR.py    # 1DGR mosaic chart generator
│   │   ├── BactFragAnn_for_BactEvolTraj.py  # SV circular diagram generator
│   │   ├── bactag_src/                # BactAG Go source
│   │   ├── bactcg_src/                # BactCG Go source (cobra CLI unified)
│   │   ├── bactpg_src/                # BactPG Go source (main.go only)
│   │   ├── bactpga_src/               # BactPGA Go source
│   │   ├── onedgr/                    # OneDGR Python package
│   │   └── onedgr_src/                # OneDGR Go source
│   ├── web/                           # Web interface
│   │   ├── app.py                     # Flask application (10-strain limit)
│   │   ├── templates/                 # HTML templates
│   │   │   ├── index.html             # Home page (strain limit notice)
│   │   │   ├── document.html          # Full documentation
│   │   │   └── module.html            # Module configuration page
│   │   └── static/
│   │       ├── style.css              # Stylesheet
│   │       └── workflow.svg           # Pipeline workflow diagram
│   └── data/
│       └── pipeline_example.json      # Example pipeline configuration
├── Genomevo_web/                      # Standalone web deployment
│   ├── index.html                     # Static landing page (direct access)
│   ├── app.py                         # Flask app with sub-path deployment support
│   ├── .htaccess                      # Apache mod_rewrite rules
│   └── templates/                     # Web interface templates
└── workflow.svg                       # Master workflow diagram
```

---

## 📖 Input / Output Specifications

### BactAG

| Item | Format | Description |
|------|--------|-------------|
| **Input: tree** | Newick text file | A single file in the tree directory, e.g. `(A,B),(C,D);` |
| **Input: genomes** | FASTA files (`.fasta`) | One multi-FASTA file per strain |
| **Output: AOG** | FASTA + GenBank | Reconstructed ancestral genomes |
| **Output: bactID.txt** | Tabular text | Lineage records: `Parent+Child Outside Sibling = AG_ID` |

### BactCG

| Item | Format | Description |
|------|--------|-------------|
| **Input: proteins** | FASTA files (`.fasta`) | One file per strain, protein sequences |
| **Output: core genes** | FASTA + MEG | Aligned core gene sequences |
| **Output: SNPs** | Text matrix | SNP positions across strains |

### BactPG

| Item | Format | Description |
|------|--------|-------------|
| **Input: sequences** | FASTA files | Protein or nucleotide sequences |
| **Output: PG.txt** | Tab-delimited matrix | Gene presence/absence patterns across strains |

### BactPGA

| Item | Format | Description |
|------|--------|-------------|
| **Input: GenBank** | `.gbk` | NCBI-format GenBank file |
| **Input: PG matrix** | `PG.txt` | From BactPG |
| **Output: annotation table** | `.tab.txt` | Gene table with PG cluster UIDs |

### Bact1DGR

| Item | Format | Description |
|------|--------|-------------|
| **Input: base strain** | FASTA `.fasta` | Ancestral genome sequence |
| **Input: patch strains** | FASTA `.fasta` | Descendant genomes |
| **Input: bactID.txt** | Text | From BactAG |
| **Output: 1DGR** | `*.1dgr.txt` | Tab-delimited fragment coordinates |

### BactEvolTraj

| Item | Format | Description |
|------|--------|-------------|
| **Input: FASTA** | `.fasta` | Genome sequences for all nodes |
| **Input: GenBank** | `.gbk` | Gene annotations for all nodes |
| **Input: tree edges** | Python list | `[("parent","child"), ...]` |
| **Output: SV tables** | CSV | Insertion/deletion events per branch |
| **Output: charts** | PNG/PDF | Matplotlib visualizations |

### BactFragAnn

| Item | Format | Description |
|------|--------|-------------|
| **Input: 1DGR files** | `.txt` | From Bact1DGR |
| **Input: GenBank files** | `.gbk` | Gene annotations |
| **Output: HTML** | `.html` | Interactive Plotly dashboards |

---

## 🔧 Dependencies

### External Tools (must be on `$PATH`)

| Tool | Used By | Installation |
|------|---------|-------------|
| `progressiveMauve` | BactAG, Bact1DGR, BactEvolTraj | `sudo apt install mauve-aligner` |
| `blastn` / `blastp` | BactCG, BactPG | `sudo apt install ncbi-blast+` |
| `cd-hit` | BactCG, BactPG | `sudo apt install cd-hit` |
| `clustalw2` | BactCG | Bundled in `genomevo/bin/` |

### Python Packages (auto-installed via pip)

`biopython`, `pandas`, `numpy`, `matplotlib`, `plotly`, `flask`

---

## 🧪 Testing

Post-installation verification:

```bash
# Check all dependencies
genomevo check

# Verify binary accessibility
genomevo/bin/BactAG --help

# Test the web interface (launch locally)
genomevo web --port 8080
```

---

## 📚 Citation

If you use GenomEvo in your research, please cite:

> Wang Y, Chen P, Zheng M, et al. GenomEvo: an efficient system delineating and annotating the evolutionary trajectories of bacterial genomes automatically. *(In preparation)*

---

## 📄 License

MIT License

---

## 🌍 Related Databases

- **Prior work (ESG tools)**: [https://resources.szu-bioinf.org/ESG/tools](https://resources.szu-bioinf.org/ESG/tools)
- **EEG database (E. coli)**: [https://resources.szu-bioinf.org/EEG](https://resources.szu-bioinf.org/EEG) — *E. coli* evolutionary network
- **ESEEG database (Salmonella)**: [https://resources.szu-bioinf.org/ESEEG](https://resources.szu-bioinf.org/ESEEG) — *Salmonella enterica* subsp. *enterica* evolutionary network

## 🔗 Links

- **Software homepage**: [https://tools.szu-bioinf.org/GenomEvo/](https://tools.szu-bioinf.org/GenomEvo/)
- **GitHub**: [https://github.com/YejunWang/GenomEvo](https://github.com/YejunWang/GenomEvo)
