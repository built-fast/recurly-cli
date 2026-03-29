#!/usr/bin/env bash
# check-skill-drift.sh — verify that SKILL.md references match .surface entries.
#
# Three checks:
# 1. Every `recurly <cmd>` in SKILL.md resolves to a CMD in .surface
# 2. Every --<flag> on a line with `recurly <cmd>` exists on the resolved
#    command, its ancestors, or its descendants in .surface
# 3. Every leaf CMD in .surface is referenced in SKILL.md (coverage)
#
# Exit 0 if clean (or all drift baselined), 1 if unbaselined drift found.
# Entries in .surface-skill-drift are treated as accepted/known drift.

set -euo pipefail

SKILL="skills/recurly/SKILL.md"
SURFACE=".surface"
BASELINE=".surface-skill-drift"

if [[ ! -f "$SKILL" ]]; then
  echo "ERROR: $SKILL not found" >&2
  exit 1
fi
if [[ ! -f "$SURFACE" ]]; then
  echo "ERROR: $SURFACE not found" >&2
  exit 1
fi

# Strip YAML frontmatter (between first pair of --- lines)
skill_content=$(awk 'BEGIN{n=0} /^---$/{n++; next} n==0 || n>=2{print}' "$SKILL")

# Load baseline exceptions
declare -A baseline
if [[ -f "$BASELINE" ]]; then
  while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    baseline["$line"]=1
  done < "$BASELINE"
fi

# Drift tracking (associative arrays for deduplication, counters for safe access)
declare -A drift_all         # all detected drift (including baselined)
declare -A drift_unbaselined # drift not covered by baseline
total_drift=0
unbaselined_drift=0

add_drift() {
  local key="$1"
  # Skip duplicates
  [[ -n "${drift_all[$key]:-}" ]] && return
  drift_all["$key"]=1
  total_drift=$((total_drift + 1))
  if [[ -z "${baseline[$key]:-}" ]]; then
    drift_unbaselined["$key"]=1
    unbaselined_drift=$((unbaselined_drift + 1))
  fi
}

# resolve_cmd: strip trailing words until a CMD match is found in .surface.
# Prints the resolved CMD and returns 0, or returns 1 if no match.
resolve_cmd() {
  local cmd="$1"
  while [[ "$cmd" != "recurly" && -n "$cmd" ]]; do
    if grep -q "^CMD ${cmd}$" "$SURFACE"; then
      echo "$cmd"
      return 0
    fi
    cmd="${cmd% *}"
  done
  return 1
}

# check_flag: verify flag exists on the resolved command, ancestors, or descendants.
check_flag() {
  local resolved="$1" flag="$2"

  # Check resolved command and descendants
  if grep -q "^FLAG ${resolved} .*--${flag} " "$SURFACE"; then
    return 0
  fi

  # Check ancestors
  local ancestor="$resolved"
  while [[ "$ancestor" == *" "* ]]; do
    ancestor="${ancestor% *}"
    if grep -q "^FLAG ${ancestor} --${flag} " "$SURFACE"; then
      return 0
    fi
  done

  return 1
}

# ── Check 1: Command references ─────────────────────────────────────────────

while IFS= read -r match; do
  # Strip flags and trailing whitespace
  cmd=$(echo "$match" | sed 's/ --.*//; s/[[:space:]]*$//')
  [[ "$cmd" == "recurly" ]] && continue

  # Strip <placeholder> and [optional] args
  clean_cmd=$(echo "$cmd" | sed 's/ <[^>]*>//g; s/ \[[^]]*\]//g; s/[[:space:]]*$//')
  [[ -z "$clean_cmd" || "$clean_cmd" == "recurly" ]] && continue

  if ! resolve_cmd "$clean_cmd" > /dev/null 2>&1; then
    add_drift "CMD: ${clean_cmd}"
  fi
done < <(echo "$skill_content" | grep -oE 'recurly [a-z][-a-z0-9 ]*' | sort -u)

# ── Check 2: Flags on lines containing commands ─────────────────────────────

while IFS= read -r line; do
  # Extract the first command from the line
  match=$(echo "$line" | grep -oE 'recurly [a-z][-a-z0-9 ]*' | head -1 || true)
  [[ -z "$match" ]] && continue

  cmd=$(echo "$match" | sed 's/ --.*//; s/[[:space:]]*$//')
  [[ "$cmd" == "recurly" ]] && continue

  clean_cmd=$(echo "$cmd" | sed 's/ <[^>]*>//g; s/ \[[^]]*\]//g; s/[[:space:]]*$//')
  [[ -z "$clean_cmd" || "$clean_cmd" == "recurly" ]] && continue

  resolved=$(resolve_cmd "$clean_cmd" 2>/dev/null) || continue

  # Extract all --flag patterns from the line
  while IFS= read -r flag_match; do
    [[ -z "$flag_match" ]] && continue
    flag_name="${flag_match#--}"

    if ! check_flag "$resolved" "$flag_name"; then
      add_drift "FLAG: ${resolved} --${flag_name}"
    fi
  done < <(echo "$line" | grep -oE '\-\-[a-z][-a-z0-9]*' | sort -u)
done < <(echo "$skill_content" | grep 'recurly [a-z]' | grep -- '--[a-z]')

# ── Check 3: Coverage — every leaf CMD must be in SKILL.md ───────────────────

# Collect all CMDs
mapfile -t all_cmds < <(grep '^CMD ' "$SURFACE" | sed 's/^CMD //')

# Identify group commands (any CMD that is a strict prefix of another CMD)
declare -A is_group
for cmd in "${all_cmds[@]}"; do
  for other in "${all_cmds[@]}"; do
    if [[ "$other" != "$cmd" && "$other" == "${cmd} "* ]]; then
      is_group["$cmd"]=1
      break
    fi
  done
done

for cmd in "${all_cmds[@]}"; do
  # Skip root and group-only commands
  [[ "$cmd" == "recurly" ]] && continue
  [[ -n "${is_group[$cmd]:-}" ]] && continue

  if ! echo "$skill_content" | grep -qF "$cmd"; then
    add_drift "UNDOCUMENTED: ${cmd}"
  fi
done

# ── Report ───────────────────────────────────────────────────────────────────

if [[ $total_drift -eq 0 ]]; then
  echo "No skill drift detected."
  exit 0
fi

if [[ $unbaselined_drift -eq 0 ]]; then
  echo "All drift is baselined. OK."
  exit 0
fi

# Print sorted drift entries
printf '%s\n' "${!drift_unbaselined[@]}" | sort

echo ""
echo "Found ${unbaselined_drift} unbaselined drift issue(s)."
echo "Fix the drift or add entries to ${BASELINE}."
exit 1
