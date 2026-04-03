

from glob import glob
import json
import os
import re
import subprocess
from typing import List

import numpy as np
import pandas as pd


def load_json(path):
    with open(path, 'r') as f:
        return json.load(f)


def find_json_files(benchmark_path):
    return sorted(glob(f"{benchmark_path}/*.json"))


def compute_throughput(benchmark_path):
    json_files = find_json_files(benchmark_path)
    if not json_files:
        return None

    leader_throughputs = []
    number_of_peers = None
    vallen = None
    runtime = None

    for jf in json_files:
        data = load_json(jf)
        roles = data["Setup"]["NodeRoleS"].split(",")
        number_of_peers = roles.count("1") + roles.count("2")
        vallen = data["Setup"].get("Vallen", vallen)
        cluster_batch_size = data["Setup"].get("ClusterBatchSize", None)
        clusterListS = data["Setup"].get("ClusterListS", [])
        number_of_clusters = None
        alternating = data["Setup"].get("AlternatingClusterPrepare", "false")
        nonoptimistic = data["Setup"].get("ClusterPrepareToAllAndEnableHashForClusterCommit", "false")
        if clusterListS:
            number_of_clusters = len(np.unique(clusterListS.split(",")))

        myid = int(data["Setup"]["MyId"])
        role = roles[myid]
        if role == "0":
            leader_throughputs.append(float(data["Throughput"]))
            if runtime is None:
                runtime = data.get("RunTime")

    return {
        "benchmark": benchmark_path,
        "throughput": float(np.sum(leader_throughputs)) if leader_throughputs else 0.0,
        "num_peers": number_of_peers if number_of_peers is not None else 0,
        "vallen": vallen,
        "runtime": runtime,
        "cluster_batch_size": cluster_batch_size,
        "number_of_clusters": number_of_clusters,
        "alternating": alternating,
        "nonoptimistic": nonoptimistic
    }


def process_throughput_benchmarks(benchmarks_list):
    rows = []
    for b in benchmarks_list:
        row = compute_throughput(b)
        if row:
            rows.append(row)
    return pd.DataFrame(rows)


