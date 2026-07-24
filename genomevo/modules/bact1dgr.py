"""
Bact1DGR module wrapper – 1D Genomic Representation.

Generates 1D linear representations comparing ancestral to descendant genomes.

Inputs:
  - base_strain:  Base/ancestral strain name (must match .fasta filename)
  - patches:      List of patch/descendant strain names
  - fasta_dir:    Directory containing FASTA files for all strains
  - bactid_file:  Path to bactID.txt (from BactAG) for ancestry resolution

Outputs:
  - *.1dgr.txt files in output_dir/Final_Results/

External dependencies: progressiveMauve (must be in $PATH)
"""

import os
import sys
import subprocess
from ..config import BIN_DIR, BACT1DGR_DEFAULT_WORKERS, logger


def run_bact1dgr(
    base_strain: str,
    patches: list = None,
    fasta_dir: str = ".",
    bactid_file: str = None,
    output_dir: str = "OneDGR_Output",
    workers: int = BACT1DGR_DEFAULT_WORKERS,
    dry_run: bool = False,
) -> str:
    """
    Run 1D genomic representation analysis.

    Parameters
    ----------
    base_strain : str
        Name of the base/ancestral strain (must match .fasta file).
    patches : list, optional
        List of patch strain names. If None, resolved from bactid_file.
    fasta_dir : str
        Directory containing .fasta files.
    bactid_file : str, optional
        Path to bactID.txt for ancestry resolution.
    output_dir : str
        Output directory.
    workers : int
        Number of parallel workers (default: 8).
    dry_run : bool

    Returns
    -------
    str : Path to the final output file (*.1DGR.txt).
    """
    import shutil
    if shutil.which("progressiveMauve") is None:
        raise RuntimeError("progressiveMauve is required but not found in $PATH.")

    # Build the command using the onedgr main_cli
    # We invoke the Python module directly to avoid installation issues
    module_dir = os.path.dirname(os.path.abspath(__file__))
    onedgr_dir = os.path.join(module_dir, "onedgr")

    # Add to sys.path so that onedgr package is importable
    if onedgr_dir not in sys.path:
        sys.path.insert(0, os.path.dirname(onedgr_dir))

    from onedgr.main import main_cli
    from onedgr.pipeline import run_full_pipeline
    from onedgr.utils import parse_bactid, get_bactid_ancestors

    # Determine patches
    final_patches = []
    if bactid_file:
        logger.info(f"Reading patch order from {bactid_file}...")
        ancestral_patches = get_bactid_ancestors(bactid_file, base_strain)
        if ancestral_patches:
            final_patches = ancestral_patches
            logger.info(f"Found {len(final_patches)} ancestral nodes")
        else:
            # Fallback
            bactid_patches = parse_bactid(bactid_file)
            if base_strain in bactid_patches:
                base_idx = bactid_patches.index(base_strain)
                final_patches = list(reversed(bactid_patches[base_idx + 1:]))

    if patches:
        if final_patches:
            selected = set(patches)
            final_patches = [p for p in final_patches if p in selected]
        else:
            final_patches = list(patches)

    if not final_patches:
        raise ValueError(
            "No patches determined. Provide --patches or --bactid_file."
        )

    # Validate FASTA files
    base_fasta = os.path.join(fasta_dir, f"{base_strain}.fasta")
    if not os.path.isfile(base_fasta):
        raise FileNotFoundError(f"Base FASTA not found: {base_fasta}")

    for p in final_patches:
        pf = os.path.join(fasta_dir, f"{p}.fasta")
        if not os.path.isfile(pf):
            raise FileNotFoundError(f"Patch FASTA not found: {pf}")

    logger.info(f"Starting OneDGR pipeline")
    logger.info(f"  Base strain: {base_strain}")
    logger.info(f"  Patches    : {final_patches}")
    logger.info(f"  Workers    : {workers}")

    if dry_run:
        logger.info("[DRY RUN] Would run OneDGR pipeline")
        return os.path.join(output_dir, "Final_Results", f"{base_strain}.1DGR.txt")

    os.makedirs(output_dir, exist_ok=True)

    # Set up binary paths for onedgr (so it doesn't try to build Go from source)
    # Monkey-patch the binary paths
    bin_dir = BIN_DIR
    onedgr_src = os.path.join(module_dir, "onedgr_src", "go")
    import onedgr.go_builder as go_builder_mod

    # Pre-build the Go binaries if needed
    go_builder_mod.build_go_binaries(onedgr_src, bin_dir)

    run_full_pipeline(base_strain, final_patches, output_dir, fasta_dir, max_workers=workers)

    final_output = os.path.join(output_dir, "Final_Results", f"{base_strain}.1DGR.txt")
    logger.info(f"OneDGR pipeline completed. Output: {final_output}")
    return final_output
