import pandas as pd
import plotly.graph_objects as go
import plotly.express as px
from Bio import SeqIO
import os
import glob
import json
import plotly.io as pio

# =============================================================================
#                           用户配置区域 (USER CONFIGURATION)
# =============================================================================

# 1. 基础工作目录
BASE_DIR = "/home/wangyiwei/1DGR_draw" 

# 2. 文件夹名称设置
TXT_FOLDER_NAME = "1DGR_en"  # 1DGR文件 (文件名可能是乱码ID)
GBK_FOLDER_NAME = "GBK_en"   # GBK文件 (文件名通常是标准名a开头)
OUTPUT_FOLDER_NAME = "Mosaic_Single_Charts_en" # 结果输出文件夹

# 3. ID 映射表 ("标准名 a开头" : "原始ID 乱码")
NAME_MAP = {
    "aAB1E":   "2ItB4rO",
    "aAB1EF":  "A5P4Lqk",
    "aAB1EFD": "TFSiK9E",
    "aD":      "AMI3bqA",
    "aEC1":    "OFj8oMq",
    "aEC":     "krcLSkM",
    "aB2G":    "8HjsyXM",
    "aB2":     "epvvXiR",
    "aG":      "IpEoZmn",
    "aE":      "9CECJAF",
    "aAB1":    "lM939JR",
    "aB1":     "ZBTdNo4",
    "aA":      "7diSDxH"
}

# 4. 需要绘图的节点列表
#    既然只画单图，这里其实只需要列出节点名即可。
#    为了方便，依然沿用之前的列表格式，脚本会自动提取 Child 进行绘制。
RELATIONSHIPS = [
    ("entericaAOG", "mA"),     
    ("entericaAOG", "mB"),
    ("mB", "mC"),   
    ("mC", "mD"),  
    ("mD", "mE"),
    ("mD", "mF")
]

# 反向映射表 (乱码 -> 标准名)
RAW_TO_STD_MAP = {v: k for k, v in NAME_MAP.items()}

# =============================================================================
#                       工具函数
# =============================================================================
def clean_id(raw_id):
    """将乱码 ID 转换为标准名 (a开头)"""
    if raw_id.startswith('a'):
        return raw_id
    return RAW_TO_STD_MAP.get(raw_id, raw_id)

# =============================================================================
#                       核心处理类
# =============================================================================

