#!/usr/bin/env bash
# test-visual.sh — Build helm and capture screenshots of specific views.
#
# Usage:
#   scripts/test-visual.sh [mode]     Capture a single view (default: sessions)
#   scripts/test-visual.sh all        Capture all testable views
#   scripts/test-visual.sh clean      Remove captured screenshots
#
# Modes: sessions, bookmarks, projects, clone
#
# Output: .screenshots/<mode>.png (gitignored)
#
# Requirements: must be run from within tmux.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
HELM_BIN="$PROJECT_DIR/helm"
SCREENSHOT_DIR="$PROJECT_DIR/.screenshots"

WIDTH="${HELM_TEST_WIDTH:-150}"
HEIGHT="${HELM_TEST_HEIGHT:-40}"
DELAY="${HELM_TEST_DELAY:-0.8}"

MODES=(sessions bookmarks projects clone)

die() { echo "error: $*" >&2; exit 1; }

build() {
    echo "Building helm..."
    (cd "$PROJECT_DIR" && go build -o helm ./cmd/helm/) || die "build failed"
}

capture() {
    local mode="$1"
    local outfile="$SCREENSHOT_DIR/${mode}.png"

    local cmd="$HELM_BIN"
    if [[ "$mode" != "sessions" ]]; then
        cmd="$HELM_BIN --initial-view $mode"
    fi

    # Launch popup in background — it runs helm then sleeps to keep visible
    tmux display-popup -w"$WIDTH" -h"$HEIGHT" -E -B \
        "bash -c '$cmd & PID=\$!; sleep 3; kill \$PID 2>/dev/null'" &

    # Wait for helm to render, then capture
    sleep "$DELAY"
    screencapture -x "$outfile" 2>/dev/null || die "screencapture failed"

    # Wait for popup to self-close
    sleep 0.2

    echo "  $mode -> $outfile"
}

clean() {
    if [[ -d "$SCREENSHOT_DIR" ]]; then
        rm -rf "$SCREENSHOT_DIR"
        echo "Cleaned .screenshots/"
    else
        echo "Nothing to clean"
    fi
}

# Main
mode="${1:-sessions}"

if [[ "$mode" == "clean" ]]; then
    clean
    exit 0
fi

if [[ -z "${TMUX:-}" ]]; then
    die "must be run from within tmux"
fi

mkdir -p "$SCREENSHOT_DIR"
build

if [[ "$mode" == "all" ]]; then
    echo "Capturing all views..."
    for m in "${MODES[@]}"; do
        capture "$m"
        sleep 0.5
    done
    echo "Done. Screenshots in .screenshots/"
else
    valid=false
    for m in "${MODES[@]}"; do
        [[ "$m" == "$mode" ]] && valid=true
    done
    $valid || die "unknown mode: $mode (valid: ${MODES[*]})"

    capture "$mode"
fi
