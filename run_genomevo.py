#!/usr/bin/env python3
"""
GenomEvo entry point script.

Usage:
    python run_genomevo.py bactag --tree ./tree --gene ./genomes
    python run_genomevo.py web --port 8080
"""

import sys
import os

# Add the current directory to path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from genomevo.cli import main

if __name__ == "__main__":
    main()
