"""
BactPGA module wrapper – Pan-Genome Annotation.

Inputs:
  - gbk_file:     Input GenBank file (.gbk)
  - pg_file:      Pan-genome matrix file (PG.txt)
  - seq_dir:      Directory of sequence files (for pipeline mode)
  - mutbest_dir:  Directory containing mutual best hit files (for annotate mode)
  - strain_name:  Strain name for annotation

Outputs:
  - Annotated gene table (.tab.txt)

Subcommands:
  - parse:    Extract features from GenBank -> .tab.txt
  - annotate: Map PG UIDs to gene table -> annotated .tab.txt
  - pipeline: Full auto pipeline (parse + annotate)
"""

import os
import subprocess
from ..config import BACTPGA_BIN, BACTCG_BIN, logger


def run_bactpga_parse(
    gbk_file: str,
    output_file: str = None,
    dry_run: bool = False,
) -> str:
    """
    Parse a GenBank file to extract gene features.

    Parameters
    ----------
    gbk_file : str
        Path to input GenBank file.
    output_file : str, optional
        Path to write the parsed table. If None, prints to stdout.
    dry_run : bool

    Returns
    -------
    str : Parsed table content (if output_file is None) or path to output_file.
    """
    if not os.path.isfile(gbk_file):
        raise FileNotFoundError(f"GenBank file not found: {gbk_file}")

    cmd = [BACTPGA_BIN, "parse", os.path.abspath(gbk_file)]

    logger.info(f"BactPGA parse: {gbk_file}")

    if dry_run:
        logger.info(f"[DRY RUN] {' '.join(cmd)}")
        return output_file or ""

    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        raise RuntimeError(f"BactPGA parse failed: {result.stderr}")

    if output_file:
        with open(output_file, "w") as f:
            f.write(result.stdout)
        logger.info(f"Parsed table written to: {output_file}")
        return output_file
    else:
        return result.stdout


def run_bactpga_annotate(
    gbk_file: str,
    pg_file: str,
    mutbest_dir: str,
    strain_name: str,
    output_file: str = None,
    mode: str = "PGAG",
    seq_dir: str = None,
    dry_run: bool = False,
) -> str:
    """
    Annotate gene table with pan-genome UIDs.

    Parameters
    ----------
    gbk_file : str
        Path to GenBank file.
    pg_file : str
        Path to pan-genome matrix (PG.txt).
    mutbest_dir : str
        Directory containing mutual best hit files.
    strain_name : str
        Strain name for annotation.
    output_file : str, optional
        Path to write the annotated table.
    mode : str
        Annotation mode: "PGAG" or "RAST" (default: "PGAG").
    seq_dir : str, optional
        Sequence directory for annotation.
    dry_run : bool

    Returns
    -------
    str : Annotated table content or path to output_file.
    """
    # First parse the GBK to get the gene table
    parsed = run_bactpga_parse(gbk_file, output_file=None, dry_run=dry_run)
    if dry_run:
        return output_file or ""

    # Write parsed to a temp file
    import tempfile
    with tempfile.NamedTemporaryFile(mode="w", suffix=".tab.txt", delete=False) as tmp:
        tmp.write(parsed)
        tab_file = tmp.name

    cmd = [
        BACTPGA_BIN, "annotate",
        "-pg", os.path.abspath(pg_file),
        "-tab", tab_file,
        "-mode", mode,
        "-strain", strain_name,
        "-mutbestDir", os.path.abspath(mutbest_dir),
    ]
    if seq_dir:
        cmd.extend(["-seqDir", os.path.abspath(seq_dir)])

    logger.info(f"BactPGA annotate: strain={strain_name}, mode={mode}")

    if dry_run:
        logger.info(f"[DRY RUN] {' '.join(cmd)}")
        return output_file or ""

    result = subprocess.run(cmd, capture_output=True, text=True)
    os.unlink(tab_file)

    if result.returncode != 0:
        raise RuntimeError(f"BactPGA annotate failed: {result.stderr}")

    if output_file:
        with open(output_file, "w") as f:
            f.write(result.stdout)
        logger.info(f"Annotated table written to: {output_file}")
        return output_file
    else:
        return result.stdout


