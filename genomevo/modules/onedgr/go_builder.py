import os
import subprocess
import glob

def build_go_binaries(src_dir, bin_dir):
    """
    Compiles all .go files in src_dir and places binaries in bin_dir.
    """
    if not os.path.exists(bin_dir):
        os.makedirs(bin_dir)

    go_files = glob.glob(os.path.join(src_dir, "*.go"))
    
    for go_file in go_files:
        base_name = os.path.splitext(os.path.basename(go_file))[0]
        output_bin = os.path.join(bin_dir, base_name)
        
        # Check if binary exists and is newer than source
        if os.path.exists(output_bin):
            src_mtime = os.path.getmtime(go_file)
            bin_mtime = os.path.getmtime(output_bin)
            if bin_mtime > src_mtime:
                print(f"Binary {base_name} is up to date.")
                continue
        
        print(f"Compiling {base_name}...")
        try:
            subprocess.check_call(["go", "build", "-o", output_bin, go_file])
            print(f"Successfully compiled {base_name}")
        except subprocess.CalledProcessError as e:
            print(f"Failed to compile {base_name}: {e}")
            raise

if __name__ == "__main__":
    import sys
    # Default paths if run directly
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    src_dir = os.path.join(project_root, "src", "go")
    bin_dir = os.path.join(project_root, "bin")
    try:
        build_go_binaries(src_dir, bin_dir)
    except Exception:
        sys.exit(1)
