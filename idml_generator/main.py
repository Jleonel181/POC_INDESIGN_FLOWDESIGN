#!/usr/bin/env python3
"""
Usage:
  # From backend API
  python main.py --edition 12 --output output/edition_12.idml

  # From local JSON file
  python main.py --json layout.json --output output/edition_12.idml
"""
import argparse
import sys
import os

sys.path.insert(0, os.path.dirname(__file__))

from src.container import build_use_case_from_api, build_use_case_from_json

BACKEND_URL = os.getenv("BACKEND_URL", "http://localhost:3001")


def main():
    parser = argparse.ArgumentParser(description="Generate IDML from layout JSON")
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--edition", type=int, help="Edition ID to fetch from backend API")
    source.add_argument("--json", type=str, help="Path to local layout JSON file")
    parser.add_argument("--output", required=True, help="Output .idml file path")
    args = parser.parse_args()

    if args.edition:
        use_case = build_use_case_from_api(BACKEND_URL)
        edition_id = args.edition
    else:
        use_case = build_use_case_from_json(args.json)
        edition_id = 0  # not used when reading from JSON

    output = use_case.execute(edition_id, args.output)
    print(f"✓ IDML generated: {output}")


if __name__ == "__main__":
    main()
