"""
BactFragAnn module – Fragment Annotation & Visualization.

Generates interactive HTML dashboards for:
  - 1DGR fragment-to-gene annotation (Mosaic charts)
  - Evolutionary SV region annotation (circular plots)

Inputs:
  - mode:          "1dgr" or "evoltraj"
  - base_dir:      Working directory containing data folders
  - txt_folder:    Folder containing 1DGR .txt files (mode=1dgr)
  - gbk_folder:    Folder containing GenBank .gbk files
  - output_folder: Output folder name
  - name_map:      Dict mapping standard names to raw IDs
  - relationships: List of (parent, child) tuples
  - sv_file:       Path to SV event details CSV/TXT (mode=evoltraj)

Outputs:
  - Interactive HTML files (Plotly)

Dependencies: plotly, biopython, pandas
"""

import os
import sys
from ..config import logger


def run_bactfragann(
    mode: str = "1dgr",
    base_dir: str = ".",
    txt_folder: str = "1DGR_en",
    gbk_folder: str = "GBK_en",
    output_folder: str = "Mosaic_Charts",
    name_map: dict = None,
    relationships: list = None,
    sv_file: str = None,
    dry_run: bool = False,
) -> str:
    """
    Run fragment annotation and visualization.

    Parameters
    ----------
    mode : str
        "1dgr" for 1DGR fragment annotation, "evoltraj" for SV annotation.
    base_dir : str
        Base working directory containing data folders.
    txt_folder : str
        Folder name containing 1DGR output text files (mode=1dgr).
    gbk_folder : str
        Folder name containing GenBank files.
    output_folder : str
        Output folder name.
    name_map : dict
        Mapping of standard names to raw IDs (mode=1dgr).
    relationships : list
        List of (parent, child) tuples (mode=1dgr).
    sv_file : str
        Path to SV event details file (mode=evoltraj).
        Auto-detects Large_SV_Details.csv or All_Node_Pairs_Detailed_Info.txt.
    dry_run : bool

    Returns
    -------
    str : Path to output folder.
    """
    logger.info(f"Starting BactFragAnn ({mode})")
    logger.info(f"  Base dir : {base_dir}")
    logger.info(f"  Output   : {output_folder}")

    if dry_run:
        logger.info(f"[DRY RUN] mode={mode} base_dir={base_dir}")
        return os.path.join(base_dir, output_folder)

    if mode == "1dgr":
        from .BactFragAnn_for_1DGR import run_1dgr_mosaic

        results = run_1dgr_mosaic(
            base_dir=base_dir,
            txt_folder=txt_folder,
            gbk_folder=gbk_folder,
            output_folder=output_folder,
            name_map=name_map,
            relationships=relationships,
        )
        logger.info(f"BactFragAnn (1dgr) completed. Generated {len(results)} HTML files.")

    elif mode == "evoltraj":
        from .BactFragAnn_for_BactEvolTraj import run_evoltraj_circles

        out_dir = os.path.join(base_dir, output_folder)
        results = run_evoltraj_circles(
            file_path=sv_file,
            output_dir=out_dir,
        )
        logger.info(f"BactFragAnn (evoltraj) completed. Generated {len(results)} HTML files.")

    else:
        raise ValueError(f"Unknown BactFragAnn mode: {mode}")

    return os.path.abspath(os.path.join(base_dir, output_folder))
