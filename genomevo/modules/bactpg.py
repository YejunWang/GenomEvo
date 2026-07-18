"""
BactPG module wrapper – Pan-Genome Analysis.

Inputs:
  - seq_dir:   Directory containing sequence files (.fasta for protein)
  - output_dir: Output directory for pan-genome results
  - similarity: Similarity threshold for clustering (default: 0.7)
  - threads:   Number of parallel threads (default: 30)

Outputs:
  - PG.txt: Pan-genome matrix (in output_dir/result/PG.txt)
  - Clustering and iteration data under output_dir/result/

External dependencies: cd-hit, blastp, makeblastdb

Pipeline steps:
  1. CD-HIT clustering of each strain's proteome
  2. Batch partitioning by cumulative sequence count
  3. All-vs-all BLASTP within each batch
  4. Orthology filtering and PG temporary matrix construction
  5. Iterative merging across batches
  6. Final PG matrix generation (PG.txt)
"""

import os
import shutil
import subprocess
from ..config import BACTPG_BIN, logger


def run_bactpg(
    seq_dir: str,
    output_dir: str = "BactPG_Results",
    similarity: float = 0.7,
    threads: int = 30,
    dry_run: bool = False,
) -> str:
    """
    Run pan-genome analysis using BactPG.

    Parameters
    ----------
    seq_dir : str
        Directory containing sequence FASTA files (.fasta extension required).
    output_dir : str
        Output directory for pan-genome results.
        BactPG creates result/ subdirectory inside this directory.
    similarity : float
        Sequence similarity threshold for CD-HIT and BLAST filtering (default: 0.7).
    threads : int
        Number of parallel threads for CD-HIT and BLASTP (default: 30).
    dry_run : bool
        If True, print command without executing.

    Returns
    -------
    str : Path to the output directory containing result/PG.txt.
    """
    if not os.path.isdir(seq_dir):
        raise FileNotFoundError(f"Sequence directory not found: {seq_dir}")

    os.makedirs(output_dir, exist_ok=True)

    # BactPG runs from within output_dir; it writes to ./result/ inside cwd.
    # The --yes flag enables non-interactive auto-deletion of existing result folder.
    cmd = [
        BACTPG_BIN,
        os.path.abspath(seq_dir),
        str(similarity),
        str(threads),
        "--yes",
    ]

    logger.info("Starting BactPG pan-genome analysis")
    logger.info("  Sequence dir : %s", seq_dir)
    logger.info("  Output dir   : %s", output_dir)
    logger.info("  Similarity   : %s", similarity)
    logger.info("  Threads      : %s", threads)

    if dry_run:
        logger.info("[DRY RUN] %s", " ".join(cmd))
        return output_dir

    # Run BactPG from within output_dir
    logger.info("Executing: %s", " ".join(cmd))
    result = subprocess.run(
        cmd,
        cwd=output_dir,
        capture_output=False,
    )

    if result.returncode != 0:
        raise RuntimeError(
            f"BactPG exited with code {result.returncode}. "
            f"Check logs in {output_dir}/result/"
        )

    pg_txt = os.path.join(output_dir, "result", "PG.txt")
    if os.path.isfile(pg_txt):
        logger.info("BactPG completed successfully. PG.txt: %s", pg_txt)
    else:
        logger.warning("BactPG completed but PG.txt not found at %s", pg_txt)

    return os.path.abspath(output_dir)
