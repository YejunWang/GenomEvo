#!/usr/bin/env python3
import argparse
import sys
import os
from .pipeline import run_full_pipeline
from .utils import parse_bactid

def main_cli():
    parser = argparse.ArgumentParser(description="OneDGR Pipeline: 1D Genomic Representation for bacterial genomes.")
    
    parser.add_argument("base_strain", help="The base strain name (must match .fasta filename)")
    parser.add_argument("patches", nargs='*', help="List of patch strain names (must match .fasta filenames). If not provided, --bactid must be used.")
    
    parser.add_argument("--bactid", help="Path to bactID.txt file to determine patch order.")
    parser.add_argument("-o", "--output", default="OneDGR_Output", help="Output directory")
    parser.add_argument("-f", "--fasta-dir", default=".", help="Directory containing FASTA files")
    parser.add_argument("-w", "--workers", type=int, default=4, help="Number of parallel workers")
    
    args = parser.parse_args()

    # Determine patches list
    final_patches = []
    
    if args.bactid:
        print(f"Reading patch order from {args.bactid}...")
        
        # New Logic: Trace ancestry to find specific patches
        print(f"Tracing ancestry for '{args.base_strain}' in bactID.txt...")
        from .utils import get_bactid_ancestors
        ancestral_patches = get_bactid_ancestors(args.bactid, args.base_strain)
        
        if not ancestral_patches:
            print(f"Warning: No ancestors found for '{args.base_strain}' (or strain not found) in {args.bactid}.", file=sys.stderr)
            # If no ancestors found, maybe user wants *all* patches?
            # Or assume base_strain is root?
            # Or fall back to 'parse_bactid' and filter by position as before?
            print("Falling back to full list search...")
            # If get_bactid_ancestors fails (maybe line format is weird), try simple positional logic
            try:
                bactid_patches = parse_bactid(args.bactid)
                if args.base_strain in bactid_patches:
                     base_idx = bactid_patches.index(args.base_strain)
                     # Fallback: All nodes AFTER base_strain in file (as before)
                     potential = bactid_patches[base_idx+1:]
                     final_patches = list(reversed(potential))
                     print(f"Fallback: Using {len(final_patches)} nodes found after '{args.base_strain}' in file order.")
                else:
                     print(f"Warning: Strain '{args.base_strain}' not found in bactID.txt list.", file=sys.stderr)
                     # If base not found, try to use ALL patches except base?
                     final_patches = [p for p in bactid_patches if p != args.base_strain]
                     print(f"Fallback: Using all {len(final_patches)} nodes in file (excluding base).")

            except Exception as e:
               print(f"Error in fallback patch selection: {e}", file=sys.stderr)
               final_patches = []
 
        else:
            print(f"Found {len(ancestral_patches)} ancestral nodes: {ancestral_patches[0]} ... {ancestral_patches[-1]}")
            final_patches = ancestral_patches

        # If positional patches are provided, use them as a filter/selection BUT respect the evolutionary order derived above
        if args.patches:
            # Check if specified patches are in calculated ancestors
            # This allows user to sub-select from the valid lineage
            selected_set = set(args.patches)
            filtered = [p for p in final_patches if p in selected_set]
            
            # Warn if user provided patches NOT in lineage
            diff = selected_set - set(final_patches)
            if diff:
                 print(f"Warning: The following patches specified on command line are NOT direct ancestors of {args.base_strain} and will be ignored: {diff}", file=sys.stderr)
            
            final_patches = filtered

    if not final_patches:
        print("Error: No patches to process after filtering.", file=sys.stderr)
        sys.exit(1)

    # Check if Fastas exist
    base_fasta = os.path.join(args.fasta_dir, f"{args.base_strain}.fasta")
    if not os.path.exists(base_fasta):
        print(f"Error: Base FASTA not found: {base_fasta}", file=sys.stderr)
        sys.exit(1)
        
    for p in final_patches:
        p_fasta = os.path.join(args.fasta_dir, f"{p}.fasta")
        if not os.path.exists(p_fasta):
            print(f"Error: Patch FASTA not found: {p_fasta}", file=sys.stderr)
            sys.exit(1)

    print("Starting OneDGR Pipeline...")
    try:
        run_full_pipeline(args.base_strain, final_patches, args.output, args.fasta_dir, max_workers=args.workers)
    except Exception as e:
        print(f"An error occurred: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main_cli()
