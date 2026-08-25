package domain

import (
	"regexp"
	"strings"
	"time"
)

var treeCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]{2,39}$`)

func ValidateCaseProfile(treeCode string, grade ProtectionGrade, species string, diameter, crown float64, start, end time.Time) error {
	if !treeCodePattern.MatchString(treeCode) {
		return Invalid("TREE_CODE", "古树编号应为 3 至 40 位大写字母、数字或连字符")
	}
	if grade != GradeOne && grade != GradeTwo && grade != GradeThree {
		return Invalid("PROTECTION_GRADE", "保护等级必须为 I、II 或 III")
	}
	if len(strings.TrimSpace(species)) < 2 || len(species) > 100 {
		return Invalid("SPECIES", "树种名称长度应为 2 至 100 个字符")
	}
	if diameter < 20 || diameter > 1000 {
		return Invalid("TRUNK_DIAMETER", "胸径必须处于 20 至 1000 厘米之间")
	}
	if crown <= 0 || crown > 100 {
		return Invalid("CROWN_RADIUS", "冠幅半径必须处于 0 至 100 米之间")
	}
	if start.IsZero() || end.IsZero() || !end.After(start) {
		return Invalid("PLANNED_WINDOW", "计划移栽窗口起止时间无效")
	}
	if end.Sub(start) > 30*24*time.Hour {
		return Invalid("PLANNED_WINDOW", "计划移栽窗口不得超过 30 天")
	}
	return nil
}

func ValidateActor(actor string) error {
	actor = strings.TrimSpace(actor)
	if len(actor) < 2 || len(actor) > 80 {
		return Invalid("ACTOR", "责任人标识长度应为 2 至 80 个字符")
	}
	return nil
}

func RequiredSectors() []string { return []string{"N", "E", "S", "W"} }

func validSector(sector string) bool {
	for _, wanted := range RequiredSectors() {
		if sector == wanted {
			return true
		}
	}
	return false
}
