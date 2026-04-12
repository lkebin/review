#!/bin/bash
# Run vim-code-review tests
# Usage: ./test/run.sh

set -e

cd "$(dirname "$0")/.."

echo "Running vim-code-review tests..."

# Run tests with Vim
vim -u NONE -N -Es -S test/run_tests.vim

echo "All tests passed!"
