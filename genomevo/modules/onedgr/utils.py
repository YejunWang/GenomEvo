import subprocess
import os
import sys
import re

def run_command(cmd, explanation, output_file=None):
    """
    Runs a shell command.
    """
    print(f"Running: {explanation}")
    # print(f"Command: {cmd}")
    
    if output_file:
        with open(output_file, 'w') as f:
            proc = subprocess.Popen(cmd, shell=True, stdout=f, stderr=sys.stderr)
            proc.wait()
            if proc.returncode != 0:
                print(f"Error running command: {cmd}", file=sys.stderr)
                raise subprocess.CalledProcessError(proc.returncode, cmd)
    else:
        proc = subprocess.Popen(cmd, shell=True, stderr=sys.stderr)
        proc.wait()
        if proc.returncode != 0:
            print(f"Error running command: {cmd}", file=sys.stderr)
            raise subprocess.CalledProcessError(proc.returncode, cmd)

def count_fasta_bases(fasta_file):
    """
    Counts bases in a FASTA file (ignoring headers and whitespace).
    """
    count = 0
    try:
        with open(fasta_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line.startswith('>'):
                    continue
                count += len(line)
    except FileNotFoundError:
        print(f"FASTA file not found: {fasta_file}", file=sys.stderr)
        return 0
    return count

def parse_bactid(filepath):
    """
    Parses bactID.txt to extract ALL unique strain/node names in order.
    Logic:
    1. Parse each line: "LeftParts OutsidE Group = Result"
    2. Extract names from LeftParts (split by +), Group (split by +), and Result.
    3. Add them to the list in order of appearance (uniquely).
    No filtering of internal IDs is performed based on user request.
    """
    patches = []
    seen = set()
    
    def add_if_new(name):
        name = name.strip()
        if name and name not in seen:
            patches.append(name)
            seen.add(name)

    try:
        with open(filepath, 'r') as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                
                # Parse: A+B OutsidE C = D
                # 1. Split by '=' to get D
                if '=' in line:
                    parts_eq = line.split('=')
                    result_part = parts_eq[-1] # D
                    remainder = "=".join(parts_eq[:-1]) # A+B OutsidE C
                    
                    # 2. Split remainder by 'OutsidE' to get A+B and C
                    if 'OutsidE' in remainder:
                        parts_outside = remainder.split('OutsidE')
                        left_part = parts_outside[0] # A+B
                        group_part = parts_outside[1] # C
                        
                        # Process Left (A+B)
                        for token in left_part.split('+'):
                            add_if_new(token)
                            
                        # Process Group (C) - C can also have +
                        for token in group_part.split('+'):
                            add_if_new(token)
                            
                        # Process Result (D)
                        add_if_new(result_part)
                else:
                    # Fallback for lines without '=' if any
                    pass
    
    except FileNotFoundError:
        print(f"bactID file not found: {filepath}", file=sys.stderr)
        sys.exit(1)
                    
    return patches

def get_bactid_ancestors(filepath, base_strain):
    """
    Parses bactID.txt to find all ancestors/nodes that contain the base_strain,
    tracing the path up to the root (the oldest ancestor).
    Returns a list of ancestor names in Deepest -> Closest order (Root -> Parent).
    """
    child_to_parents = {}
    ordered_nodes = []
    node_set = set()

    try:
        with open(filepath, 'r') as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                
                if "OutsidE" in line and "=" in line:
                    parts_eq = line.split('=')
                    parent = parts_eq[-1].strip()
                    remainder = "=".join(parts_eq[:-1])
                    
                    if parent not in node_set:
                        ordered_nodes.append(parent)
                        node_set.add(parent)
                        
                    if 'OutsidE' in remainder:
                        parts_out = remainder.split('OutsidE')
                        left_part = parts_out[0]
                        group_part = parts_out[1]
                        
                        children = []
                        children.extend([t.strip() for t in left_part.split('+')])
                        children.extend([t.strip() for t in group_part.split('+')])
                        
                        for child in children:
                            if child:
                                if child not in child_to_parents:
                                    child_to_parents[child] = []
                                child_to_parents[child].append(parent)

    except FileNotFoundError:
        print(f"bactID file not found: {filepath}", file=sys.stderr)
        return []

    ancestors = set()
    queue = [base_strain]
    visited = {base_strain}
    
    while queue:
        current = queue.pop(0)
        # Check parents
        if current in child_to_parents:
            parents = child_to_parents[current]
            for p in parents:
                if p not in visited:
                    visited.add(p)
                    ancestors.add(p)
                    queue.append(p)
    
    # Filter ordered_nodes
    relevant_ancestors = [n for n in ordered_nodes if n in ancestors]
    
    # Reverse to Deepest (Root) -> Closest (Parent)
    final_list = list(reversed(relevant_ancestors))
    
    # If base_strain itself is in the list (circular?), remove it.
    if base_strain in final_list:
        final_list.remove(base_strain)
        
    return final_list

def get_bactid_ancestors(filepath, base_strain):
    """
    Parses bactID.txt to find all ancestors/nodes that contain the base_strain,
    tracing the path up to the root (the oldest ancestor).
    Returns a list of ancestor names in Deepest -> Closest order (Root -> Parent).
    """
    child_to_parents = {}
    ordered_nodes = [] # To keep track of file order of nodes
    node_set = set()

    try:
        with open(filepath, 'r') as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                
                # Logic: Left + Right OutsidE Group = Parent
                if "OutsidE" in line and "=" in line:
                    # Split into LHS (children/groups) and RHS (Parent)
                    # A+B OutsidE C = D
                    # C and A+B are the components merge to form D context?
                    # Regardless, D is the *next step* node.
                    # Children: A, B, C. Parent: D.
                    
                    parts_eq = line.split('=')
                    parent = parts_eq[-1].strip()
                    remainder = "=".join(parts_eq[:-1])
                    
                    if parent not in node_set:
                        ordered_nodes.append(parent)
                        node_set.add(parent)
                        
                    if 'OutsidE' in remainder:
                        parts_out = remainder.split('OutsidE')
                        left_part = parts_out[0]
                        group_part = parts_out[1]
                        
                        children = []
                        children.extend([t.strip() for t in left_part.split('+')])
                        children.extend([t.strip() for t in group_part.split('+')])
                        
                        for child in children:
                            if child:
                                if child not in child_to_parents:
                                    child_to_parents[child] = []
                                child_to_parents[child].append(parent)

    except FileNotFoundError:
        print(f"bactID file not found: {filepath}", file=sys.stderr)
        return []

    # Traverse up from base_strain
    ancestors = set()
    queue = [base_strain]
    visited = {base_strain}
    
    while queue:
        current = queue.pop(0)
        if current in child_to_parents:
            parents = child_to_parents[current]
            for p in parents:
                if p not in visited:
                    visited.add(p)
                    ancestors.add(p) # Add parent to ancestors
                    queue.append(p)
    
    # Filter ordered_nodes to include only ancestors
    # ordered_nodes follows file order (usually bottom-up or top-down depending on file)
    # The file seems to build nodes: A+B->C. So C appears AFTER A and B.
    # So ancestors appear LATER in the file.
    # ordered_nodes = [Parent, GrandParent, Root]
    relevant_ancestors = [n for n in ordered_nodes if n in ancestors]
    
    # We want Deepest -> Closest.
    # If file is ChildIdx...RootIdx.
    # Then relevant_ancestors is [Parent, GrandParent, Root].
    # So we need to REVERSE it to get [Root, GrandParent, Parent].
    
    final_list = list(reversed(relevant_ancestors))
    
    return final_list
