"""
GenomEvo: An integrated system for bacterial comparative and evolutionary genomic analysis.

GenomEvo is an AOG (Ancestral Orthologous Genome)-centered pipeline that integrates:
  - BactAG:  Ancestral genome reconstruction
  - BactCG:  Core genome analysis
  - BactPG:  Pan-genome analysis
  - BactPGA: Pan-genome annotation
  - Bact1DGR: 1D genomic representation
  - BactEvolTraj: Evolutionary trajectory analysis
  - BactFragAnn: Fragment annotation & visualization

Usage:
    genomevo run --pipeline full --input <dir> --output <dir>
    genomevo web  # Launch web UI
"""

__version__ = "2.0.0"
__author__ = "Yejun Wang Lab, Shenzhen University"
__all__ = [
    "cli",
    "config",
    "modules",
    "web",
]
