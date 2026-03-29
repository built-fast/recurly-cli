#!/usr/bin/env bash
# check-skill-drift.sh — verify that every `recurly <cmd>` and `--<flag>`
# reference in SKILL.md has a corresponding entry in .surface.
#
# Exit 0 if clean, 1 if mismatches found.
# Known/accepted mismatches can be listed in .surface-skill-drift (one per line).

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

# Load baseline exceptions (if file exists)
declare -A exceptions
if [[ -f "$BASELINE" ]]; then
  while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    exceptions["$line"]=1
  done < "$BASELINE"
fi

errors=0

# check_cmd_in_surface tries to match a command string against .surface CMD
# entries. It progressively strips trailing words (which may be example
# arguments like "acct-1" or "sub123") until a CMD match is found.
# Returns 0 if any prefix matches, 1 if none do.
check_cmd_in_surface() {
  local cmd="$1"
  while [[ "$cmd" != "recurly" && -n "$cmd" ]]; do
    if grep -q "^CMD ${cmd}$" "$SURFACE"; then
      return 0
    fi
    # Strip the last word
    cmd="${cmd% *}"
  done
  return 1
}

# Extract `recurly <subcommand>` references from SKILL.md.
while IFS= read -r match; do
  # Normalize: strip everything after flags (--) and trim trailing whitespace
  cmd=$(echo "$match" | sed 's/ --.*//; s/[[:space:]]*$//')

  # Skip bare "recurly"
  [[ "$cmd" == "recurly" ]] && continue

  # Strip placeholder arguments like <account_id>, [flags], etc.
  clean_cmd=$(echo "$cmd" | sed 's/ <[^>]*>//g; s/ \[[^]]*\]//g')
  clean_cmd=$(echo "$clean_cmd" | sed 's/[[:space:]]*$//')

  [[ -z "$clean_cmd" || "$clean_cmd" == "recurly" ]] && continue

  # Check if this command (or a prefix of it) exists in .surface
  if ! check_cmd_in_surface "$clean_cmd"; then
    key="CMD ${clean_cmd}"
    if [[ -z "${exceptions[$key]:-}" ]]; then
      echo "DRIFT: command '${clean_cmd}' in SKILL.md not found in .surface" >&2
      ((errors++))
    fi
  fi
done < <(grep -oE 'recurly [a-z][-a-z0-9 ]*' "$SKILL" | sort -u)

# Extract --flag references from SKILL.md and verify they exist in .surface
while IFS= read -r flag; do
  if ! grep -q -- "--${flag} " "$SURFACE"; then
    key="FLAG --${flag}"
    if [[ -z "${exceptions[$key]:-}" ]]; then
      echo "DRIFT: flag '--${flag}' in SKILL.md not found in .surface" >&2
      ((errors++))
    fi
  fi
done < <(grep -oE '\-\-[a-z][-a-z0-9]*' "$SKILL" | sed 's/^--//' | sort -u)

if [[ $errors -gt 0 ]]; then
  echo "Found ${errors} drift issue(s). Update SKILL.md or .surface-skill-drift." >&2
  exit 1
fi

echo "No skill drift detected."
