import pandas as pd
import plotly.graph_objects as go
from plotly.subplots import make_subplots
import numpy as np
import json
import os

# ---------------------------------------------------------
# 1. HTML 模板 (包含 CSS样式 和 JS交互逻辑)
# ---------------------------------------------------------
html_template = """
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Genome Map - {title}</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <style>
        body {{ margin: 0; padding: 0; display: flex; height: 100vh; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; overflow: hidden; }}
        
        /* 左侧图表容器 */
        #plot-container {{ flex: 3; height: 100%; position: relative; }}
        
        /* 右侧信息面板 */
        #info-panel {{ 
            flex: 1; 
            height: 100%; 
            overflow-y: auto; 
            padding: 25px; 
            background: #f8f9fa; 
            border-left: 1px solid #ddd; 
            box-shadow: -2px 0 10px rgba(0,0,0,0.05); 
            min-width: 350px;
            box-sizing: border-box;
        }}
        
        /* 信息卡片样式 */
        .info-card {{ background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); animation: fadeIn 0.3s; }}
        h2 {{ margin-top: 0; color: #2c3e50; font-size: 1.5rem; border-bottom: 2px solid #eee; padding-bottom: 10px; margin-bottom: 20px; }}
        
        .field-group {{ margin-bottom: 15px; }}
        .label {{ font-weight: 600; color: #7f8c8d; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 5px; }}
        .value {{ font-size: 1rem; color: #2c3e50; word-break: break-word; line-height: 1.4; }}
        
        .sequence-box {{ 
            font-family: 'Consolas', 'Monaco', monospace; 
            background: #f1f3f5; 
            padding: 10px; 
            border-radius: 4px; 
            max-height: 200px; 
            overflow-y: auto; 
            font-size: 0.8rem; 
            color: #495057;
            border: 1px solid #e9ecef;
            white-space: pre-wrap;
            word-break: break-all;
        }}
        
        .placeholder {{ text-align: center; color: #adb5bd; margin-top: 50%; transform: translateY(-50%); }}
        .badge {{ display: inline-block; padding: 4px 8px; border-radius: 4px; color: white; font-weight: bold; font-size: 0.9rem; }}
        
        @keyframes fadeIn {{ from {{ opacity: 0; transform: translateY(10px); }} to {{ opacity: 1; transform: translateY(0); }} }}
    </style>
</head>
<body>
    <div id="plot-container"></div>
    <div id="info-panel">
        <div class="placeholder">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="16" x2="12" y2="12"></line>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
            </svg>
            <p>Click on a colored region (Red/Blue)<br>to view detailed gene information.</p>
        </div>
    </div>

    <script>
        // Inject data from Python
        var plotData = {plot_data};
        var layout = {plot_layout};
        
        var config = {{responsive: true, displayModeBar: true}};
        
        // Create Plot
        Plotly.newPlot('plot-container', plotData, layout, config);
        
        // Add Click Listener
        var myPlot = document.getElementById('plot-container');
        myPlot.on('plotly_click', function(data){{
            var pt = data.points[0];
            // Access custom data we bound in Python
            // Structure: [Type, Start, End, Length, Genes, Sequence]
            var cd = pt.customdata; 
            
            if(cd && cd.length > 0) {{
                var color = cd[0] === 'Insertion' ? '#d62728' : '#1f77b4';
                var geneText = cd[4] ? cd[4] : '<em>Intergenic / Non-coding</em>';
                
                var html = `
                    <div class="info-card">
                        <h2>Event Details</h2>
                        
                        <div class="field-group">
                            <div class="label">Mutation Type</div>
                            <div class="value"><span class="badge" style="background-color: ${{color}}">${{cd[0]}}</span></div>
                        </div>
                        
                        <div class="field-group">
                            <div class="label">Genomic Position (EC)</div>
                            <div class="value">${{cd[1]}} - ${{cd[2]}}</div>
                        </div>
                        
                        <div class="field-group">
                            <div class="label">Length</div>
                            <div class="value">${{cd[3]}} bp</div>
                        </div>
                        
                        <div class="field-group">
                            <div class="label" style="color: #e74c3c;">Affected Genes</div>
                            <div class="value" style="font-weight: bold;">${{geneText}}</div>
                        </div>
                        
                        <div class="field-group">
                            <div class="label">Sequence Preview</div>
                            <div class="sequence-box">${{cd[5] ? cd[5] : 'N/A'}}</div>
                        </div>
                    </div>
                `;
                document.getElementById('info-panel').innerHTML = html;
            }}
        }});
    </script>
</body>
</html>
"""

# ---------------------------------------------------------
# 2. 数据处理与生成逻辑
# ---------------------------------------------------------
def parse_custom_txt(file_path):
    if file_path.endswith('.csv'):
        df = pd.read_csv(file_path)
    else:
        # Fallback for old format
        data = []
        headers = None
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
            
        start_idx = 0
        for i, line in enumerate(lines):
            if line.startswith('Parent') and 'Child' in line:
                headers = line.strip().split('\t')
                start_idx = i + 1
                break
                
        if not headers: return pd.DataFrame()
            
        for line in lines[start_idx:]:
            parts = line.strip().split('\t')
            if len(parts) >= len(headers):
                row = parts[:len(headers)]
                data.append(row)
                
        df = pd.DataFrame(data, columns=headers)
    
    numeric_cols = ['Start_EC', 'End_EC', 'Ref_Length_EC', 'Actual_Seq_Length']
    for col in numeric_cols:
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors='coerce')
    return df

