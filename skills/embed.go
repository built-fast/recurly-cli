package skills

import "embed"

//go:embed recurly/SKILL.md
var FS embed.FS

// SkillMD returns the raw contents of the embedded SKILL.md file.
func SkillMD() ([]byte, error) {
	return FS.ReadFile("recurly/SKILL.md")
}
