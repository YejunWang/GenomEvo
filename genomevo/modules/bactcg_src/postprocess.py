import argparse
import logging
import os
import shutil
from concurrent.futures import ThreadPoolExecutor
import subprocess

logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")

def step2_get_fa(cdhit_dir, cg_file, out_dir):
    logging.info("Starting Step 2: get_fa_file-all")
    os.makedirs(out_dir, exist_ok=True)
    seq = {}
    for filename in os.listdir(cdhit_dir):
        if not filename.endswith('.fasta'): continue
        indexName = filename.replace('.fasta', '')
        with open(os.path.join(cdhit_dir, filename), 'r') as f:
            current_seq = ""
            for line in f:
                if line.startswith('>'):
                    current_seq = ">" + indexName + "---" + line.strip()[1:]
                    seq[current_seq] = ""
                elif line.strip():
                    seq[current_seq] += line.strip()
    
    with open(cg_file, 'r') as f:
        headers = f.readline().strip().split('\t')
        indexList = [h.strip() + "---" for h in headers]
        
        for line in f:
            if not line.strip(): continue
            parts = line.strip().split('\t')
            outLine = [indexList[i] + parts[i].strip() for i in range(min(len(parts), len(indexList)))]
            
            if not outLine: continue
            out_filename = outLine[0].replace(indexList[0], '')
            
            with open(os.path.join(out_dir, out_filename + ".fa"), 'w') as outf:
                for el in outLine:
                    k = ">" + el
                    if k in seq:
                        outf.write(k + "\n" + seq[k] + "\n")

def step3_get_mega(fa_dir, out_meg_dir):
    logging.info("Starting Step 3: get_mega-all")
    os.makedirs(out_meg_dir, exist_ok=True)
    fa_files = [f for f in os.listdir(fa_dir) if f.endswith('.fa')]
    
    # Try resolving clustalw2 from $PATH, or use local
    clustalw2 = shutil.which("clustalw2")
    if not clustalw2:
        clustalw2 = os.path.abspath("Tool/clustalw2")
    
    gcg2meg = os.path.abspath("Tool/gcg2meg.pl")
        
    def run_clustal(f):
        base = f.replace('.fa', '')
        in_path = os.path.abspath(os.path.join(fa_dir, f))
        out_gcg = os.path.abspath(os.path.join(out_meg_dir, base + '.gcg'))
        out_meg = os.path.abspath(os.path.join(out_meg_dir, base + '.meg'))
        
        # Run Clustal
        subprocess.run([clustalw2, f"-infile={in_path}", "-type=protein", "-output=gcg", f"-outfile={out_gcg}", "-pwmatrix=GONNET", "-pwgapopen=10", "-pwgapext=0.1", "-gapopen=10", "-gapext=0.2", "-gapdist=4", "-align", "-quiet"])
        
        # Convert GCG to MEG using the native perl script as the original did
        if os.path.exists(out_gcg):
            subprocess.run(["perl", gcg2meg, out_gcg, out_meg])

    with ThreadPoolExecutor(max_workers=os.cpu_count()) as executor:
        executor.map(run_clustal, fa_files)

def step4_get_snp_mega(meg_dir, final_out, log_out):
    logging.info("Starting Step 4: get_SNP_mega-all with indel reduction pipeline")
    
    # 1. Process indels for each .meg mapping
    reduced_dir = os.path.join(os.path.dirname(meg_dir), "4.SNP_mega_temp")
    os.makedirs(reduced_dir, exist_ok=True)
    
    for f in os.listdir(meg_dir):
        if not f.endswith(".meg"): continue
        file_path = os.path.join(meg_dir, f)
        
        prot_dict = {}
        with open(file_path, 'r') as p1:
            faa_name = None
            for i, line in enumerate(p1):
                if i < 4: continue
                line = line.strip()
                if line.startswith("#") and not line.startswith("#MEGA"):
                    faa_name = line
                    prot_dict[faa_name] = ""
                elif faa_name:
                    prot_dict[faa_name] += line
                    
        if not prot_dict:
            continue

        line_one = list(prot_dict.values())[0]
        
        del_num_list = set()
        for col_idx in range(len(line_one)):
            ref_aa = line_one[col_idx]
            identical = True
            for seq in prot_dict.values():
                if len(seq) > col_idx and seq[col_idx] != ref_aa:
                    identical = False
                    break
            if identical:
                del_num_list.add(col_idx)

        prot_dict_new = {}
        for key, seq in prot_dict.items():
            new_seq = "".join([seq[i] for i in range(len(seq)) if i not in del_num_list])
            prot_dict_new[key] = new_seq

        with open(os.path.join(reduced_dir, "new-" + f), "w", encoding="utf-8") as outf:
            outf.write("#mega\n!Title CG;\n!Format DataType=Protein indel=-;\n\n")
            for k, v in prot_dict_new.items():
                outf.write(k + "\n" + v + "\n\n")

    # 2. Combine reduced megas
    prot_sequence_dict = {}
    prot_log_dict = {}
    
    for i, f in enumerate(os.listdir(reduced_dir)):
        file_path = os.path.join(reduced_dir, f)
        
        with open(file_path, 'r') as p1:
            faa_name = None
            for line_idx, line in enumerate(p1):
                if line_idx < 4: continue
                line = line.strip()
                if not line: continue
                if line.startswith("#") and not line.startswith("#mega"):
                    # Original logic strips down to strain name
                    faa_name = line.replace("---", "").replace("#", "").split()[0]
                    if i == 0:
                        prot_sequence_dict[faa_name] = ""
                elif faa_name:
                    if i == 0:
                        prot_sequence_dict[faa_name] += line
                        prot_log_dict[f] = len(prot_sequence_dict[faa_name])
                    else:
                        prev_len = len(prot_sequence_dict.get(faa_name, ""))
                        if faa_name in prot_sequence_dict:
                            prot_sequence_dict[faa_name] += line
                            prot_log_dict[f] = len(prot_sequence_dict[faa_name]) - prev_len
                            
    os.makedirs(os.path.dirname(final_out), exist_ok=True)
    with open(final_out, "w", encoding="utf-8") as output_line:
        output_line.write("#mega\n!Title CG;\n!Format DataType=Protein indel=-;\n\n")
        for k, v in prot_sequence_dict.items():
            output_line.write("#" + k + "\n" + v + "\n\n")

    with open(log_out, "w", encoding="utf-8") as log_file:
        for k, v in prot_log_dict.items():
            log_file.write(f"{k}\t{v}\n")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Post-process CG results")
    parser.add_argument("--cdhit-dir", default="output/CG_results/1.cd-hit_output/1.cd-hit_fatsa", help="Dir with cdhit fasta")
    parser.add_argument("--cg-file", default="output/CG_results/CG_ALL.txt", help="CG core gene table txt")
    parser.add_argument("--outbase", default="output/CG_results/all-strain-together", help="Output base directory")
    args = parser.parse_args()
    
    step2_dir = os.path.join(args.outbase, "2.result")
    step3_dir = os.path.join(args.outbase, "3.result")
    final_meg = os.path.join(args.outbase, "4.result", "all_core_gene.meg")
    final_log = os.path.join(args.outbase, "4.result", "4.log_order_len.txt")
    
    step2_get_fa(args.cdhit_dir, args.cg_file, step2_dir)
    step3_get_mega(step2_dir, step3_dir)
    step4_get_snp_mega(step3_dir, final_meg, final_log)
    logging.info(f"Done! Final mega file at {final_meg}")