class MosaicDashboard:
    def __init__(self, gbk_path, map_file_path, node_name):
        self.gbk_path = gbk_path
        self.map_file_path = map_file_path
        self.node_name = node_name
        
        self.genome_length = 0
        self.genes = []
        self.segments = []
        self.gene_dict = {} 
        
    def parse_data(self):
        """解析 GBK 和 1DGR"""
        # 1. 读取 GBK
        print(f"[{self.node_name}] Reading GBK...")
        try:
            record = SeqIO.read(self.gbk_path, "genbank")
            self.genome_length = len(record.seq)
            
            gene_list = []
            for feature in record.features:
                if feature.type in ["CDS", "rRNA", "tRNA"]:
                    gene_name = feature.qualifiers.get('gene', ['Unknown'])[0]
                    if gene_name == "Unknown":
                        gene_name = feature.qualifiers.get('locus_tag', ['Unknown'])[0]
                    
                    product = feature.qualifiers.get('product', [''])[0]
                    start = int(feature.location.start)
                    end = int(feature.location.end)
                    
                    gene_list.append({
                        "Name": gene_name,
                        "Product": product,
                        "Start": start,
                        "End": end,
                        "Midpoint": (start + end) / 2,
                        "Length": end - start
                    })
            self.genes = pd.DataFrame(gene_list)
        except Exception as e:
            print(f"Error parsing GBK: {e}")
            return False

        # 2. 读取 Map
        print(f"[{self.node_name}] Reading Map...")
        seg_data = []
        try:
            with open(self.map_file_path, 'r') as f:
                for line in f:
                    parts = line.strip().split()
                    if len(parts) < 4: continue
                    if '-' in parts[2]:
                        s_txt, e_txt = parts[2].split('-')
                        start, end = int(s_txt), int(e_txt)
                        
                        raw_source_id = parts[3]
                        # 关键：ID 清洗
                        final_source_id = clean_id(raw_source_id)
                        
                        seg_data.append({
                            "Fragment_ID": parts[1],
                            "Start": start,
                            "End": end,
                            "Source_ID": final_source_id,
                            "Length": end - start + 1
                        })
            self.segments = pd.DataFrame(seg_data)
        except Exception as e:
            print(f"Error parsing TXT: {e}")
            return False
            
        return True

    def process_mapping(self):
        if self.segments.empty: return

        # 颜色配置 (仅针对 Source)
        sources = self.segments['Source_ID'].unique()
        plotly_palette = px.colors.qualitative.Dark24 + px.colors.qualitative.Alphabet
        final_palette = plotly_palette 
        color_map = {src: final_palette[i % len(final_palette)] for i, src in enumerate(sources)}
        
        # 强制颜色覆盖 (保持原有设计)
        custom_overrides = {
            'aB1': '#D62839', 'aB2': '#1e90ff', 'aA': '#006400',
            'aE': '#00ffff', 'mF': '#ff00ff', 'aD': '#9400d3',
            'aEC1': '#000000', 'entericaAOG': '#D3D3D3', 'mA': '#E69F00',
            'mB': '#CC79A7', 'mC': '#808000', 'mD': '#708090', 'mE': '#F0E442'
        }
        color_map.update(custom_overrides)
        
        self.segments['Color'] = self.segments['Source_ID'].map(color_map).fillna("#cccccc")
        
        # 基因映射 (用于点击交互)
        for frag_id in self.segments['Fragment_ID']:
            self.gene_dict[frag_id] = []
        self.gene_dict["Unknown"] = []

        if not self.genes.empty:
            for idx, gene in self.genes.iterrows():
                g_mid = gene['Midpoint'] + 1
                match = self.segments[(self.segments['Start'] <= g_mid) & (self.segments['End'] >= g_mid)]
                if not match.empty:
                    frag_id = match.iloc[0]['Fragment_ID']
                    self.gene_dict[frag_id].append({
                        "name": gene['Name'],
                        "product": gene['Product'],
                        "loc": f"{gene['Start']}-{gene['End']}"
                    })

    def generate_dashboard(self, output_file):
        if self.segments.empty: return

        def bp_to_theta(bp): return (bp / self.genome_length) * 360
        def bp_len_to_width(l): return (l / self.genome_length) * 360

        # --- 单图布局 ---
        fig = go.Figure()
        
        seg_mid = (self.segments['Start'] + self.segments['End']) / 2
        
        # 绘制唯一的极坐标柱状图
        fig.add_trace(go.Barpolar(
            r=[1.0] * len(self.segments), 
            theta=bp_to_theta(seg_mid),
            width=bp_len_to_width(self.segments['Length']),
            marker_color=self.segments['Color'],
            marker_line_width=0, opacity=0.9,
            name='Source',
            customdata=self.segments['Fragment_ID'], # 用于 JS 交互
            hoverinfo='text',
            hovertext="<b>ID:</b> " + self.segments['Fragment_ID'] + 
                      "<br><b>Source:</b> " + self.segments['Source_ID']
        ))

        # 隐藏坐标轴和网格
        fig.update_layout(
            template="plotly_white",
            polar=dict(
                radialaxis=dict(visible=False, range=[0, 1.1]), 
                angularaxis=dict(visible=False, showgrid=False, showticklabels=False),
                hole=0.4 # 中间留白
            ),
            title=dict(
                text=f"<b>Mosaic Map: {self.node_name}</b>",
                y=0.95, x=0.5, xanchor='center', yanchor='top'
            ),
            margin=dict(t=50, b=20, l=20, r=20),
            height=None,
            showlegend=False,
            clickmode='event+select'
        )

        plot_div = pio.to_html(fig, full_html=False, include_plotlyjs='cdn')
        js_data = json.dumps(self.gene_dict)
        
        html_content = f"""
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="utf-8">
            <title>Mosaic Map: {self.node_name}</title>
            <style>
                body {{ margin: 0; padding: 0; font-family: 'Segoe UI', sans-serif; overflow: hidden; height: 100vh; display: flex; flex-direction: column; }}
                .navbar {{
                    background-color: #2c3e50; color: white; padding: 0 20px; height: 50px;
                    display: flex; align-items: center; justify-content: space-between;
                    box-shadow: 0 2px 5px rgba(0,0,0,0.2); z-index: 10;
                }}
                .navbar .title {{ font-weight: bold; font-size: 18px; }}
                .toggle-btn {{
                    background: none; border: 1px solid #7f8c8d; color: white; padding: 5px 15px; cursor: pointer; border-radius: 4px; font-size: 14px; transition: 0.2s;
                }}
                .toggle-btn:hover {{ background: #34495e; }}
                .main-container {{ display: flex; flex: 1; overflow: hidden; position: relative; }}
                .chart-area {{
                    flex: 1; background: #fff; position: relative; transition: all 0.3s ease;
                    display: flex; align-items: center; justify-content: center;
                }}
                .sidebar {{
                    width: 350px; background: #f8f9fa; border-left: 1px solid #ddd;
                    display: flex; flex-direction: column; transition: width 0.3s ease; overflow: hidden;
                }}
                .sidebar.collapsed {{ width: 0; border: none; }}
                .sidebar-content {{ padding: 20px; overflow-y: auto; flex: 1; min-width: 350px; }}
                h3 {{ margin-top: 0; border-bottom: 2px solid #e67e22; padding-bottom: 10px; font-size: 16px; color: #333; }}
                .gene-card {{
                    background: white; border-left: 4px solid #3498db; margin-bottom: 8px; padding: 10px;
                    box-shadow: 0 2px 4px rgba(0,0,0,0.05); font-size: 13px;
                }}
                .gene-name {{ font-weight: bold; color: #2c3e50; font-size: 14px; }}
                .gene-prod {{ color: #555; margin: 4px 0; }}
                .gene-loc {{ color: #999; font-size: 11px; }}
                .info-msg {{ color: #888; text-align: center; margin-top: 50px; font-style: italic; }}
            </style>
        </head>
        <body>
            <div class="navbar">
                <span class="title">{self.node_name}</span>
                <button class="toggle-btn" onclick="toggleSidebar()">☰ Gene Details</button>
            </div>
            <div class="main-container">
                <div class="chart-area" id="chart-container">
                    {plot_div}
                </div>
                <div class="sidebar" id="sidebar">
                    <div class="sidebar-content">
                        <h3 id="panel-title">Gene Details</h3>
                        <div id="gene-list">
                            <p class="info-msg">Click on any colored region to see annotated genes.</p>
                        </div>
                    </div>
                </div>
            </div>
            <script>
                var geneData = {js_data};
                var plotDiv = document.getElementsByClassName('plotly-graph-div')[0];
                var sidebar = document.getElementById('sidebar');
                var isCollapsed = false;

                plotDiv.on('plotly_click', function(data){{
                    if(data.points.length > 0){{
                        var pt = data.points[0];
                        var fragID = pt.customdata; 
                        if(fragID) {{
                            renderSidebar(fragID);
                            if(isCollapsed) toggleSidebar();
                        }}
                    }}
                }});

                function renderSidebar(fragID) {{
                    var container = document.getElementById('gene-list');
                    var title = document.getElementById('panel-title');
                    title.innerHTML = "Fragment: " + fragID;
                    container.innerHTML = "";
                    var genes = geneData[fragID];
                    if(!genes || genes.length === 0) {{
                        container.innerHTML = "<p class='info-msg'>No genes annotated in this fragment.</p>";
                        return;
                    }}
                    var html = "";
                    genes.forEach(g => {{
                        html += `
                        <div class="gene-card">
                            <div class="gene-name">${{g.name}}</div>
                            <div class="gene-prod">${{g.product}}</div>
                            <div class="gene-loc">${{g.loc}}</div>
                        </div>`;
                    }});
                    container.innerHTML = html;
                }}

                function toggleSidebar() {{
                    isCollapsed = !isCollapsed;
                    if(isCollapsed) {{
                        sidebar.classList.add('collapsed');
                    }} else {{
                        sidebar.classList.remove('collapsed');
                    }}
                    setTimeout(function() {{
                        Plotly.Plots.resize(plotDiv);
                    }}, 350);
                }}
                window.onresize = function() {{
                    Plotly.Plots.resize(plotDiv);
                }};
            </script>
        </body>
        </html>
        """
        
        with open(output_file, "w", encoding='utf-8') as f:
            f.write(html_content)
        print(f"[{self.node_name}] Done! Saved to: {output_file}")