class TikzPlotGenerator:
    def __init__(self, xs: List[List[any]], ys: List[List[any]], cat: List[any], xlabel, ylabel, filename, width="12cm", height="8cm", bar=False, stack=False, symbolic_x=False, sort_x_key=lambda x: x):

        self.xs = xs
        self.ys = ys
        self.cat = cat

        self.xlabel = xlabel
        self.ylabel = ylabel
        self.filename = filename
        self.width = width
        self.height = height
        self.bar = bar
        self.stack = stack
        self.symbolic_x = symbolic_x
        self.sort_x_key = sort_x_key

    def _preamble(self):
        return r"""\documentclass[tikz,border=6pt]{standalone}

\usepackage{graphicx}
\usepackage{tikz}
\usepackage{pgfplots}
\pgfplotsset{compat=1.18}
\usepgfplotslibrary{fillbetween}

"""

    def _color_palette(self):
        return r"""
% Colorblind-friendly palette

\pgfplotsset{every axis/.append style={line width=1pt}}

        \definecolor{cb_grey}{HTML}{4477AA}
        \definecolor{cb_indigo}{HTML}{EE6677}
        \definecolor{cb_green}{HTML}{228833}
        \definecolor{cb_teal}{HTML}{CCBB44}
        \definecolor{cb_blue}{HTML}{66CCEE}
        \definecolor{cb_rose}{HTML}{AA3377}
        \definecolor{cb_purple}{HTML}{BBBBBB}


        \pgfplotscreateplotcyclelist{colorblind}{
            {color=cb_indigo, mark=*},
            {color=cb_teal,   mark=square*},
            {color=cb_blue,   mark=triangle*},
            {color=cb_rose,   mark=diamond*},
            {color=cb_purple, mark=o},
            {color=cb_green,  mark=+},
            {color=cb_grey,   mark=x},
        }
        
        \pgfplotsset{/pgfplots/bar cycle list/.style={/pgfplots/cycle list={
            {fill=cb_indigo!30, color=cb_indigo,},
           {fill=cb_teal!30, color=cb_teal},
            {fill=cb_blue!30, color=cb_blue},
            {fill=cb_rose!30, color=cb_rose},
            {fill=cb_purple!30, color=cb_purple},
            {fill=cb_green!30, color=cb_green},
            {fill=cb_grey!30, color=cb_grey},},},}


"""

    def _document_start(self):
        return r"""
\begin{document}
\begin{tikzpicture}
"""

    def _axis_begin(self):
        # build axis options modularly so dynamic parts are inserted cleanly
        options = [
            "legend style={cells={anchor=east},legend pos=outer north east,}",
            "ymajorgrids=true",
            "grid style=dashed",
            f"width={self.width}",
            f"height={self.height}",
            "scale only axis",
            f"xlabel={{{self.xlabel}}}",
            f"ylabel={{{self.ylabel}}}",
            "cycle list name=colorblind",
        ]
        if self.bar:
            options.append("ybar" + (" stacked" if self.stack else ""))

        return "\\begin{axis}[\n        " + ",\n        ".join(options)

    def _latex_header(self):
        # assemble header from modular pieces
        parts = [
            self._preamble(),
            self._color_palette(),
            self._document_start(),
            self._axis_begin()
        ]
        return "".join(parts)

    def _xticks(self):
        xticks = []
        for x in self.xs:
            distinct_x_values = np.unique(x, axis=0)
            xticks.extend(list(distinct_x_values))
        xticks = list(sorted(np.unique(xticks, axis=0), key=self.sort_x_key))

        print(xticks)
        if self.symbolic_x:
            self.num_to_tick = {}
            for i, tick in enumerate(xticks):
                self.num_to_tick[tick] = i + 1
                xticks[i] = str(tick)

            formatted_xtick_string = ",\n        xtick={" + \
                ",".join(map(str, range(1, len(xticks)+1))) + "},\n"
            formatted_xtick_string += "x tick label style={rotate=45,anchor=east},\n"
            formatted_xtick_string += "        xticklabels={" + \
                ",".join(map(str, xticks)) + "}\n    ]\n"
        else:
            formatted_xtick_string = ",\n        xtick={" + \
                ",".join(map(str, xticks)) + "}\n    ]\n"

        return formatted_xtick_string

    def _plot_lines(self):
        plot = ""

        for i, (xs, ys, cat) in enumerate(zip(self.xs, self.ys, self.cat)):
            if self.symbolic_x:
                coords = " ".join(
                    [f"({self.num_to_tick[x]}, {y})" for x, y in zip(xs, ys)])
            else:
                coords = " ".join([f"({x}, {y})" for x, y in zip(xs, ys)])
            options = [
                f"name path={i}"
            ]
            if self.bar:
                options.append("ybar")
            plot += r"\addplot+"+f"[{",".join(options)}] " + \
                " coordinates {" + coords + "};\n"
            plot += f"\\addlegendentry{{{cat}}}\n"
        return plot

    def _latex_footer(self):
        return r"""\end{axis}
\end{tikzpicture}
\end{document}
"""

    def generate(self):
        latex_code = self._latex_header()
        latex_code += self._xticks()
        latex_code += self._plot_lines()
        latex_code += self._latex_footer()
        return latex_code

    def save(self):
        self.latex_code = self.generate()
        with open(self.filename, "w") as f:
            f.write(self.latex_code)
        return self.latex_code

    def compile(self, output_dir="outputs"):
        ret_code = os.system(
            f"pdflatex -interaction=nonstopmode -halt-on-error -output-directory={output_dir} {self.filename}")

        if ret_code != 0:
            with open(self.filename.replace(".tex", ".log"), "r") as f:
                log_content = f.read()
            raise Exception(
                f" LaTeX compilation failed with return code {ret_code}, {log_content}, {self.latex_code}")

        return ret_code


