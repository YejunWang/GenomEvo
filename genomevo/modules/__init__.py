"""
GenomEvo modules – wrappers for each analysis component.

Each sub-module provides a run() function with a consistent signature,
accepting input paths, output directory, and keyword parameters.
"""

from .bactag import run_bactag
from .bactcg import run_bactcg
from .bactpg import run_bactpg
from .bactpga import run_bactpga
from .bact1dgr import run_bact1dgr
from .bactevoltraj import run_bactevoltraj
from .bactfragann import run_bactfragann

__all__ = [
    "run_bactag",
    "run_bactcg",
    "run_bactpg",
    "run_bactpga",
    "run_bact1dgr",
    "run_bactevoltraj",
    "run_bactfragann",
]
