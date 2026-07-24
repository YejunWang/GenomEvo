"""
BactAG module wrapper – Ancestral Genome Reconstruction.

Inputs:
  - tree_dir:   Directory containing the phylogenetic tree file (Newick format)
  - gene_dir:   Directory containing genome FASTA files (.fasta)
  - id_file:    Path to write the bactID.txt output file
  - output_dir: Working/output directory (default: "BactAG_Results")

Outputs:
  - {output_dir}/output/  : Reconstructed AOGs in FASTA format
  - {output_dir}/bactID.txt : Lineage map
  - {output_dir}/log/     : Processing logs

External dependencies: progressiveMauve (must be in $PATH)
"""

import os
import sys
import subprocess
from ..config import BACTAG_BIN, BACTAG_DEFAULT_THREADS, logger


def run_bactag(
    tree_dir: str,
    gene_dir: str,
    id_file: str = "bactID.txt",
    threads: int = BACTAG_DEFAULT_THREADS,
    output_dir: str = "BactAG_Results",
    dry_run: bool = False,
) -> str:
    """
    Run ancestral genome reconstruction using BactAG.

    Parameters
    ----------
    tree_dir : str
        Directory containing exactly one Newick-format tree file.
    gene_dir : str
        Directory containing input genome FASTA files (one per strain).
    id_file : str
        Path where bactID.txt will be written (default: "bactID.txt").
    threads : int
        Number of parallel threads (default: 20).
    output_dir : str
        Directory name for final output bundle (default: "BactAG_Results").
    dry_run : bool
        If True, print command without executing.

    Returns
    -------
    str : Path to the output directory.
    """
    # Validate inputs
    if not os.path.isdir(tree_dir):
        raise FileNotFoundError(f"Tree directory not found: {tree_dir}")
    if not os.path.isdir(gene_dir):
        raise FileNotFoundError(f"Gene directory not found: {gene_dir}")

    fasta_files = [f for f in os.listdir(gene_dir) if f.endswith((".fasta", ".fa", ".fna"))]
    if not fasta_files:
        raise FileNotFoundError(f"No .fasta files found in {gene_dir}")

    tree_files = os.listdir(tree_dir)
    if len(tree_files) != 1:
        logger.warning(f"Tree directory should contain exactly 1 file, found {len(tree_files)}")

    # Ensure progressiveMauve is available
    import shutil
    if shutil.which("progressiveMauve") is None:
        raise RuntimeError(
            "progressiveMauve is required but not found in $PATH. "
            "Install it via: sudo apt install mauve-aligner"
        )

    logger.info(f"Starting BactAG with {threads} threads")
    logger.info(f"  Tree dir : {tree_dir}")
    logger.info(f"  Gene dir : {gene_dir}")
    logger.info(f"  ID file  : {id_file}")

    cmd = [
        BACTAG_BIN,
        "-t", str(threads),
        "-tree", tree_dir,
        "-gene", gene_dir,
        "-id", id_file,
    ]

    if dry_run:
        logger.info(f"[DRY RUN] {' '.join(cmd)}")
        return output_dir

    logger.info(f"Executing: {' '.join(cmd)}")

    # Run BactAG inside the designated output directory as working directory
    os.makedirs(output_dir, exist_ok=True)
    result = subprocess.run(cmd, capture_output=False, cwd=output_dir)

    if result.returncode != 0:
        raise RuntimeError(f"BactAG exited with code {result.returncode}")

    logger.info("BactAG completed successfully.")
    logger.info(f"  AOGs       : {os.path.join(output_dir, 'output')}")
    logger.info(f"  Lineage map: {os.path.join(output_dir, id_file)}")
    logger.info(f"  Logs       : {os.path.join(output_dir, 'log')}")
    return os.path.abspath(output_dir)