def trim_cpu_usage(cpu_usage_text):
    i = 0
    cpu_usage_lines = cpu_usage_text.split("\n")
    for line in cpu_usage_lines:
        if "usr" in line:
            break
        i += 1
    trimmed_cpu_usage = "\n".join(cpu_usage_lines[i:])
    trimmed_cpu_usage = re.sub(' +', ' ', trimmed_cpu_usage)
    return trimmed_cpu_usage


def extract_cpu_and_send_metrics(trimmed_cpu_usage):
    cpu = []
    send = []
    for row in trimmed_cpu_usage.splitlines()[1:]:
        split_cpu_line = row.strip().split(" ")
        idl = float(split_cpu_line[2])
        cpu.append(100-idl)
        send.append(split_cpu_line[8])
    return cpu, send


def clean_dool_output(dool_str):
    """
    Cleans the output from the 'dool' tool by removing ANSI escape sequences and non-ASCII characters.

    Args:
        dool_str (str): Raw output string from 'dool'.

    Returns:
        str: Cleaned output string.
    """
    # Remove ANSI escape sequences
    ansi_escape = re.compile(r'\x1b\[[0-9;]*[a-zA-Z]')
    cleaned = ansi_escape.sub('', dool_str)
    # Replace non-ASCII characters with space
    cleaned = re.sub(r'[^\x00-\x7F]+', ' ', cleaned)
    return cleaned


def compute_cpu_usage(benchmark_path, agg=np.max):
    json_files = find_json_files(benchmark_path)
    if not json_files:
        return None
    number_of_peers = None
    vallen = None
    runtime = None

    rows = []
    for jf in json_files:
        data = load_json(jf)
        roles = data["Setup"]["NodeRoleS"].split(",")
        number_of_peers = roles.count("1") + roles.count("2")
        vallen = data["Setup"].get("Vallen", vallen)
        myid = int(data["Setup"]["MyId"])
        role = int(roles[myid])
        cpu_usage = data["Dool"]
        cpu_usage, send = extract_cpu_and_send_metrics(
            trim_cpu_usage(clean_dool_output(cpu_usage)))

        cluster_batch_size = data["Setup"].get("ClusterBatchSize", None)
        clusterListS = data["Setup"].get("ClusterListS", [])
        number_of_clusters = None
        alternating = data["Setup"].get("AlternatingClusterPrepare", "false")
        nonoptimistic = data["Setup"].get("ClusterPrepareToAllAndEnableHashForClusterCommit", "false")

        if clusterListS:
            number_of_clusters = len(np.unique(clusterListS.split(",")))

        rows.append({
            "benchmark": benchmark_path,
            "num_peers": number_of_peers if number_of_peers is not None else 0,
            "vallen": vallen,
            "runtime": runtime,
            "cpu_usage": agg(np.array(cpu_usage)),
            "role": role,
            "cluster_batch_size": cluster_batch_size,
            "number_of_clusters": number_of_clusters,
            "alternating": alternating,
            "nonoptimistic": nonoptimistic
        })
    return rows


def process_cpu_usage(benchmarks_list, agg=np.max):
    rows = []
    for b in benchmarks_list:
        benchmark = compute_cpu_usage(b, agg=agg)
        if benchmark:
            rows.extend(benchmark)
    return pd.DataFrame(rows)


def get_role_df(df, role):
    """Filter dataframe by role."""
    return df.query(f"role == {role}")


def get_all_unique(df, columns):
    """Get unique values for given columns."""
    return [df[col].unique() for col in columns]


def complete_multiindex(df, index_cols, fill_value=0):
    """Reindex dataframe to fill missing combinations with fill_value."""
    df_unique = df.groupby(index_cols, as_index=False).sum(numeric_only=True)
    uniques = [df_unique[col].unique() for col in index_cols]
    mux = pd.MultiIndex.from_product(uniques, names=index_cols)
    return df_unique.set_index(index_cols).reindex(mux, fill_value=fill_value).reset_index()


