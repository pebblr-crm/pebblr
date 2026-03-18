#!/usr/bin/env bash
# coverage-report.sh — Collect Go + TypeScript coverage and post a PR comment.
#
# Usage:
#   scripts/coverage-report.sh              # print report to stdout
#   scripts/coverage-report.sh --comment    # also post as a GitHub PR comment
#
# Environment:
#   GITHUB_REPOSITORY  — owner/repo   (set automatically in GitHub Actions)
#   GITHUB_REF         — refs/pull/N/merge (used to extract PR number)
#
# To add a coverage gate later, set COVERAGE_THRESHOLD (e.g. 80) and uncomment
# the gate section at the bottom of this script.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COVERAGE_DIR="${REPO_ROOT}/coverage"
mkdir -p "${COVERAGE_DIR}"

# ── Go coverage ─────────────────────────────────────────────────────────────
echo "▸ Running Go tests with coverage…"
go test -coverprofile="${COVERAGE_DIR}/go.out" ./... > /dev/null 2>&1 || true

GO_TOTAL="n/a"
if [[ -f "${COVERAGE_DIR}/go.out" ]]; then
    GO_TOTAL=$(go tool cover -func="${COVERAGE_DIR}/go.out" \
        | grep '^total:' \
        | awk '{print $NF}')
fi

# ── Frontend coverage ───────────────────────────────────────────────────────
echo "▸ Running frontend tests with coverage…"
(cd "${REPO_ROOT}/web" && bun run test -- --coverage --reporter=json \
    --outputFile="${COVERAGE_DIR}/web.json") > /dev/null 2>&1 || true

WEB_TOTAL="n/a"
if [[ -f "${COVERAGE_DIR}/web.json" ]]; then
    # Vitest JSON reporter puts totals under .testResults; coverage summary
    # is in the coverage-final or text output. We parse the text summary.
    # Fall back to the simpler vitest text output approach.
    true
fi

# Try parsing Vitest coverage summary from text reporter as fallback
(cd "${REPO_ROOT}/web" && bun run test -- --coverage \
    2>&1 | grep 'All files' | head -1) > "${COVERAGE_DIR}/web-text.tmp" 2>/dev/null || true

if [[ -s "${COVERAGE_DIR}/web-text.tmp" ]]; then
    # Vitest text output: "All files  |  XX.XX  | ..."
    WEB_TOTAL=$(awk -F'|' '{gsub(/[ \t]+/,"",$2); print $2"%"}' "${COVERAGE_DIR}/web-text.tmp")
fi

# ── Build markdown report ───────────────────────────────────────────────────
REPORT=$(cat <<EOF
## Code Coverage Report

| Component | Coverage |
|-----------|----------|
| **Go (backend)** | ${GO_TOTAL} |
| **TypeScript (frontend)** | ${WEB_TOTAL} |

<details>
<summary>Go per-package breakdown</summary>

\`\`\`
$(go tool cover -func="${COVERAGE_DIR}/go.out" 2>/dev/null || echo "No Go coverage data")
\`\`\`

</details>

<!-- coverage-report -->
EOF
)

echo ""
echo "${REPORT}"

# ── Post PR comment (optional) ──────────────────────────────────────────────
if [[ "${1:-}" == "--comment" ]]; then
    if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
        echo "⚠  GITHUB_REPOSITORY not set — skipping PR comment."
        exit 0
    fi

    PR_NUMBER=""
    if [[ "${GITHUB_REF:-}" =~ ^refs/pull/([0-9]+)/ ]]; then
        PR_NUMBER="${BASH_REMATCH[1]}"
    elif [[ -n "${GITHUB_EVENT_PATH:-}" ]]; then
        PR_NUMBER=$(jq -r '.pull_request.number // empty' "${GITHUB_EVENT_PATH}" 2>/dev/null || true)
    fi

    if [[ -z "${PR_NUMBER}" ]]; then
        echo "⚠  Could not determine PR number — skipping comment."
        exit 0
    fi

    # Delete previous coverage comment if present (idempotent updates)
    EXISTING=$(gh api "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
        --jq '.[] | select(.body | contains("<!-- coverage-report -->")) | .id' 2>/dev/null || true)

    if [[ -n "${EXISTING}" ]]; then
        for cid in ${EXISTING}; do
            gh api --method DELETE "repos/${GITHUB_REPOSITORY}/issues/comments/${cid}" > /dev/null 2>&1 || true
        done
    fi

    gh pr comment "${PR_NUMBER}" --body "${REPORT}"
    echo "✓ Coverage comment posted to PR #${PR_NUMBER}"
fi

# ── Coverage gate (uncomment when ready) ────────────────────────────────────
# COVERAGE_THRESHOLD="${COVERAGE_THRESHOLD:-0}"
# if [[ "${COVERAGE_THRESHOLD}" -gt 0 ]]; then
#     GO_NUM=$(echo "${GO_TOTAL}" | tr -d '%')
#     if (( $(echo "${GO_NUM} < ${COVERAGE_THRESHOLD}" | bc -l) )); then
#         echo "✗ Go coverage ${GO_TOTAL} is below threshold ${COVERAGE_THRESHOLD}%"
#         exit 1
#     fi
#     echo "✓ Go coverage ${GO_TOTAL} meets threshold ${COVERAGE_THRESHOLD}%"
# fi
