#!/usr/bin/env python3
"""
GenomEvo – Bacterial Genome Evolution Analysis System
=====================================================

Setup script for installing GenomEvo as a Python package.

Usage:
    pip install -e .                    # Development install
    pip install .                       # Production install

After installation, the `genomevo` command will be available:
    genomevo --help
    genomevo web --port 8080
"""

from setuptools import setup, find_packages
import os

# Read long description from README
readme_path = os.path.join(os.path.dirname(__file__), "README.md")
long_description = ""
if os.path.exists(readme_path):
    with open(readme_path, "r", encoding="utf-8") as f:
        long_description = f.read()

setup(
    name="genomevo",
    version="2.0.0",
    author="Yejun Wang Lab",
    author_email="wangyj@szu.edu.cn",
    description="An integrated system for bacterial comparative and evolutionary genomic analysis",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/YejunWang/GenomEvo",
    packages=find_packages(),
    include_package_data=True,
    package_data={
        "genomevo": [
            "bin/*",
            "modules/onedgr/**/*.py",
            "modules/onedgr_src/go/*.go",
            "modules/bactag_src/**/*.go",
            "modules/bactag_src/**/*.mod",
            "modules/bactag_src/**/*.md",
            "modules/bactcg_src/**/*.go",
            "modules/bactcg_src/**/*.mod",
            "modules/bactcg_src/**/*.sum",
            "modules/bactcg_src/**/*.py",
            "modules/bactcg_src/**/*.md",
            "modules/bactpg_src/**/*.go",
            "modules/bactpg_src/**/*.mod",
            "modules/bactpg_src/**/*.md",
            "modules/bactpga_src/**/*.go",
            "modules/bactpga_src/**/*.mod",
            "modules/bactpga_src/**/*.md",
            "modules/*.py",
            "web/templates/*.html",
            "web/static/*.css",
        ],
    },
    entry_points={
        "console_scripts": [
            "genomevo=genomevo.cli:main",
        ],
    },
    python_requires=">=3.8",
    install_requires=[
        "biopython>=1.80",
        "pandas>=1.5",
        "numpy>=1.24",
        "matplotlib>=3.7",
        "plotly>=5.14",
        "flask>=2.3",
    ],
    extras_require={
        "dev": [
            "pytest",
            "pytest-cov",
        ],
    },
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Science/Research",
        "Topic :: Scientific/Engineering :: Bio-Informatics",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "License :: OSI Approved :: MIT License",
        "Operating System :: POSIX :: Linux",
    ],
    keywords="genomics, bacteria, evolution, ancestral genome, pan-genome, core genome, bioinformatics",
)
