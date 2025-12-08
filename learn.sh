#!/bin/bash
# Inspect a Git project: structure and contents of non-ignored text files

set -euo pipefail

# --- Part 0: Check prerequisites ---
if ! git rev-parse --is-inside-work-tree &>/dev/null; then
    echo "Error: Not inside a Git repository."
    exit 1
fi

FILES=$(git ls-files --cached --others --exclude-standard)
if [ -z "$FILES" ]; then
    echo "No non-ignored files found in repo."
    exit 0
fi

# --- Part 1: Project structure ---
echo "========================================="
echo "          Project Structure"
echo "========================================="

if command -v tree &>/dev/null; then
    if tree --help 2>/dev/null | grep -q -- '--gitignore'; then
        tree --gitignore
    else
        tree -I .git
    fi
else
    echo "Warning: 'tree' not found. Flat file list shown."
    echo "$FILES"
fi

# --- Part 2: File contents ---
echo -e "\n\n========================================="
echo "      Contents of Non-Ignored Files"
echo "========================================="

MAX_SIZE=200000 # 200KB limit

echo "$FILES" | while IFS= read -r file; do
    if [ -f "$file" ] && [ "$(stat -c%s "$file")" -le $MAX_SIZE ] && grep -Iq . "$file"; then
        echo -e "\n\n-----------------------------------------"
        echo "File: $file"
        echo "-----------------------------------------"
        cat "$file"
    else
        echo -e "\n\n-----------------------------------------"
        echo "File: $file (Binary/Large/Directory - Skipped)"
        echo "-----------------------------------------"
    fi
done

echo -e "\n\nScript finished."