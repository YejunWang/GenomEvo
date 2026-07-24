"""
Global configuration for GenomEvo.

All paths, default parameters, and external tool requirements
are centralized here for easy maintenance.
"""

import os
import sys
import logging

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------

# Root of the genomevo package
PACKAGE_ROOT = os.path.dirname(os.path.abspath(__file__))

# Directory containing compiled Go binaries
BIN_DIR = os.path.join(PACKAGE_ROOT, "bin")

# External tool paths – resolved relative to BIN_DIR or looked up in $PATH
def _find_tool(name):
    """Find a tool first in BIN_DIR, then in $PATH."""
    local = os.path.join(BIN_DIR, name)
    if os.path.isfile(local) and os.access(local, os.X_OK):
        return local
    import shutil
    found = shutil.which(name)
    if found:
        return found
    return local  # fallback; will error at runtime if missing

# Go-binary tools bundled with GenomEvo
BACTAG_BIN      = _find_tool("BactAG")
BACTCG_BIN      = _find_tool("bactcg")
BACTPG_BIN      = _find_tool("BactPG")
BACTPGA_BIN     = _find_tool("bactpga")
CLUSTALW2_BIN   = _find_tool("clustalw2")

# External tools expected in $PATH
# progressiveMauve, blastn, blastp, cd-hit

def check_external_tools():
    """Verify all required external tools are accessible.  Returns list of missing."""
    import shutil
    required = ["progressiveMauve", "blastn", "blastp", "cd-hit"]
    missing = [t for t in required if shutil.which(t) is None]
    return missing

# ---------------------------------------------------------------------------
# Default parameters
# ---------------------------------------------------------------------------

DEFAULT_THREADS = 8

# BactAG
BACTAG_DEFAULT_THREADS = 20

# BactCG
BACTCG_DEFAULT_CD_CUTOFF = 0.7
BACTCG_DEFAULT_CG1_CUTOFF = 0.8

# BactPG
BACTPG_DEFAULT_SIMILARITY = 0.7
BACTPG_DEFAULT_THREADS = 30

# Bact1DGR
BACT1DGR_DEFAULT_WORKERS = 8

# BactEvolTraj
BACTEVOLTRAJ_MIN_EVENT_LENGTH = 1000

# ---------------------------------------------------------------------------
# Logging
# ---------------------------------------------------------------------------

def setup_logging(level=logging.INFO):
    logging.basicConfig(
        level=level,
        format="[%(asctime)s] %(levelname)s - %(name)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    return logging.getLogger("GenomEvo")

logger = setup_logging()
