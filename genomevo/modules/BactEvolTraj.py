"""
BactEvolTraj module wrapper – Evolutionary Trajectory Analysis.

Analyzes large structural variants (SV) along evolutionary tree branches
using 3-way progressiveMauve alignments.

Inputs:
  - root_node:    Root ancestral genome name (must match .fasta and .gbk)
  - fasta_dir:    Directory containing FASTA files
  - gbk_dir:      Directory containing GenBank files
  - tree_edges:   List of (parent, child) tuples defining the evolutionary tree
  - output_dir:   Output directory for SV analysis results

Outputs:
  - CSV tables of SV events per branch
  - Coverage matrices
  - Visualization plots (Matplotlib)

External dependencies: progressiveMauve, BioPython, pandas, matplotlib, numpy
"""

import os
import sys
import subprocess
from ..config import BACTEVOLTRAJ_MIN_EVENT_LENGTH, logger


def run_bactevoltraj(
    root_node: str,
    fasta_dir: str,
    gbk_dir: str,
    tree_edges: list,
    output_dir: str = "Final_Large_SV_Analysis",
    min_event_length: int = BACTEVOLTRAJ_MIN_EVENT_LENGTH,
    dry_run: bool = False,
) -> str:
    """
    Run evolutionary trajectory analysis.

    Parameters
    ----------
    root_node : str
        Name of the root ancestral genome (e.g., "S.enterica_subsp.enterica_AOG").
    fasta_dir : str
        Directory containing FASTA files for all nodes.
    gbk_dir : str
        Directory containing GenBank annotation files for all nodes.
    tree_edges : list of (str, str)
        List of (parent, child) tuples defining evolutionary relationships.
        Example: [("root", "mA"), ("mA", "mB"), ("mB", "mC")]
    output_dir : str
        Output directory.
    min_event_length : int
        Minimum SV event length in bp (default: 1000).
    dry_run : bool

    Returns
    -------
    str : Path to output directory.
    """
    import shutil
    if shutil.which("progressiveMauve") is None:
        raise RuntimeError("progressiveMauve is required but not found in $PATH.")

    os.makedirs(output_dir, exist_ok=True)

    # Build the analysis configuration
    module_dir = os.path.dirname(os.path.abspath(__file__))
    evolt_traj_script = os.path.join(module_dir, "BactEvolTraj.py")

    logger.info(f"Starting BactEvolTraj analysis")
    logger.info(f"  Root node  : {root_node}")
    logger.info(f"  Tree edges : {tree_edges}")
    logger.info(f"  Min event  : {min_event_length} bp")

    if dry_run:
        logger.info(f"[DRY RUN] Would run BactEvolTraj with above config")
        return output_dir

    # Dynamically execute the analysis
    # We patch the module-level variables and run
    import importlib.util
    spec = importlib.util.spec_from_file_location("BactEvolTraj", evolt_traj_script)
    bte_module = importlib.util.module_from_spec(spec)

    # Set configuration BEFORE execution
    bte_module.ROOT_NODE = root_node
    bte_module.ROOT_FASTA = os.path.join(fasta_dir, f"{root_node}.fasta")
    bte_module.EVOLUTION_TREE = tree_edges
    bte_module.OUTPUT_DIR = output_dir
    bte_module.MIN_EVENT_LENGTH = min_event_length

    spec.loader.exec_module(bte_module)

    logger.info(f"BactEvolTraj completed. Results in: {output_dir}")
    return os.path.abspath(output_dir)
