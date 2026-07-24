import os
import concurrent.futures
import shutil
from .utils import run_command, count_fasta_bases
from .merger import merge_1dgr
from .formatter import format_1dgr
from .go_builder import build_go_binaries

def run_patch_pipeline(base_prefix, patch_strain, output_dir, bins, fasta_dir):
    """
    Runs the pipeline for a single patch.
    """
    base_name = base_prefix
    
    # Files
    base_fasta = os.path.join(fasta_dir, f"{base_prefix}.fasta")
    patch_fasta = os.path.join(fasta_dir, f"{patch_strain}.fasta")
    
    xmfa = os.path.join(output_dir, f"{base_name}.vs.{patch_strain}.xmfa")
    backbone = os.path.join(output_dir, f"{base_name}.vs.{patch_strain}.backbone")
    backbone_txt = backbone + ".txt"
    homblk = os.path.join(output_dir, f"{base_name}.vs.{patch_strain}.homblk.txt")
    orthblk = os.path.join(output_dir, f"{base_name}.vs.{patch_strain}.orthBlk.txt")
    out_1dgr = os.path.join(output_dir, f"{base_name}.{patch_strain}.1dgr.txt")
    
    # Check if final output exists? Rerun for now.
    
    # Mauve
    # progressiveMauve must be in PATH
    mauve_cmd = f"progressiveMauve --output={xmfa} --output-guide-tree={xmfa}.guide_tree --backbone-output={backbone} {base_fasta} {patch_fasta}"
    run_command(mauve_cmd, f"Aligning {base_prefix} vs {patch_strain}", output_file=None) # Mauve writes to specified outputs
    
    # Backprep
    run_command(f"{bins['progBackbonePrep']} {backbone}", 
                f"Preparing backbone for {patch_strain}", output_file=backbone_txt)
    
    # Reorder
    run_command(f"{bins['homBlkReorder']} {backbone_txt}", 
                f"Reordering hom blocks for {patch_strain}", output_file=homblk)
    
    # Parsing
    run_command(f"{bins['orthoParsing']} {homblk}", 
                f"Parsing orthologs for {patch_strain}", output_file=orthblk)
    
    # Extract
    run_command(f"{bins['1dgrExt1']} {orthblk} {base_name} {patch_strain}", 
                f"Extracting 1dgr for {patch_strain}", output_file=out_1dgr)
    
    return out_1dgr

def run_full_pipeline(base_strain, patches, output_dir, fasta_dir, max_workers=4, skip_mauve=False):
    """
    Runs full pipeline.
    """
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
        
    print(f"Base Strain: {base_strain}")
    print(f"Patches: {patches}")
    print(f"Output Dir: {output_dir}")
    
    base_fasta = os.path.join(fasta_dir, f"{base_strain}.fasta")
    genome_len = count_fasta_bases(base_fasta)
    print(f"Genome Length: {genome_len}")
    
    if genome_len == 0:
        print(f"Error: Genome length is 0. check {base_fasta}")
        return

    # Binaries setup
    script_dir = os.path.dirname(os.path.abspath(__file__)) # onedgr/
    package_root = os.path.dirname(script_dir) # OneDGR/
    bin_dir = os.path.join(package_root, "bin")
    src_dir = os.path.join(package_root, "src", "go")
    
    # Build binaries if needed
    print("Checking/Building binaries...")
    build_go_binaries(src_dir, bin_dir)
    
    bins = {
        'progBackbonePrep': os.path.join(bin_dir, "progBackbonePrep"),
        'homBlkReorder': os.path.join(bin_dir, "homBlkReorder"),
        'orthoParsing': os.path.join(bin_dir, "orthoParsing"),
        '1dgrExt1': os.path.join(bin_dir, "1dgrExt1")
    }
    
    patch_results = {}
    
    # Run patches in parallel
    print(f"Running {len(patches)} patch alignments in parallel with {max_workers} workers...")
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        future_to_patch = {
            executor.submit(run_patch_pipeline, base_strain, p, output_dir, bins, fasta_dir): p 
            for p in patches
        }
        
        for future in concurrent.futures.as_completed(future_to_patch):
            p = future_to_patch[future]
            try:
                res = future.result()
                patch_results[p] = res
                print(f"Patch {p} finished: {res}")
            except Exception as exc:
                print(f"Patch {p} failed: {exc}")
                # Don't abort immediately, wait for others
                
    if len(patch_results) != len(patches):
        print("Error: Not all patches completed successfully.")
        return

    # Sequentially merge
    print("Merging results...")
    
    # Assume order of patches matters? Usually yes for "progressive" filling.
    # Start with Patch 1
    current_merged_file = patch_results[patches[0]]
    
    for i in range(1, len(patches)):
        p = patches[i]
        next_patch_file = patch_results[p]
        merged_out = os.path.join(output_dir, f"{base_strain}.merged_{i+1}.1dgr.txt")
        
        print(f"Merging {p} results...")
        merge_1dgr(current_merged_file, next_patch_file, base_strain, output_file=merged_out)
        current_merged_file = merged_out
        
    # Final format
    final_dir = os.path.join(output_dir, "Final_Results")
    if not os.path.exists(final_dir):
        os.makedirs(final_dir)

    final_output = os.path.join(final_dir, f"{base_strain}.1DGR.txt")
    print(f"Formatting final output to {final_output}...")
    
    # Prefix: Use base_strain (without special chars?)
    # Perl used base_strain directly.
    prefix = base_strain.replace("_", "")
    
    format_1dgr(current_merged_file, genome_len, base_strain, prefix, output_file=final_output)
    
    print("Pipeline completed successfully.")