# =============================================================================
#                       公共入口函数 (供 bactfragann wrapper 调用)
# =============================================================================

def run_1dgr_mosaic(base_dir=None, txt_folder=None, gbk_folder=None, 
                    output_folder=None, name_map=None, relationships=None):
    """
    执行 1DGR 马赛克图生成。
    
    Parameters
    ----------
    base_dir : str
        基础工作目录
    txt_folder : str
        1DGR 文件所在子目录名
    gbk_folder : str
        GBK 文件所在子目录名
    output_folder : str
        输出子目录名
    name_map : dict
        标准名 -> 原始ID 的映射
    relationships : list of (parent, child)
        需要绘图的节点关系列表
    """
    if base_dir is not None:
        globals()['BASE_DIR'] = base_dir
    if txt_folder is not None:
        globals()['TXT_FOLDER_NAME'] = txt_folder
    if gbk_folder is not None:
        globals()['GBK_FOLDER_NAME'] = gbk_folder
    if output_folder is not None:
        globals()['OUTPUT_FOLDER_NAME'] = output_folder
    if name_map is not None:
        globals()['NAME_MAP'] = name_map
        globals()['RAW_TO_STD_MAP'] = {v: k for k, v in name_map.items()}
    if relationships is not None:
        globals()['RELATIONSHIPS'] = relationships

    txt_dir_path = os.path.join(BASE_DIR, TXT_FOLDER_NAME)
    gbk_dir_path = os.path.join(BASE_DIR, GBK_FOLDER_NAME)
    out_dir_path = os.path.join(BASE_DIR, OUTPUT_FOLDER_NAME)

    if not os.path.exists(out_dir_path):
        os.makedirs(out_dir_path, exist_ok=True)

    print(f"Work Dir: {BASE_DIR}")
    print("Mode: Single Mosaic Map (Source Only)")

    results = []
    for _, child_std in RELATIONSHIPS:
        child_gbk = os.path.join(gbk_dir_path, f"{child_std}.gbk")
        child_txt = os.path.join(txt_dir_path, f"{child_std}.1DGR.txt")

        if os.path.exists(child_gbk) and os.path.exists(child_txt):
            print(f"\n--- Processing: {child_std} ---")
            dashboard = MosaicDashboard(child_gbk, child_txt, node_name=child_std)
            if dashboard.parse_data():
                dashboard.process_mapping()
                out_path = os.path.join(out_dir_path, f"{child_std}_Mosaic.html")
                dashboard.generate_dashboard(out_path)
                results.append(out_path)
        else:
            print(f"[Warning] Missing files for {child_std} (GBK: {os.path.exists(child_gbk)}, TXT: {os.path.exists(child_txt)})")

    return results


# =============================================================================
#                                主程序入口
# =============================================================================

if __name__ == "__main__":
    run_1dgr_mosaic()