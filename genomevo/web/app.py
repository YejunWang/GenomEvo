"""
GenomEvo Web UI – Flask-based web interface.
Academic publication-style GUI with module explanation and run functionality.
"""

import os, sys, json, uuid, threading
from datetime import datetime
from flask import Flask, render_template, request, jsonify, redirect, url_for, flash

_PKG_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if _PKG_ROOT not in sys.path:
    sys.path.insert(0, _PKG_ROOT)


def create_app():
    app = Flask(
        __name__,
        template_folder=os.path.join(os.path.dirname(__file__), "templates"),
        static_folder=os.path.join(os.path.dirname(__file__), "static"),
    )
    app.secret_key = os.urandom(24).hex()
    jobs = {}

    @app.route("/")
    def index():
        modules = [
            {"id":"bactag","name":"BactAG","name_cn":"祖先基因组重建",
             "desc":"Reconstruct bacterial ancestral genomes using a phylogenetically guided, bottom-up Backbone-Patching (BP) strategy. Automatically resolves AOGs for every phylogenetic node defined by the input Newick tree.",
             "desc_cn":"采用系统发育引导的自下而上Backbone-Patching策略重建细菌祖先基因组，自动解析输入Newick进化树所定义的每个系统发育节点的祖先直系同源基因组（AOG）。",
             "inputs":"Newick tree file + multi-FASTA genome files (.fasta)","outputs":"Ancestral genomes (FASTA) + bactID.txt lineage map","depends":"progressiveMauve"},
            {"id":"bactcg","name":"BactCG","name_cn":"核心基因组分析",
             "desc":"Identify the core genome from protein sequences using CD-HIT redundancy clustering and BLAST-based reciprocal best-hit orthology analysis. Produces MEG-format alignments and SNP matrices.",
             "desc_cn":"通过CD-HIT冗余聚类和基于BLAST的相互最佳比对直系同源分析鉴定核心基因组，输出MEG格式多序列比对和SNP矩阵。",
             "inputs":"Protein FASTA files (one per strain)","outputs":"Core gene clusters + MEG alignments + SNP data","depends":"BLAST+, CD-HIT, ClustalW2"},
            {"id":"bactpg","name":"BactPG","name_cn":"泛基因组分析",
             "desc":"Perform pan-genome analysis to catalog gene presence/absence across all input strains. Uses iterative batch BLAST with automatic consolidation into a comprehensive presence-absence matrix.",
             "desc_cn":"通过迭代批次BLAST与自动合并进行泛基因组分析，生成跨所有菌株的基因存在/缺失矩阵，系统刻画细菌种群的泛基因组组成。",
             "inputs":"Sequence FASTA files directory","outputs":"PG.txt pan-genome matrix + clustering data","depends":"BLAST+, CD-HIT"},
            {"id":"bactpga","name":"BactPGA","name_cn":"泛基因组注释",
             "desc":"Annotate genomic features with pan-genome cluster UIDs via mutual best hit alignment. Supports PGAG and RAST formats. Extracts CDS, rRNA, tRNA, ncRNA, misc_feature from GenBank files.",
             "desc_cn":"通过相互最佳比对将基因组特征映射至泛基因组簇UID。支持PGAG与RAST注释格式，可从GenBank文件中提取CDS、rRNA、tRNA、ncRNA及misc_feature元件。",
             "inputs":"GenBank (.gbk) + PG.txt + mutual best hit files","outputs":"Annotated gene table (.tab.txt)","depends":"None (pure Go)"},
            {"id":"bact1dgr","name":"Bact1DGR","name_cn":"一维基因组表示",
             "desc":"Generate 1D linear genomic representations comparing ancestral to descendant genomes via progressiveMauve alignment and orthologous block extraction. Automatically traces strain lineages through bactID.txt.",
             "desc_cn":"通过progressiveMauve比对和直系同源块提取，生成祖先与后代基因组之间的一维线性比较表示。可自动通过bactID.txt追溯菌株谱系关系。",
             "inputs":"Base strain + patches + FASTA + bactID.txt","outputs":"*.1dgr.txt fragment coordinate maps","depends":"progressiveMauve, BactAG output"},
            {"id":"bactevoltraj","name":"BactEvolTraj","name_cn":"演化轨迹分析",
             "desc":"Analyze large structural variants (insertions, deletions, inversions) along evolutionary tree branches using three-way progressiveMauve alignment with gene-level annotation of affected regions.",
             "desc_cn":"利用三向progressiveMauve比对分析沿进化树分支的大片段结构变异（插入、缺失、倒位），并对受影响的基因组区域进行基因级注释。",
             "inputs":"FASTA genomes + GenBank annotations + tree edges","outputs":"SV event tables (CSV) + coverage matrices + plots","depends":"progressiveMauve, BioPython, Matplotlib, NumPy, Pandas"},
            {"id":"bactfragann","name":"BactFragAnn","name_cn":"片段注释与可视化",
             "desc":"Create interactive HTML visualization dashboards for genomic fragment annotation. Supports mosaic charts (1DGR output) and circular SV event plots with clickable gene-level details.",
             "desc_cn":"生成基因组片段注释的交互式HTML可视化仪表板。支持1DGR马赛克图和结构变异环形图，可点击查看基因级详细信息。",
             "inputs":"1DGR/SV data + GenBank annotation files","outputs":"Interactive HTML dashboards (Plotly)","depends":"Plotly, BioPython, Pandas"},
        ]
        return render_template("index.html", modules=modules)

    @app.route("/document")
    def document():
        """Software documentation with workflow explanation and worked example."""
        return render_template("document.html")

    @app.route("/module/<module_id>")
    def module_page(module_id):
        module_info = {
            "bactag": {
                "title": "BactAG — Ancestral Genome Reconstruction",
                "title_cn": "BactAG — 祖先基因组重建",
                "overview": "Reconstructs bacterial ancestral genomes (AOGs) for each phylogenetic node defined by the input Newick-format tree, using a Backbone-Patching strategy with progressiveMauve alignment. The output AOGs and bactID.txt lineage map serve as essential inputs for downstream modules (Bact1DGR, BactEvolTraj).",
                "overview_cn": "基于Backbone-Patching策略和progressiveMauve比对，为输入Newick格式进化树所定义的每个系统发育节点重建细菌祖先基因组（AOG）。输出的AOG和bactID.txt谱系映射表是下游模块（Bact1DGR、BactEvolTraj）的关键输入。",
                "form_fields": [
                    {"name":"tree_dir","label":"Tree Directory","label_cn":"进化树目录","type":"text","placeholder":"/path/to/tree_dir","required":True,"help":"Directory containing exactly one Newick-format phylogenetic tree file. The tree defines the hierarchical relationships among strains and determines the reconstruction order.","help_cn":"包含恰好一个Newick格式系统发育树文件的目录。该树定义了菌株间的层级关系，并决定了祖先基因组重建的顺序。"},
                    {"name":"gene_dir","label":"Genome Directory","label_cn":"基因组目录","type":"text","placeholder":"/path/to/genome_dir","required":True,"help":"Directory containing multi-FASTA genome files (.fasta extension). Each file corresponds to one strain; the filename (without extension) is used as the strain identifier.","help_cn":"包含多FASTA格式基因组文件（.fasta扩展名）的目录。每个文件对应一个菌株，文件名（不含扩展名）将作为菌株标识符。"},
                    {"name":"threads","label":"Threads (-t)","label_cn":"线程数","type":"number","placeholder":"20","required":False,"help":"Number of parallel processing threads controlling concurrency of Mauve alignment tasks. Default: 20. Increase for larger datasets; reduce if memory is limited.","help_cn":"并行处理线程数，控制Mauve比对任务的并发度。默认值：20。大数据集可适当增加，内存受限时可减小。"},
                    {"name":"id_file","label":"bactID Output Path (-id)","label_cn":"bactID输出路径","type":"text","placeholder":"bactID.txt","required":False,"help":"Path where the bactID.txt lineage map will be written. This file records all parent-child-sibling relationships resolved during reconstruction and is required by Bact1DGR.","help_cn":"bactID.txt谱系映射表的输出路径。该文件记录了重建过程中解析的所有父子-兄弟关系，是Bact1DGR的必需输入。"},
                    {"name":"output_dir","label":"Output Directory","label_cn":"输出目录","type":"text","placeholder":"BactAG_Results","required":False,"help":"Working/output directory where all BactAG results will be placed. Default: 'BactAG_Results'. The Go binary runs inside this directory.","help_cn":"BactAG的工作/输出目录，所有结果将存放在此。默认：'BactAG_Results'。Go程序将在该目录内运行。"},
                ],
            },
            "bactcg": {
                "title": "BactCG — Core Genome Analysis",
                "title_cn": "BactCG — 核心基因组分析",
                "overview": "Identifies the core genome from protein FASTA files. Pipeline: (1) optional QC filtering of abnormally small proteomes, (2) CD-HIT clustering to remove intra-strain redundancy, (3) all-vs-all BLASTP for inter-strain comparison, (4) BactCG orthology computation using reciprocal best-hit criteria.",
                "overview_cn": "从蛋白质FASTA文件鉴定核心基因组。流程：(1) 可选的QC过滤以剔除异常小的蛋白质组，(2) CD-HIT聚类去除株内冗余，(3) 全对全BLASTP进行株间比较，(4) 基于相互最佳比对标准的BactCG直系同源计算。",
                "form_fields": [
                    {"name":"input_dir","label":"Input Directory (-i)","label_cn":"输入目录","type":"text","placeholder":"/path/to/protein_files","required":True,"help":"Directory containing protein FASTA files, one per strain. File extension must be .fasta. Each file should contain all predicted protein sequences for one strain.","help_cn":"包含蛋白质FASTA文件的目录，每个菌株一个文件。扩展名必须为.fasta，每个文件应包含一个菌株的所有预测蛋白质序列。"},
                    {"name":"output_dir","label":"Output Directory (-o)","label_cn":"输出目录","type":"text","placeholder":"/path/to/output","required":True,"help":"Directory where all BactCG results will be written, including CD-HIT output, BLAST results, and final core genome data.","help_cn":"所有BactCG结果的输出目录，包括CD-HIT输出、BLAST比对结果和最终核心基因组数据。"},
                    {"name":"ref_strain","label":"Reference Strain (--ref)","label_cn":"参考菌株","type":"text","placeholder":"MG1655","required":True,"help":"Reference strain identifier used as the anchor for orthology determination. Must match the filename (without .fasta extension) of one input protein file. All other strains are compared against this reference.","help_cn":"用作直系同源判定锚点的参考菌株标识符。必须与某个输入蛋白质文件的文件名（不含.fasta扩展名）匹配。所有其他菌株均与此参考菌株进行比较。"},
                    {"name":"cd_cutoff","label":"CD-HIT Identity (--cd-c)","label_cn":"CD-HIT序列一致性阈值","type":"number","placeholder":"0.7","required":False,"help":"Sequence identity threshold for CD-HIT clustering. Range: 0.0–1.0. Default: 0.7. Higher values produce tighter, more stringent clusters.","help_cn":"CD-HIT聚类的序列一致性阈值，范围0.0–1.0，默认0.7。较高的值产生更紧密、更严格的聚类。"},
                    {"name":"cd_s","label":"CD-HIT Length Diff (--cd-s)","label_cn":"CD-HIT长度差异阈值","type":"number","placeholder":"0.7","required":False,"help":"Length difference cutoff for CD-HIT. Range: 0.0–1.0. Default: 0.7. Value 0 means no length restriction; 1.0 requires identical length.","help_cn":"CD-HIT的长度差异阈值，范围0.0–1.0，默认0.7。值为0表示无长度限制，1.0要求长度完全相同。"},
                    {"name":"cg1","label":"Orthology Cutoff 1 (--cg1)","label_cn":"直系同源阈值1","type":"number","placeholder":"0.8","required":False,"help":"Primary orthology cutoff for core gene inclusion. Range: 0.0–1.0. Default: 0.8. Defines the minimum similarity/coverage for core genome membership.","help_cn":"核心基因纳入的主要直系同源阈值，范围0.0–1.0，默认0.8。定义了基因被纳入核心基因组所需的最低相似性/覆盖度。"},
                    {"name":"cg2","label":"Orthology Cutoff 2 (--cg2)","label_cn":"直系同源阈值2","type":"number","placeholder":"0.9","required":False,"help":"Secondary, stricter orthology threshold. Range: 0.0–1.0. Default: 0.9. Applied in a second pass for higher-confidence core gene selection.","help_cn":"更严格的次级直系同源阈值，范围0.0–1.0，默认0.9。用于第二轮更高置信度的核心基因筛选。"},
                    {"name":"skip_qc","label":"Skip QC Filtering","label_cn":"跳过QC过滤","type":"checkbox","placeholder":"","required":False,"help":"If checked, the QC step (filtering protein files smaller than 90% of the average size) will be skipped. Use when all input proteomes are known to be complete.","help_cn":"若勾选，将跳过QC步骤（过滤小于平均大小90%的蛋白质文件）。当所有输入蛋白质组已知是完整且高质量时可使用。"},
                ],
            },
            "bactpg": {
                "title": "BactPG — Pan-Genome Analysis",
                "title_cn": "BactPG — 泛基因组分析",
                "overview": "Performs comprehensive pan-genome analysis to generate a gene presence-absence matrix. Workflow: (1) CD-HIT clustering per strain, (2) random batch grouping, (3) all-vs-all BLAST within each batch, (4) filtering by similarity/coverage, (5) iterative consolidation into the final PG.txt matrix.",
                "overview_cn": "执行全面的泛基因组分析以生成基因存在/缺失矩阵。工作流程：(1) 每个菌株的CD-HIT聚类，(2) 随机批次分组，(3) 每批次内全对全BLAST，(4) 按相似性/覆盖度过滤，(5) 迭代合并生成最终的PG.txt矩阵。",
                "form_fields": [
                    {"name":"seq_dir","label":"Sequence Directory (-s)","label_cn":"序列目录","type":"text","placeholder":"/path/to/sequences","required":True,"help":"Directory containing sequence FASTA files (protein or nucleotide). Each file represents one strain; the filename (without extension) is the strain identifier in the output PG matrix.","help_cn":"包含序列FASTA文件（蛋白质或核酸）的目录。每个文件代表一个菌株，文件名（不含扩展名）将在输出PG矩阵中用作菌株标识符。"},
                    {"name":"output_dir","label":"Output Directory","label_cn":"输出目录","type":"text","placeholder":"BactPG_Results","required":False,"help":"Directory where BactPG results will be written. Default: 'BactPG_Results'. The result/PG.txt file will be created inside this directory.","help_cn":"BactPG结果的输出目录。默认：'BactPG_Results'。结果文件result/PG.txt将创建在此目录内。"},
                    {"name":"similarity","label":"Similarity Threshold","label_cn":"相似性阈值","type":"number","placeholder":"0.7","required":False,"help":"Sequence similarity threshold for gene clustering. Range: 0.0–1.0. Default: 0.7. Genes with pairwise similarity above this threshold are grouped into the same pan-genome cluster.","help_cn":"基因聚类的序列相似性阈值，范围0.0–1.0，默认0.7。高于此阈值的基因将被归入同一泛基因组簇。"},
                    {"name":"threads","label":"Number of Threads (-t)","label_cn":"线程数","type":"number","placeholder":"30","required":False,"help":"Number of parallel BLAST threads. Default: 30. BactPG is computationally intensive; increase this value on multi-core servers for faster processing.","help_cn":"并行BLAST线程数，默认30。BactPG计算密集，在多核服务器上可适当增大以加速处理。"},
                ],
            },
            "bactpga": {
                "title": "BactPGA — Pan-Genome Annotation",
                "title_cn": "BactPGA — 泛基因组注释",
                "overview": "Annotates genomic features with pan-genome cluster UIDs. Three modes: (a) pipeline — full automatic parse + CG + annotate workflow; (b) parse — extract gene features from GenBank only; (c) annotate — assign PG UIDs to a pre-parsed gene table using mutual best hits.",
                "overview_cn": "将基因组特征注释至泛基因组簇UID。三种模式：(a) pipeline — 全自动解析+CG+注释工作流；(b) parse — 仅从GenBank提取基因特征；(c) annotate — 使用相互最佳比对为预解析基因表分配PG UID。",
                "form_fields": [
                    {"name":"mode","label":"Operation Mode","label_cn":"运行模式","type":"select","options":["pipeline","parse","annotate"],"required":True,"help":"pipeline: full automatic workflow (parse + CG + annotate). parse: extract features from GenBank only. annotate: assign PG UIDs to an already-parsed gene table.","help_cn":"pipeline: 全自动工作流。parse: 仅从GenBank提取特征。annotate: 为已解析的基因表分配PG UID。"},
                    {"name":"gbk_file","label":"GenBank File (--gbk)","label_cn":"GenBank文件","type":"text","placeholder":"/path/to/genome.gbk","required":True,"help":"Input GenBank format file (.gbk) containing genome sequence and gene annotations. Required for all modes.","help_cn":"输入GenBank格式文件（.gbk），包含基因组序列和基因注释。所有模式均必需。"},
                    {"name":"pg_file","label":"PG Matrix (--pg)","label_cn":"泛基因组矩阵","type":"text","placeholder":"/path/to/PG.txt","required":True,"help":"Path to the pan-genome presence-absence matrix (PG.txt) generated by BactPG. Required for annotate and pipeline modes.","help_cn":"BactPG生成的泛基因组存在/缺失矩阵（PG.txt）的路径。annotate和pipeline模式必需。"},
                    {"name":"seq_dir","label":"Protein Sequence Directory (--seq)","label_cn":"蛋白质序列目录","type":"text","placeholder":"/path/to/sequences","required":False,"help":"Directory containing protein FASTA files. Required for pipeline mode. Used for BLAST-based mutual best hit computation.","help_cn":"包含蛋白质FASTA文件的目录。pipeline模式必需，用于基于BLAST的相互最佳比对计算。"},
                    {"name":"strain","label":"Strain Name (--strain)","label_cn":"菌株名称","type":"text","placeholder":"MG1655","required":False,"help":"Target strain name for annotation. Required for annotate mode. Must match the strain identifier used in the PG matrix.","help_cn":"注释的目标菌株名称。annotate模式必需，必须与PG矩阵中使用的菌株标识符匹配。"},
                    {"name":"mutbest_dir","label":"Mutual Best Hit Dir (--mutbestDir)","label_cn":"相互最佳比对目录","type":"text","placeholder":"/path/to/mutbest_results","required":False,"help":"Directory containing .mutbest.filt.txt mutual best hit files. Required for annotate mode. Generated by BactCG.","help_cn":"包含.mutbest.filt.txt相互最佳比对文件的目录。annotate模式必需，由BactCG生成。"},
                    {"name":"cov1","label":"Coverage Threshold 1 (--cov1)","label_cn":"覆盖度阈值1","type":"number","placeholder":"0.7","required":False,"help":"Primary coverage threshold for orthology filtering in pipeline mode. Range: 0.0–1.0. Default: 0.7.","help_cn":"pipeline模式中直系同源过滤的主要覆盖度阈值，范围0.0–1.0，默认0.7。"},
                    {"name":"cov2","label":"Coverage Threshold 2 (--cov2)","label_cn":"覆盖度阈值2","type":"number","placeholder":"0.7","required":False,"help":"Secondary coverage threshold for orthology filtering in pipeline mode. Range: 0.0–1.0. Default: 0.7.","help_cn":"pipeline模式中直系同源过滤的次级覆盖度阈值，范围0.0–1.0，默认0.7。"},
                ],
            },
            "bact1dgr": {
                "title": "Bact1DGR — 1D Genomic Representation",
                "title_cn": "Bact1DGR — 一维基因组表示",
                "overview": "Generates 1D linear representations comparing an ancestral genome to its descendants. Pipeline: (1) trace strain lineage using bactID.txt, (2) run progressiveMauve alignment for each ancestor-descendant pair, (3) extract orthologous blocks, (4) merge results into a unified 1D coordinate map.",
                "overview_cn": "生成祖先基因组与其后代的一维线性比较表示。流程：(1) 利用bactID.txt追溯菌株谱系，(2) 对每个祖先-后代对运行progressiveMauve比对，(3) 提取直系同源块，(4) 合并结果为统一的一维坐标图谱。",
                "form_fields": [
                    {"name":"base_strain","label":"Base Strain (--base)","label_cn":"基础菌株","type":"text","placeholder":"AG_0001","required":True,"help":"Base/ancestral strain name. Must match a .fasta filename (without extension) in the FASTA directory. Serves as the reference coordinate system for the 1D representation.","help_cn":"基础/祖先菌株名称，必须与FASTA目录中的某个.fasta文件名（不含扩展名）匹配。该菌株将作为一维表示的参考坐标系统。"},
                    {"name":"patches","label":"Patch Strains (--patches)","label_cn":"补丁菌株（逗号分隔）","type":"text","placeholder":"strainA,strainB,strainC","required":False,"help":"Comma-separated list of descendant/patch strain names. If not provided, strains are automatically resolved from the bactID file based on the base strain's ancestral lineage.","help_cn":"逗号分隔的后代/补丁菌株名称列表。若不提供，将根据基础菌株的祖先谱系从bactID文件中自动解析。"},
                    {"name":"fasta_dir","label":"FASTA Directory (-f)","label_cn":"FASTA目录","type":"text","placeholder":"/path/to/genomes","required":True,"help":"Directory containing FASTA files (.fasta) for all strains referenced in the analysis. Both base and patch strain files must be present.","help_cn":"包含分析中所有参考菌株FASTA文件（.fasta扩展名）的目录。基础菌株和补丁菌株的文件都必须存在。"},
                    {"name":"bactid_file","label":"bactID.txt File (--bactid)","label_cn":"bactID.txt文件","type":"text","placeholder":"/path/to/bactID.txt","required":False,"help":"Path to the bactID.txt lineage map generated by BactAG. When provided, the ancestral lineage is automatically traced to determine the correct patch order.","help_cn":"BactAG生成的bactID.txt谱系映射表路径。提供后，将自动追溯基础菌株的祖先谱系以确定正确的补丁顺序。"},
                    {"name":"workers","label":"Parallel Workers (-w)","label_cn":"并行工作进程数","type":"number","placeholder":"8","required":False,"help":"Number of parallel Mauve alignment workers. Default: 8. Each worker processes one ancestor-descendant pair. Increase for multi-core servers.","help_cn":"并行Mauve比对工作进程数，默认8。每个进程处理一对祖先-后代的比对。多核服务器上可适当增大。"},
                ],
            },
            "bactevoltraj": {
                "title": "BactEvolTraj — Evolutionary Trajectory Analysis",
                "title_cn": "BactEvolTraj — 演化轨迹分析",
                "overview": "Analyzes large structural variants (insertions, deletions, inversions) along each branch of a user-defined evolutionary tree using three-way progressiveMauve alignment. For each SV event, affected genes are identified and annotated from the corresponding GenBank file. Outputs include event tables, coverage matrices, and publication-quality plots.",
                "overview_cn": "利用三向progressiveMauve比对分析沿进化树各分支的大片段结构变异（插入、缺失、倒位）。每个SV事件会从对应GenBank文件中识别并注释受影响的基因。输出包括事件表、覆盖度矩阵和发表级图表。",
                "form_fields": [
                    {"name":"root_node","label":"Root Node Name (--root-node)","label_cn":"根节点名称","type":"text","placeholder":"S.enterica_subsp.enterica_AOG","required":True,"help":"Name of the root ancestral genome. Must have corresponding .fasta and .gbk files. Serves as the reference coordinate system for all comparisons.","help_cn":"根祖先基因组名称。必须有对应的.fasta和.gbk文件，将作为所有比较的参考坐标系统。"},
                    {"name":"fasta_dir","label":"FASTA Directory (--fasta-dir)","label_cn":"FASTA目录","type":"text","placeholder":"/path/to/genomes","required":True,"help":"Directory containing FASTA genome files for all nodes in the evolutionary tree. Each node must have a corresponding .fasta file.","help_cn":"包含进化树中所有节点FASTA基因组文件的目录。每个节点必须有对应的.fasta文件。"},
                    {"name":"gbk_dir","label":"GenBank Directory (--gbk-dir)","label_cn":"GenBank目录","type":"text","placeholder":"/path/to/annotations","required":True,"help":"Directory containing GenBank annotation files for all nodes. Each node must have a corresponding .gbk file for gene-level annotation of SV events.","help_cn":"包含所有节点GenBank注释文件的目录。每个节点必须有对应的.gbk文件以进行SV事件的基因级注释。"},
                    {"name":"tree_edges","label":"Tree Edges JSON (--tree)","label_cn":"进化树边 (JSON格式)","type":"textarea","placeholder":'[\n  ["root_AOG", "node_mA"],\n  ["node_mA", "node_mB"],\n  ["node_mB", "node_mC"]\n]',"required":True,"help":"List of (parent, child) tuples in JSON format defining the evolutionary tree topology. Each tuple represents one branch; SV events are analyzed independently for each branch.","help_cn":"JSON格式的（父节点, 子节点）元组列表，定义进化树拓扑结构。每个元组代表一个分支，SV事件针对每个分支独立分析。"},
                    {"name":"min_event_length","label":"Min Event Length bp (--min-len)","label_cn":"最小事件长度 (bp)","type":"number","placeholder":"1000","required":False,"help":"Minimum length threshold for SV events in base pairs. Default: 1000. Events shorter than this are excluded to reduce noise from small indels.","help_cn":"结构变异事件的最小长度阈值（碱基对），默认1000。短于此阈值的事件将被排除以减少小片段indel的噪声。"},
                ],
            },
            "bactfragann": {
                "title": "BactFragAnn — Fragment Annotation & Visualization",
                "title_cn": "BactFragAnn — 片段注释与可视化",
                "overview": "Generates interactive HTML visualization dashboards using Plotly. Two modes: (a) 1dgr — mosaic charts showing fragment-to-gene mappings from Bact1DGR output; (b) evoltraj — circular plots displaying structural variant events from BactEvolTraj output with clickable gene annotations. Both modes integrate seamlessly with upstream modules.",
                "overview_cn": "使用Plotly生成交互式HTML可视化仪表板。两种模式：(a) 1dgr — 展示Bact1DGR片段-基因映射的马赛克图；(b) evoltraj — 展示BactEvolTraj结构变异事件的环形图，支持点击查看基因注释。两种模式均可与上游模块无缝联动。",
                "form_fields": [
                    {"name":"vis_mode","label":"Visualization Mode (--mode)","label_cn":"可视化模式","type":"select","options":["1dgr","evoltraj"],"required":True,"help":"1dgr: Mosaic charts for 1D genomic representation fragments (input from Bact1DGR). evoltraj: Circular SV event plots (input from BactEvolTraj).","help_cn":"1dgr: 一维基因组表示片段的马赛克图（输入来自Bact1DGR）。evoltraj: 演化轨迹分析的结构变异环形图（输入来自BactEvolTraj）。"},
                    {"name":"base_dir","label":"Working Directory (--base-dir)","label_cn":"工作目录","type":"text","placeholder":"/path/to/working_dir","required":True,"help":"Base working directory containing the input data subfolders. Point this to your Bact1DGR or BactEvolTraj results directory.","help_cn":"包含输入数据子文件夹的基础工作目录。请指向Bact1DGR或BactEvolTraj的结果目录。"},
                    {"name":"output_folder","label":"Output Folder","label_cn":"输出文件夹","type":"text","placeholder":"Mosaic_Charts","required":False,"help":"Name of the output folder (created under base_dir). Default: 'Mosaic_Charts'.","help_cn":"输出文件夹名称（在base_dir下创建）。默认：'Mosaic_Charts'。"},
                    {"name":"txt_folder","label":"1DGR Text Folder (1dgr mode)","label_cn":"1DGR文本文件夹","type":"text","placeholder":"1DGR_en","required":False,"help":"Name of the subfolder (under base_dir) containing 1DGR output .txt files. Default: '1DGR_en'. Only needed for 1dgr mode.","help_cn":"base_dir下包含1DGR输出.txt文件的子文件夹名称。默认：'1DGR_en'。仅1dgr模式需要。"},
                    {"name":"gbk_folder","label":"GenBank Folder","label_cn":"GenBank文件夹","type":"text","placeholder":"GBK_en","required":False,"help":"Name of the subfolder (under base_dir) containing GenBank .gbk annotation files. Default: 'GBK_en'. Required for gene-level annotation in both modes.","help_cn":"base_dir下包含GenBank .gbk注释文件的子文件夹名称。默认：'GBK_en'。两种模式的基因级注释都需要。"},
                    {"name":"sv_file","label":"SV Event File (evoltraj mode)","label_cn":"SV事件文件","type":"text","placeholder":"Large_SV_Details.csv","required":False,"help":"Path to the SV event details file (CSV or tab-separated). Leave empty to auto-detect. Only needed for evoltraj mode.","help_cn":"SV事件详情文件路径（CSV或Tab分隔）。留空则自动检测。仅evoltraj模式需要。"},
                ],
            },
        }

        info = module_info.get(module_id)
        if not info:
            flash(f"Unknown module: {module_id}", "error")
            return redirect(url_for("index"))
        return render_template("module.html", module_id=module_id, info=info)

    @app.route("/api/run", methods=["POST"])
    def api_run():
        data = request.get_json()
        module_id = data.get("module")
        params = data.get("params", {})
        job_id = str(uuid.uuid4())[:8]
        jobs[job_id] = {"id":job_id,"module":module_id,"status":"queued","started":datetime.now().isoformat(),"output":"","error":""}
        thread = threading.Thread(target=_run_module_job, args=(job_id, module_id, params))
        thread.daemon = True
        thread.start()
        return jsonify({"job_id":job_id,"status":"queued"})

    @app.route("/api/job/<job_id>")
    def api_job_status(job_id):
        job = jobs.get(job_id)
        if not job:
            return jsonify({"error":"Job not found"}), 404
        return jsonify(job)

    def _run_module_job(job_id, module_id, params):
        job = jobs[job_id]
        job["status"] = "running"
        try:
            from genomevo.modules import run_bactag, run_bactcg, run_bactpg, run_bactpga, run_bact1dgr, run_bactevoltraj, run_bactfragann

            module_funcs = {
                "bactag": run_bactag, "bactcg": run_bactcg, "bactpg": run_bactpg,
                "bactpga": lambda **kw: run_bactpga(mode="pipeline", **kw),
                "bact1dgr": run_bact1dgr, "bactevoltraj": run_bactevoltraj, "bactfragann": run_bactfragann,
            }
            func = module_funcs.get(module_id)
            if not func:
                raise ValueError(f"Unknown module: {module_id}")

            param_map = {
                "bactag":{"tree_dir":"tree_dir","gene_dir":"gene_dir","threads":"threads","id_file":"id_file","output_dir":"output_dir"},
                "bactcg":{"input_dir":"input_dir","output_dir":"output_dir","ref_strain":"ref_strain","cd_cutoff":"cd_cutoff","cd_s":"cd_s","cg1":"cg1_cutoff","cg2":"cg2","skip_qc":"skip_qc"},
                "bactpg":{"seq_dir":"seq_dir","output_dir":"output_dir","similarity":"similarity","threads":"threads"},
                "bactpga":{"gbk_file":"gbk_file","pg_file":"pg_file","seq_dir":"seq_dir","strain":"strain_name","mutbest_dir":"mutbest_dir","cov1":"cov1","cov2":"cov2"},
                "bact1dgr":{"base_strain":"base_strain","fasta_dir":"fasta_dir","bactid_file":"bactid_file","workers":"workers","patches":"patches"},
                "bactevoltraj":{"root_node":"root_node","fasta_dir":"fasta_dir","gbk_dir":"gbk_dir","min_event_length":"min_event_length"},
                "bactfragann":{"base_dir":"base_dir","mode":"vis_mode","txt_folder":"txt_folder","gbk_folder":"gbk_folder","output_folder":"output_folder","sv_file":"sv_file"},
            }
            mapping = param_map.get(module_id, {})
            func_kwargs = {}
            for wk, fk in mapping.items():
                val = params.get(wk)
                if val is not None and val != "":
                    if fk in ("threads","workers","min_event_length"): val = int(val)
                    elif fk in ("cd_cutoff","cd_s","cg1_cutoff","cg2","similarity","cov1","cov2"): val = float(val)
                    elif fk == "skip_qc": val = True if val == "on" else False
                    elif fk == "patches": val = [p.strip() for p in val.split(",") if p.strip()]
                    func_kwargs[fk] = val

            if module_id == "bactevoltraj" and "tree_edges" in params:
                func_kwargs["tree_edges"] = json.loads(params["tree_edges"])
            if module_id == "bactpga":
                mv = params.get("mode","pipeline")
                if mv == "parse":
                    result = run_bactpga(mode="parse", gbk_file=func_kwargs.get("gbk_file"))
                    job["status"]="completed"; job["output"]=str(result); return
                elif mv == "annotate":
                    func_kwargs["mode"] = "annotate"

            result = func(**func_kwargs)
            job["status"] = "completed"
            job["output"] = str(result)
        except Exception as e:
            job["status"] = "failed"
            job["error"] = str(e)

    return app
