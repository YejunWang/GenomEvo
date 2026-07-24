"""
BactCG module wrapper – Core Genome Analysis.

Complete pipeline:
  1. QC filtering (optional) → CD-HIT clustering → core gene orthology → CG.tab.txt
  2. Extract per-family FASTA files (getfa)
  3. Multiple sequence alignment with clustalw2 → MEGA format (clustal)
  4. SNP concatenation → all_core_gene.meg (snpmega)

Inputs:
  - input_dir:    Directory containing protein FASTA files (.fasta)
  - output_dir:   Output directory for results
  - ref_strain:   Reference strain ID (used as anchor)

Outputs:
  - CG_ALL.txt:                   Core gene presence-absence table
  - all-strain-together/2.result/: Per-family FASTA files
  - all-strain-together/3.mega/:  MEG-format multiple sequence alignments
  - all-strain-together/4.SNP_mega/all_core_gene.meg: SNP concatenated alignment

External dependencies: blastp, cd-hit, clustalw2
"""

import os
import subprocess
from ..config import BACTCG_BIN, logger


def _run_step(cmd, log_file, stdin_input=None):
    """Run a subprocess with output redirected to a log file (avoids pipe deadlock)."""
    logger.info(f"Executing: {' '.join(cmd)}")
    with open(log_file, "w") as f:
        result = subprocess.run(
            cmd,
            input=stdin_input,
            stdout=f,
            stderr=subprocess.STDOUT,
            text=True,
        )
    if result.returncode != 0:
        with open(log_file) as f:
            tail = "".join(f.readlines()[-10:])
        raise RuntimeError(
            f"Command exited with code {result.returncode}\n"
            f"  Command: {' '.join(cmd)}\n"
            f"  Log tail: {tail}"
        )
    logger.info(f"  -> Completed (log: {log_file})")


def run_bactcg(
    input_dir: str,
    output_dir: str,
    ref_strain: str,
    cd_cutoff: float = 0.7,
    cd_s: float = 0.7,
    cg1_cutoff: float = 0.8,
    cg2: float = 0.9,
    skip_qc: bool = False,
    dry_run: bool = False,
) -> str:
    """
    Run complete core genome analysis using BactCG (4-step pipeline).

    Parameters
    ----------
    input_dir : str
        Directory containing protein FASTA files (one per strain).
    output_dir : str
        Output directory for core genome results.
    ref_strain : str
        Reference strain name (e.g., "BMU_04865").
    cd_cutoff : float
        CD-HIT sequence identity threshold (default: 0.7).
    cd_s : float
        CD-HIT length difference cutoff (default: 0.7).
    cg1_cutoff : float
        BactCG first orthology cutoff (default: 0.8).
    cg2 : float
        BactCG second parameter (default: 0.9).
    skip_qc : bool
        Skip the QC filtering step (default: False).
    dry_run : bool
        If True, print commands without executing.

    Returns
    -------
    str : Path to the output directory.
    """
    if not os.path.isdir(input_dir):
        raise FileNotFoundError(f"Input directory not found: {input_dir}")

    os.makedirs(output_dir, exist_ok=True)
    stdin_input = "n\n" if skip_qc else "y\n"

    # Paths for subsequent steps
    abs_output = os.path.abspath(output_dir)
    cg_all_txt = os.path.join(abs_output, "CG_ALL.txt")
    cdhit_fatsa = os.path.join(abs_output, "1.cd-hit_output", "1.cd-hit_fatsa")
    all_strain_dir = os.path.join(abs_output, "all-strain-together")
    fa_dir = os.path.join(all_strain_dir, "2.result")
    mega_dir = os.path.join(all_strain_dir, "3.mega")
    snp_dir = os.path.join(all_strain_dir, "4.SNP_mega")

    # =========================================================================
    # Step 1: QC → CD-HIT → Core genome orthology (bactcg run)
    # =========================================================================
    cmd_run = [
        BACTCG_BIN, "run",
        "--ref", ref_strain,
        "--cd-c", str(cd_cutoff),
        "--cd-s", str(cd_s),
        "--cg1", str(cg1_cutoff),
        "--cg2", str(cg2),
        "-i", input_dir,
        "-o", output_dir,
    ]

    logger.info("=== BactCG Step 1/4: QC → CD-HIT → Core genome orthology ===")
    logger.info(f"  Input      : {input_dir}")
    logger.info(f"  Output     : {output_dir}")
    logger.info(f"  Reference  : {ref_strain}")
    logger.info(f"  CD cutoff  : {cd_cutoff}")
    logger.info(f"  CG1 cutoff : {cg1_cutoff}")
    logger.info(f"  Skip QC    : {skip_qc}")

    if dry_run:
        logger.info(f"[DRY RUN] {' '.join(cmd_run)}")
        return output_dir

    _run_step(cmd_run, os.path.join(abs_output, "step1_run.log"), stdin_input=stdin_input)

    if not os.path.isfile(cg_all_txt):
        raise RuntimeError(f"Expected output not found: {cg_all_txt}")

    # =========================================================================
    # Step 2: Extract per-family FASTA files (bactcg getfa)
    # =========================================================================
    cmd_getfa = [
        BACTCG_BIN, "getfa",
        "-c", cdhit_fatsa,
        "-g", cg_all_txt,
        "-o", all_strain_dir,
    ]

    logger.info("=== BactCG Step 2/4: Extract per-family FASTA (getfa) ===")
    _run_step(cmd_getfa, os.path.join(abs_output, "step2_getfa.log"))

    # =========================================================================
    # Step 3: Clustalw2 MSA → MEGA format (bactcg clustal)
    # =========================================================================
    cmd_clustal = [
        BACTCG_BIN, "clustal",
        "-d", fa_dir,
        "--gcgdir", os.path.join(all_strain_dir, "3.gcg"),
        "--megadir", mega_dir,
    ]

    logger.info("=== BactCG Step 3/4: Clustalw2 alignment → MEGA (clustal) ===")
    _run_step(cmd_clustal, os.path.join(abs_output, "step3_clustal.log"))

    # =========================================================================
    # Step 4: SNP concatenation (bactcg snpmega)
    # =========================================================================
    cmd_snpmega = [
        BACTCG_BIN, "snpmega",
        "-i", mega_dir,
        "-o", snp_dir,
    ]

    logger.info("=== BactCG Step 4/4: SNP concatenation (snpmega) ===")
    _run_step(cmd_snpmega, os.path.join(abs_output, "step4_snpmega.log"))

    logger.info("BactCG complete pipeline finished successfully.")
    logger.info(f"  Core gene table: {cg_all_txt}")
    logger.info(f"  Per-family FASTA: {fa_dir}")
    logger.info(f"  MEGA alignments: {mega_dir}")
    logger.info(f"  SNP MEGA: {os.path.join(snp_dir, 'all_core_gene.meg')}")

    return abs_output
