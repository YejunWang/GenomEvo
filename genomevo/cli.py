#!/usr/bin/env python3
"""
GenomEvo CLI – Unified command-line interface for bacterial genome evolution analysis.

Usage:
    genomevo bactag      --tree <dir> --gene <dir> [--threads N]
    genomevo bactcg      --input <dir> --output <dir> --ref <strain>
    genomevo bactpg      --seq <dir> [--similarity S] [--threads N]
    genomevo bactpga     --mode parse|annotate|pipeline [options]
    genomevo bact1dgr    --base <strain> [--patches P1,P2] [--fasta-dir <dir>]
    genomevo bactevoltraj  --config <config.json>
    genomevo bactfragann --mode 1dgr|evoltraj --base-dir <dir>
    genomevo pipeline    --config <pipeline.json>   # Full pipeline
    genomevo web         [--port PORT]              # Launch web UI
    genomevo check                                 # Check dependencies
"""

import argparse
import json
import os
import sys

# Ensure the package root is on the path
_PKG_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if _PKG_ROOT not in sys.path:
    sys.path.insert(0, _PKG_ROOT)

from genomevo import __version__
from genomevo.config import logger, check_external_tools
from genomevo.modules import (
    run_bactag,
    run_bactcg,
    run_bactpg,
    run_bactpga,
    run_bact1dgr,
    run_bactevoltraj,
    run_bactfragann,
)


def cmd_check(args):
    """Check system dependencies."""
    print(f"GenomEvo v{__version__}")
    print("=" * 50)
    import shutil

    tools = {
        "progressiveMauve": "Mauve genome aligner",
        "blastn": "NCBI BLAST+ (nucleotide)",
        "blastp": "NCBI BLAST+ (protein)",
        "cd-hit": "CD-HIT clustering",
    }

    print("\nExternal tools:")
    all_ok = True
    for tool, desc in tools.items():
        found = shutil.which(tool)
        status = "✓ FOUND" if found else "✗ MISSING"
        if not found:
            all_ok = False
        print(f"  {status:12s} {tool:20s} ({desc})")

    print("\nBundled binaries:")
    from genomevo.config import BIN_DIR
    bundled = ["BactAG", "bactcg", "BactPG", "bactpga", "clustalw2"]
    for b in bundled:
        path = os.path.join(BIN_DIR, b)
        exists = os.path.isfile(path)
        status = "✓ FOUND" if exists else "✗ MISSING"
        print(f"  {status:12s} {b}")

    print("\nPython packages:")
    pkgs = ["Bio", "plotly", "pandas", "numpy", "matplotlib"]
    for p in pkgs:
        try:
            __import__(p)
            print(f"  ✓ FOUND      {p}")
        except ImportError:
            print(f"  ✗ MISSING    {p}")

    status = "ALL CHECKS PASSED" if all_ok else "SOME TOOLS MISSING"
    print(f"\n{status}")


def cmd_bactag(args):
    """Run ancestral genome reconstruction."""
    output = run_bactag(
        tree_dir=args.tree,
        gene_dir=args.gene,
        id_file=args.id_file,
        threads=args.threads,
        output_dir=args.output,
        dry_run=args.dry_run,
    )
    print(f"\nBactAG completed. Output: {output}")


def cmd_bactcg(args):
    """Run core genome analysis."""
    output = run_bactcg(
        input_dir=args.input,
        output_dir=args.output,
        ref_strain=args.ref,
        cd_cutoff=args.cd_c,
        cd_s=args.cd_s,
        cg1_cutoff=args.cg1,
        cg2=args.cg2,
        skip_qc=args.skip_qc,
        dry_run=args.dry_run,
    )
    print(f"\nBactCG completed. Output: {output}")


def cmd_bactpg(args):
    """Run pan-genome analysis."""
    output = run_bactpg(
        seq_dir=args.seq,
        output_dir=args.output,
        similarity=args.similarity,
        threads=args.threads,
        dry_run=args.dry_run,
    )
    print(f"\nBactPG completed. Output: {output}")