def combine(row, cols):
    s = "("
    for col in cols:
        s += f"{row[col]},"
    s = s[:-1] + ")"
    return s


def add_combined_column(df, cols, new_col):
    """Add a combined string column from multiple columns."""
    df[new_col] = df[cols].apply(
        lambda row: combine(row, cols), axis=1)
    return df


def add_percent_column(df, group_col, value_col, percent_col):
    """Add a percent column based on group sums."""
    group_sum = df.groupby(group_col)[value_col].sum()
    df[percent_col] = df.apply(
        lambda row: row[value_col] / group_sum[row[group_col]
                                               ] if group_sum[row[group_col]] > 0 else 0,
        axis=1
    )
    return df


def plot_tikz(df, x_col, y_col, cat_col, filename, output_dir="outputs", **kwargs):
    """Create and save a TikzPlotGenerator plot."""
    plot = TikzPlotGenerator(
        xs=extract_groups_as_lists(df, x_col, cat_col),
        ys=extract_groups_as_lists(df, y_col, cat_col),
        cat=extract_unique_categories(df, cat_col),
        filename=filename,
        **kwargs
    )
    plot.save()

    return plot.compile(output_dir=output_dir)


def extract_groups_as_lists(df, col, cat):
    return [group[col].tolist() for (_, group) in df.groupby(cat, sort=True)]


def extract_unique_categories(df, cat):
    return [c for (c, _) in df.groupby(cat, sort=True)]


def compute_bytes_sent(benchmark_path, agg=np.max):
    json_files = find_json_files(benchmark_path)
    if not json_files:
        return None
    number_of_peers = None
    vallen = None
    runtime = None

    rows = []
    typs = set()
    for jf in json_files:
        data = load_json(jf)
        roles = data["Setup"]["NodeRoleS"].split(",")
        number_of_peers = roles.count("1") + roles.count("2")
        vallen = data["Setup"].get("Vallen", vallen)
        myid = int(data["Setup"]["MyId"])
        role = int(roles[myid])
        bytesSentTo = data["BytesSentTo"]
        bytesSent = {}
        cluster_batch_size = data["Setup"].get("ClusterBatchSize", None)
        clusterListS = data["Setup"].get("ClusterListS", [])
        alternating = data["Setup"].get("AlternatingClusterPrepare", "false")
        nonoptimistic = data["Setup"].get("ClusterPrepareToAllAndEnableHashForClusterCommit", "false")

        number_of_clusters = None
        if clusterListS:
            number_of_clusters = len(np.unique(clusterListS.split(",")))

        for bytesKeyValue in bytesSentTo:
            toRole = roles[int(bytesKeyValue["CID"])]
            bytesSent[(bytesKeyValue["Typ"])] = bytesSent.get(
                bytesKeyValue["Typ"], 0) + bytesKeyValue["Value"]
            typs.add(bytesKeyValue["Typ"])

        for t in typs:

            rows.append({
                "benchmark": benchmark_path,
                "num_peers": number_of_peers if number_of_peers is not None else 0,
                "vallen": vallen,
                "runtime": runtime,
                "typ": t,
                "bytes_sent": bytesSent.get(t, 0) / (roles.count("1") if str(role) == "1" else roles.count("2")),
                "role": role,
                "cluster_batch_size": cluster_batch_size,
                "number_of_clusters": number_of_clusters,
                "alternating": alternating,
                "nonoptimistic": nonoptimistic})

    return rows


def process_bytes_sent(benchmarks_list, agg=np.max):
    rows = []
    for b in benchmarks_list:
        benchmark = compute_bytes_sent(b, agg=agg)
        if benchmark:
            rows.extend(benchmark)
    return pd.DataFrame(rows)