def run_bactpga_pipeline(
    gbk_file: str,
    pg_file: str,
    seq_dir: str,
    output_dir: str,
    cov1: float = 0.7,
    cov2: float = 0.7,
    cg_path: str = None,
    annot_mode: str = "PGAG",
    dry_run: bool = False,
) -> str:
    """
    Run the full BactPGA pipeline (parse + annotate) automatically.

    Parameters
    ----------
    gbk_file : str
        Path to GenBank file.
    pg_file : str
        Path to pan-genome matrix.
    seq_dir : str
        Directory of sequence files.
    output_dir : str
        Output directory.
    cov1 : float
        Coverage threshold 1 (default: 0.7).
    cov2 : float
        Coverage threshold 2 (default: 0.7).
    cg_path : str, optional
        Path to CG executable (default: auto-detected).
    annot_mode : str
        Annotation mode: "PGAG" or "RAST" (default: "PGAG").
    dry_run : bool

    Returns
    -------
    str : Path to output directory.
    """
    if not os.path.isfile(gbk_file):
        raise FileNotFoundError(f"GenBank file not found: {gbk_file}")
    if not os.path.isfile(pg_file):
        raise FileNotFoundError(f"PG file not found: {pg_file}")

    os.makedirs(output_dir, exist_ok=True)

    cmd = [
        BACTPGA_BIN, "pipeline",
        "-gbk", os.path.abspath(gbk_file),
        "-pg", os.path.abspath(pg_file),
        "-seq", os.path.abspath(seq_dir),
        "-cov1", str(cov1),
        "-cov2", str(cov2),
        "-mode", annot_mode,
        "-out", os.path.abspath(output_dir),
    ]
    # Use the new bactcg (BactCG 2.0) binary by default
    cg_bin = cg_path if cg_path else os.path.join(os.path.dirname(BACTPGA_BIN), "bactcg")
    if os.path.isfile(cg_bin):
        cmd.extend(["-cg", os.path.abspath(cg_bin)])

    logger.info(f"BactPGA pipeline: gbk={gbk_file} pg={pg_file}")

    if dry_run:
        logger.info(f"[DRY RUN] {' '.join(cmd)}")
        return output_dir

    result = subprocess.run(cmd, capture_output=False)
    if result.returncode != 0:
        raise RuntimeError(f"BactPGA pipeline failed with code {result.returncode}")

    logger.info("BactPGA pipeline completed successfully.")
    return os.path.abspath(output_dir)


def run_bactpga(
    mode: str = "pipeline",
    gbk_file: str = None,
    pg_file: str = None,
    seq_dir: str = None,
    output_dir: str = "BactPGA_Results",
    mutbest_dir: str = None,
    strain_name: str = None,
    output_file: str = None,
    cov1: float = 0.7,
    cov2: float = 0.7,
    cg_path: str = None,
    annot_mode: str = "PGAG",
    dry_run: bool = False,
):
    """
    Unified BactPGA dispatch.

    Parameters
    ----------
    mode : str
        "parse" | "annotate" | "pipeline"
    gbk_file : str
        Path to input GenBank file (.gbk).
    pg_file : str
        Path to pan-genome matrix (PG.txt).
    seq_dir : str
        Directory of sequence files (pipeline mode).
    output_dir : str
        Output directory for results.
    mutbest_dir : str
        Directory containing mutual best hit files (annotate mode).
    strain_name : str
        Strain name for annotation.
    output_file : str, optional
        Output file path (parse/annotate mode).
    cov1 : float
        Coverage threshold 1 for pipeline (default: 0.7).
    cov2 : float
        Coverage threshold 2 for pipeline (default: 0.7).
    cg_path : str, optional
        Path to CG executable (pipeline mode).
    annot_mode : str
        Annotation mode: "PGAG" or "RAST" (default: "PGAG").
    dry_run : bool
        If True, print commands without executing.
    """
    if mode == "parse":
        return run_bactpga_parse(gbk_file, output_file, dry_run)
    elif mode == "annotate":
        return run_bactpga_annotate(gbk_file, pg_file, mutbest_dir, strain_name, output_file, mode=annot_mode, seq_dir=seq_dir, dry_run=dry_run)
    elif mode == "pipeline":
        return run_bactpga_pipeline(gbk_file, pg_file, seq_dir, output_dir, cov1=cov1, cov2=cov2, cg_path=cg_path, annot_mode=annot_mode, dry_run=dry_run)
    else:
        raise ValueError(f"Unknown BactPGA mode: {mode}")