def generate_interactive_genome_maps(file_path):
    df = parse_custom_txt(file_path)
    if df.empty:
        print("Error: Could not parse data.")
        return []

    df['Branch'] = df['Parent'] + '->' + df['Child']
    unique_branches = df['Branch'].unique()
    results = []

    for branch in unique_branches:
        branch_data = df[df['Branch'] == branch]
        
        # Determine Genome Size
        coords = branch_data[['Start_EC', 'End_EC']].values.flatten()
        coords = coords[~np.isnan(coords)]
        if len(coords) == 0: continue
        max_coord = coords.max()
        genome_size = max_coord * 1.05 if max_coord > 0 else 1000
        
        # Create Subplots
        fig = make_subplots(
            rows=1, cols=2,
            specs=[[{'type': 'polar'}, {'type': 'polar'}]],
            subplot_titles=("Gain Map (Insertion)", "Loss Map (Deletion)"),
            horizontal_spacing=0.15
        )
        
        # Common Settings
        # Increase outer radius (base_r) and reduce inner radius (base_inner)
        # to make the circular ring larger while shrinking the central hole.
        base_r = 4    # ring thickness / outer extension (was 1)
        base_inner = 3 # inner radius (was 5)
        
        # Helper to add traces
        def add_genome_trace(ax_col, event_type, color):
            # 1. Background Ring
            fig.add_trace(go.Barpolar(
                r=[base_r], theta=[0], width=[360], base=base_inner,
                marker_color='#e9ecef', marker_line_color='#dfe6e9', marker_line_width=0.8,
                hoverinfo='none', showlegend=False,
                customdata=[None] # Placeholder
            ), row=1, col=ax_col)
            
            # 2. Event Traces
            events = branch_data[branch_data['Type'] == event_type]
            if not events.empty:
                theta_vals = []
                width_vals = []
                custom_data_list = []
                
                for _, row in events.iterrows():
                    start = row['Start_EC']
                    
                    if event_type == 'Insertion':
                        theta = (start / genome_size) * 360
                        width = 360 * 0.006 # Slightly larger visual width for insertions
                    else: # Deletion
                        end = row['End_EC']
                        length = row['Ref_Length_EC']
                        center = (start + end) / 2
                        theta = (center / genome_size) * 360
                        width = max((length / genome_size) * 360, 360 * 0.004)

                    # Prepare Data for JS: [Type, Start, End, Actual_Seq_Length, Genes, Sequence]
                    actual_len = row['Actual_Seq_Length']
                    length_value = int(actual_len) if pd.notna(actual_len) else 'N/A'
                    
                    affected_genes_val = row.get('Affected_Genes', 'N/A')
                    sequence_val = row.get('Sequence', '')
                    
                    c_data = [
                        event_type, 
                        int(start), 
                        int(row['End_EC']), 
                        length_value,
                        str(affected_genes_val) if pd.notna(affected_genes_val) else "",
                        str(sequence_val) if pd.notna(sequence_val) else ""
                    ]
                    
                    theta_vals.append(theta)
                    width_vals.append(width)
                    custom_data_list.append(c_data)
                
                fig.add_trace(go.Barpolar(
                    r=[base_r] * len(theta_vals),
                    theta=theta_vals,
                    width=width_vals,
                    base=base_inner,
                    marker_color=color,
                    marker_line_width=0,
                    hoverinfo='text',
                    hovertext=['Click to view details'] * len(theta_vals),
                    customdata=custom_data_list,
                    name=event_type,
                    showlegend=False
                ), row=1, col=ax_col)

        # Build Traces
        add_genome_trace(1, 'Insertion', '#d62728') # Left: Red
        add_genome_trace(2, 'Deletion', '#1f77b4')  # Right: Blue
        
        # Layout
        common_polar = dict(
            radialaxis=dict(visible=False, range=[0, base_inner + base_r + 0.5]),
            angularaxis=dict(direction="clockwise", rotation=90, showgrid=False, showticklabels=False),
            bgcolor='rgba(0,0,0,0)'
        )
        
        fig.update_layout(
            title_text=f"Genome Comparison: {branch}",
            polar=common_polar,
            polar2=common_polar,
            height=800,
            margin=dict(l=20, r=20, t=80, b=20),
            paper_bgcolor='rgba(0,0,0,0)',
            plot_bgcolor='rgba(0,0,0,0)'
        )
        
        # Generate HTML
        graph_json = json.loads(fig.to_json())
        final_html = html_template.format(
            title=branch,
            plot_data=json.dumps(graph_json['data']),
            plot_layout=json.dumps(graph_json['layout'])
        )
        
        safe_name = branch.replace('->', '_to_')
        filename = f"Interactive_Genome_Map_{safe_name}.html"
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(final_html)
        print(f"Generated: {filename}")
        results.append(os.path.abspath(filename))

    return results

# ---------------------------------------------------------
# 公共入口函数 (供 bactfragann wrapper 调用)
# ---------------------------------------------------------

def run_evoltraj_circles(file_path=None, output_dir=None):
    """
    执行 EvolTraj 结构变异环图生成。

    Parameters
    ----------
    file_path : str
        SV 事件详情文件路径 (CSV 或旧格式 tab 分隔文本)
    output_dir : str
        输出目录，默认为当前目录
    """
    if output_dir:
        original_cwd = os.getcwd()
        os.makedirs(output_dir, exist_ok=True)
        os.chdir(output_dir)

    if file_path is None:
        # 尝试自动查找
        file_path = 'Large_SV_Details.csv'
    if not os.path.exists(file_path):
        file_path = 'All_Node_Pairs_Detailed_Info.txt'
    
    results = []
    if os.path.exists(file_path):
        results = generate_interactive_genome_maps(file_path)
    else:
        print(f"Error: File {file_path} not found.")

    if output_dir:
        os.chdir(original_cwd)
    return results


# ---------------------------------------------------------
# 运行脚本
# ---------------------------------------------------------
if __name__ == "__main__":
    run_evoltraj_circles()