def cmd_bactpga(args):
    """Run pan-genome annotation."""
    if args.mode == "parse":
        result = run_bactpga(
            mode="parse",
            gbk_file=args.gbk,
            output_file=args.output_file,
            dry_run=args.dry_run,
        )
    elif args.mode == "annotate":
        result = run_bactpga(
            mode="annotate",
            gbk_file=args.gbk,
            pg_file=args.pg,
            mutbest_dir=args.mutbest_dir,
            seq_dir=args.seq_dir,
            strain_name=args.strain,
            output_file=args.output_file,
            dry_run=args.dry_run,
        )
    elif args.mode == "pipeline":
        result = run_bactpga(
            mode="pipeline",
            gbk_file=args.gbk,
            pg_file=args.pg,
            seq_dir=args.seq,
            output_dir=args.output,
            cov1=args.cov1,
            cov2=args.cov2,
            cg_path=args.cg,
            annot_mode=args.mode or "PGAG",
            dry_run=args.dry_run,
        )
    else:
        print(f"Unknown mode: {args.mode}")
        sys.exit(1)
    print(f"\nBactPGA ({args.mode}) completed.")


def cmd_bact1dgr(args):
    """Run 1D genomic representation."""
    patches = args.patches.split(",") if args.patches else None
    output = run_bact1dgr(
        base_strain=args.base,
        patches=patches,
        fasta_dir=args.fasta_dir,
        bactid_file=args.bactid,
        output_dir=args.output,
        workers=args.workers,
        dry_run=args.dry_run,
    )
    print(f"\nBact1DGR completed. Output: {output}")


def cmd_bactevoltraj(args):
    """Run evolutionary trajectory analysis."""
    if args.config:
        with open(args.config) as f:
            config = json.load(f)
        root_node = config["root_node"]
        fasta_dir = config.get("fasta_dir", ".")
        gbk_dir = config.get("gbk_dir", ".")
        tree_edges = config["tree_edges"]
        output_dir = config.get("output_dir", "Final_Large_SV_Analysis")
        min_len = config.get("min_event_length", 1000)
    else:
        root_node = args.root_node
        fasta_dir = args.fasta_dir
        gbk_dir = args.gbk_dir
        tree_edges = json.loads(args.tree) if args.tree else []
        output_dir = args.output
        min_len = args.min_len

    output = run_bactevoltraj(
        root_node=root_node,
        fasta_dir=fasta_dir,
        gbk_dir=gbk_dir,
        tree_edges=tree_edges,
        output_dir=output_dir,
        min_event_length=min_len,
        dry_run=args.dry_run,
    )
    print(f"\nBactEvolTraj completed. Output: {output}")


def cmd_bactfragann(args):
    """Run fragment annotation & visualization."""
    if args.config:
        with open(args.config) as f:
            config = json.load(f)
    else:
        config = {}

    output = run_bactfragann(
        mode=args.mode,
        base_dir=config.get("base_dir", args.base_dir or "."),
        txt_folder=config.get("txt_folder", "1DGR_en"),
        gbk_folder=config.get("gbk_folder", "GBK_en"),
        output_folder=config.get("output_folder", args.output or "Mosaic_Charts"),
        name_map=config.get("name_map"),
        relationships=config.get("relationships"),
        dry_run=args.dry_run,
    )
    print(f"\nBactFragAnn completed. Output: {output}")


def cmd_pipeline(args):
    """Run the full GenomEvo pipeline from a JSON config file."""
    if not args.config:
        print("Error: --config <pipeline.json> is required for full pipeline mode.")
        sys.exit(1)

    with open(args.config) as f:
        config = json.load(f)

    steps = config.get("steps", [])
    print(f"GenomEvo Pipeline – {len(steps)} step(s) defined.")

    for i, step in enumerate(steps, 1):
        module = step.get("module")
        params = step.get("params", {})
        print(f"\n--- Step {i}: {module} ---")

        if module == "bactag":
            run_bactag(**params)
        elif module == "bactcg":
            run_bactcg(**params)
        elif module == "bactpg":
            run_bactpg(**params)
        elif module == "bactpga":
            run_bactpga(**params)
        elif module == "bact1dgr":
            run_bact1dgr(**params)
        elif module == "bactevoltraj":
            run_bactevoltraj(**params)
        elif module == "bactfragann":
            run_bactfragann(**params)
        else:
            print(f"  Unknown module: {module}, skipping.")

    print("\nPipeline completed.")


