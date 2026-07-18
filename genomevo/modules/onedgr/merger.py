import sys

def parse_dgr_file(filepath):
    """
    Parses a 1dgr file into a list of (start, end, label, strain).
    Returns a sorted list of intervals.
    """
    intervals = []
    try:
        with open(filepath, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                parts = line.split('\t')
                if len(parts) >= 3:
                    strain = parts[0]
                    coords = parts[1].split('-')
                    if len(coords) != 2:
                        continue # Malformed
                    start = int(coords[0])
                    end = int(coords[1])
                    label = parts[2]
                    intervals.append({'start': start, 'end': end, 'label': label, 'strain': strain})
    except FileNotFoundError:
        print(f"File not found: {filepath}", file=sys.stderr)
        return []
    
    intervals.sort(key=lambda x: x['start'])
    return intervals

def subtract_intervals(candidate, blockers):
    """
    Subtracts 'blockers' (list of intervals) from 'candidate' (single interval).
    Returns a list of remaining intervals from candidate.
    Blockers must be sorted by start.
    """
    if not blockers:
        return [candidate]
        
    result = []
    current_start = candidate['start']
    current_end = candidate['end']
    
    # Find relevant blockers
    # Optimization: blockers are sorted. We can skip those that end before candidate starts.
    # And stop when blocker starts after candidate ends.
    
    relevant_blockers = []
    for b in blockers:
        if b['end'] < current_start:
            continue
        if b['start'] > current_end:
            break
        relevant_blockers.append(b)
        
    if not relevant_blockers:
        return [candidate]
        
    for b in relevant_blockers:
        # If there is a gap before the blocker intersection
        if current_start < b['start']:
            # Add the piece before the blocker
            # Valid piece: [current_start, min(current_end, b['start'] - 1)]
            piece_end = min(current_end, b['start'] - 1)
            if piece_end >= current_start:
                result.append({'start': current_start, 'end': piece_end, 'label': candidate['label'], 'strain': candidate['strain']})
        
        # Advance current_start to after the blocker
        current_start = max(current_start, b['end'] + 1)
        if current_start > current_end:
            break
            
    # If there is simple piece left after all blockers
    if current_start <= current_end:
        result.append({'start': current_start, 'end': current_end, 'label': candidate['label'], 'strain': candidate['strain']})
        
    return result

def merge_1dgr(base_file, patch_file, strain_name, output_file=None):
    """
    Merges patch_file into base_file. Base file has priority.
    """
    base_intervals = parse_dgr_file(base_file)
    patch_intervals = parse_dgr_file(patch_file)
    
    # Provide gap filling from patch to base
    
    final_intervals = list(base_intervals)
    
    # Optimize: base_intervals is sorted.
    # For each patch interval, subtract all base_intervals.
    # This could be O(N*M). With sorted base, we can do better, but simple subtract is robust.
    
    for p in patch_intervals:
        remnants = subtract_intervals(p, base_intervals)
        final_intervals.extend(remnants)
        
    # Sort final result
    final_intervals.sort(key=lambda x: x['start'])
    
    # Merge adjacent identical intervals?
    # The original Perl script merges adjacent intervals if they have same label.
    # "if $dgr{$l} eq $dgr{$k}"
    
    merged_intervals = []
    if final_intervals:
        current = final_intervals[0]
        for next_iv in final_intervals[1:]:
            # Check for adjacency (next starts where current ends + 1) AND same label
            if (next_iv['start'] == current['end'] + 1) and (next_iv['label'] == current['label']):
                # Merge
                current['end'] = next_iv['end']
            else:
                merged_intervals.append(current)
                current = next_iv
        merged_intervals.append(current)
    
    # Output
    # If strain_name is provided, override the strain column?
    # Original perl: print $ARGV[2] ... which is strain name. Yes.
    
    output_lines = []
    for iv in merged_intervals:
        line = f"{strain_name}\t{iv['start']}-{iv['end']}\t{iv['label']}"
        output_lines.append(line)
        
    if output_file:
        with open(output_file, 'w') as f:
            f.write('\n'.join(output_lines) + '\n')
            
    return output_lines

if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("Usage: merger.py <base_file> <patch_file> <strain_name> [output_file]")
        sys.exit(1)
    
    out = None
    if len(sys.argv) > 4:
        out = sys.argv[4]
        
    res = merge_1dgr(sys.argv[1], sys.argv[2], sys.argv[3], out)
    if not out:
        for line in res:
            print(line)
