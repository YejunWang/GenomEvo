import sys
from .merger import parse_dgr_file

def format_1dgr(input_file, genome_limit, strain_name, block_prefix, output_file=None):
    """
    Fills gaps with strain_name, merges identical blocks, filters small blocks.
    output_file can be a path or None (print to stdout).
    """
    
    # 1. Parse input
    intervals = parse_dgr_file(input_file)
    
    # 2. Fill gaps and flatten
    # Create a dense representation? Or just insert gap intervals.
    
    full_intervals = []
    current_pos = 1
    
    for iv in intervals:
        start = iv['start']
        end = iv['end']
        label = iv['label']
        
        # Fill gap before this interval
        if start > current_pos:
            gap_len = start - current_pos
            full_intervals.append({'start': current_pos, 'end': start - 1, 'label': strain_name, 'strain': strain_name})
            
        full_intervals.append(iv)
        current_pos = end + 1
        
    # Fill gap at the end
    if current_pos <= genome_limit:
        full_intervals.append({'start': current_pos, 'end': genome_limit, 'label': strain_name, 'strain': strain_name})
        
    # 3. Merge adjacent identical blocks
    merged_intervals = []
    if full_intervals:
        current = full_intervals[0]
        for next_iv in full_intervals[1:]:
            if next_iv['label'] == current['label']:
                current['end'] = next_iv['end']
            else:
                merged_intervals.append(current)
                current = next_iv
        merged_intervals.append(current)
    
    # 4. Filter small blocks (Absorb preceding small blocks logic)
    # Logic:
    # Iterate. If block is small, don't output yet. Pass its start to the next block.
    # If block is big, take the passed start (if any), output.
    # If block is last, take passed start, output.
    
    final_output = []
    pending_start = None
    block_counter = 0
    
    for i, iv in enumerate(merged_intervals):
        length = iv['end'] - iv['start'] + 1 # 1-based length
        original_start = iv['start']
        
        # Determine effective start
        if pending_start is not None:
            effective_start = pending_start
        else:
            effective_start = original_start
            
        is_last = (i == len(merged_intervals) - 1)
        
        if length >= 999:
            # It's a big block. Absorb any pending.
            block_counter += 1
            # Format block ID
            # Perl: _001, _010, _100.
            block_id = f"{block_prefix}_{block_counter:03d}"
            
            final_output.append({'strain': strain_name, 'block_id': block_id, 'start': effective_start, 'end': iv['end'], 'label': iv['label']})
            pending_start = None
            
        else:
            # It's a small block.
            if is_last:
                # Force output
                block_counter += 1
                block_id = f"{block_prefix}_{block_counter:03d}"
                final_output.append({'strain': strain_name, 'block_id': block_id, 'start': effective_start, 'end': iv['end'], 'label': iv['label']})
                pending_start = None
            else:
                # Buffer it.
                # If we already have a pending start, keep it (it's even earlier).
                # If not, set it to this block's start.
                if pending_start is None:
                    pending_start = original_start
                # Logic check: The perl script sets `stF{$count} = $stF{$countBack}` if prev was small.
                # So yes, pending_start propagates.
                pass
                
    # Output
    lines = []
    for item in final_output:
        line = f"{item['strain']}\t{item['block_id']}\t{item['start']}-{item['end']}\t{item['label']}"
        lines.append(line)
        
    if output_file:
        with open(output_file, 'w') as f:
            f.write('\n'.join(lines) + '\n')
            
    return lines

if __name__ == "__main__":
    if len(sys.argv) < 5:
        print("Usage: formatter.py <input_file> <genome_length> <strain_name> <block_prefix> [output_file]")
        sys.exit(1)
        
    out = None
    if len(sys.argv) > 5:
        out = sys.argv[5]
        
    res = format_1dgr(sys.argv[1], int(sys.argv[2]), sys.argv[3], sys.argv[4], out)
    if not out:
        for line in res:
            print(line)