def cmd_web(args):
    """Launch the GenomEvo web UI."""
    try:
        from genomevo.web.app import create_app
        app = create_app()
        print(f"\nGenomEvo Web UI starting at http://localhost:{args.port}")
        print("Press Ctrl+C to stop.")
        app.run(host="0.0.0.0", port=args.port, debug=args.debug)
    except ImportError as e:
        print(f"Error: Cannot start web UI. Missing dependency: {e}")
        print("Install with: pip install flask")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description=f"GenomEvo v{__version__} – Bacterial Genome Evolution Analysis System",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  genomevo check                          # Check all dependencies
  genomevo bactag -t 20 --tree ./tree --gene ./genomes
  genomevo bactcg -i ./proteins -o ./cg_out --ref MG1655
  genomevo bactpg --seq ./proteins -s 0.7 -t 30
  genomevo web --port 8080                # Start web interface

Full pipeline:
  genomevo pipeline --config pipeline.json
        """,
    )

    parser.add_argument("--version", action="version", version=f"GenomEvo v{__version__}")
    parser.add_argument("--dry-run", action="store_true", help="Print commands without executing")

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # --- check ---
    sp_check = subparsers.add_parser("check", help="Check system dependencies")

    # --- bactag ---
    sp_ag = subparsers.add_parser("bactag", help="Ancestral genome reconstruction (BactAG)")
    sp_ag.add_argument("--tree", "-T", required=True, help="Directory containing tree file")
    sp_ag.add_argument("--gene", "-G", required=True, help="Directory containing genome FASTA files")
    sp_ag.add_argument("--id-file", default="bactID.txt", help="Output bactID.txt path")
    sp_ag.add_argument("--threads", "-t", type=int, default=20, help="Number of threads (default: 20)")
    sp_ag.add_argument("--output", "-o", default="BactAG_Results", help="Output directory")
    sp_ag.add_argument("--dry-run", action="store_true")

    # --- bactcg ---
    sp_cg = subparsers.add_parser("bactcg", help="Core genome analysis (BactCG)")
    sp_cg.add_argument("--input", "-i", required=True, help="Input directory of protein FASTA files")
    sp_cg.add_argument("--output", "-o", required=True, help="Output directory")
    sp_cg.add_argument("--ref", "-r", required=True, help="Reference strain name")
    sp_cg.add_argument("--cd-c", type=float, default=0.7, help="CD-HIT sequence identity threshold (default: 0.7)")
    sp_cg.add_argument("--cd-s", type=float, default=0.7, help="CD-HIT length difference cutoff (default: 0.7)")
    sp_cg.add_argument("--cg1", type=float, default=0.8, help="BactCG first parameter / orthology cutoff (default: 0.8)")
    sp_cg.add_argument("--cg2", type=float, default=0.9, help="BactCG second parameter (default: 0.9)")
    sp_cg.add_argument("--skip-qc", action="store_true", help="Skip QC filtering step")
    sp_cg.add_argument("--dry-run", action="store_true")

    # --- bactpg ---
    sp_pg = subparsers.add_parser("bactpg", help="Pan-genome analysis (BactPG)")
    sp_pg.add_argument("--seq", "-s", required=True, help="Directory of sequence FASTA files")
    sp_pg.add_argument("--output", "-o", default="BactPG_Results", help="Output directory")
    sp_pg.add_argument("--similarity", type=float, default=0.7, help="Similarity threshold (default: 0.7)")
    sp_pg.add_argument("--threads", "-t", type=int, default=30, help="Number of threads (default: 30)")
    sp_pg.add_argument("--dry-run", action="store_true")

    # --- bactpga ---
    sp_pga = subparsers.add_parser("bactpga", help="Pan-genome annotation (BactPGA)")
    sp_pga.add_argument("--mode", choices=["parse", "annotate", "pipeline"], default="pipeline")
    sp_pga.add_argument("--gbk", help="Input GenBank file")
    sp_pga.add_argument("--pg", help="Pan-genome matrix (PG.txt)")
    sp_pga.add_argument("--seq", help="Sequence directory (pipeline mode)")
    sp_pga.add_argument("--strain", help="Strain name (annotate mode)")
    sp_pga.add_argument("--mutbest-dir", help="Mutual best hit directory (annotate mode)")
    sp_pga.add_argument("--seq-dir", help="Sequence directory (annotate mode)")
    sp_pga.add_argument("--cov1", type=float, default=0.7, help="Coverage 1 threshold for pipeline (default: 0.7)")
    sp_pga.add_argument("--cov2", type=float, default=0.7, help="Coverage 2 threshold for pipeline (default: 0.7)")
    sp_pga.add_argument("--cg", help="Path to CG executable (pipeline mode)")
    sp_pga.add_argument("--output", "-o", default="BactPGA_Results")
    sp_pga.add_argument("--output-file", help="Output file path (parse/annotate mode)")
    sp_pga.add_argument("--dry-run", action="store_true")

    # --- bact1dgr ---
    sp_1d = subparsers.add_parser("bact1dgr", help="1D genomic representation (Bact1DGR)")
    sp_1d.add_argument("--base", required=True, help="Base/ancestral strain name")
    sp_1d.add_argument("--patches", help="Comma-separated list of patch strains")
    sp_1d.add_argument("--fasta-dir", default=".", help="Directory containing FASTA files")
    sp_1d.add_argument("--bactid", help="Path to bactID.txt")
    sp_1d.add_argument("--output", "-o", default="OneDGR_Output")
    sp_1d.add_argument("--workers", "-w", type=int, default=8)
    sp_1d.add_argument("--dry-run", action="store_true")

    # --- bactevoltraj ---
    sp_et = subparsers.add_parser("bactevoltraj", help="Evolutionary trajectory analysis (BactEvolTraj)")
    sp_et.add_argument("--config", "-c", help="JSON config file")
    sp_et.add_argument("--root-node", help="Root ancestral genome name")
    sp_et.add_argument("--fasta-dir", default=".", help="FASTA directory")
    sp_et.add_argument("--gbk-dir", default=".", help="GenBank directory")
    sp_et.add_argument("--tree", help='JSON string of tree edges, e.g. \'[["A","B"],["B","C"]]\'')
    sp_et.add_argument("--output", "-o", default="Final_Large_SV_Analysis")
    sp_et.add_argument("--min-len", type=int, default=1000, help="Min event length (default: 1000)")
    sp_et.add_argument("--dry-run", action="store_true")

    # --- bactfragann ---
    sp_fa = subparsers.add_parser("bactfragann", help="Fragment annotation & visualization (BactFragAnn)")
    sp_fa.add_argument("--mode", choices=["1dgr", "evoltraj"], default="1dgr")
    sp_fa.add_argument("--config", "-c", help="JSON config file")
    sp_fa.add_argument("--base-dir", default=".", help="Base working directory")
    sp_fa.add_argument("--output", "-o", default="Mosaic_Charts")
    sp_fa.add_argument("--dry-run", action="store_true")

    # --- pipeline ---
    sp_pl = subparsers.add_parser("pipeline", help="Run full pipeline from JSON config")
    sp_pl.add_argument("--config", "-c", required=True, help="Pipeline JSON config file")

    # --- web ---
    sp_web = subparsers.add_parser("web", help="Launch web UI")
    sp_web.add_argument("--port", "-p", type=int, default=5000, help="Web server port (default: 5000)")
    sp_web.add_argument("--debug", action="store_true", help="Enable debug mode")

    args = parser.parse_args()

    if args.command is None:
        parser.print_help()
        sys.exit(1)

    # Route to handler
    handlers = {
        "check": cmd_check,
        "bactag": cmd_bactag,
        "bactcg": cmd_bactcg,
        "bactpg": cmd_bactpg,
        "bactpga": cmd_bactpga,
        "bact1dgr": cmd_bact1dgr,
        "bactevoltraj": cmd_bactevoltraj,
        "bactfragann": cmd_bactfragann,
        "pipeline": cmd_pipeline,
        "web": cmd_web,
    }

    handler = handlers.get(args.command)
    if handler:
        handler(args)


if __name__ == "__main__":
    main()
