package domain

import (
	"math"
	"strings"
	"time"
)

type RecordSurveyInput struct {
	SurveyID            string
	Sector              string
	ProbeDepthCM        float64
	CriticalRootCount   int
	ExposedRootRatio    float64
	SoilMoisturePercent float64
	EvidenceRefs        []string
	RecordedBy          string
	Now                 time.Time
}

func (c *RelocationCase) RecordSurvey(in RecordSurveyInput) ([]Event, error) {
	if c.Status != StatusPreparing {
		return nil, Conflict("仅资料准备状态可登记根系勘查")
	}
	if in.SurveyID == "" {
		return nil, Invalid("SURVEY_ID", "勘查记录编号不能为空")
	}
	if !validSector(in.Sector) {
		return nil, Invalid("SECTOR", "方位必须为 N、E、S、W 之一")
	}
	if _, exists := c.Surveys[in.Sector]; exists {
		return nil, Conflict("该方位已登记勘查记录")
	}
	if in.ProbeDepthCM < 20 || in.ProbeDepthCM > 300 {
		return nil, Invalid("PROBE_DEPTH", "探查深度必须处于 20 至 300 厘米之间")
	}
	if in.CriticalRootCount < 0 || in.CriticalRootCount > 1000 {
		return nil, Invalid("CRITICAL_ROOT_COUNT", "关键根数量无效")
	}
	if in.ExposedRootRatio < 0 || in.ExposedRootRatio > 1 {
		return nil, Invalid("EXPOSED_ROOT_RATIO", "关键根暴露比例必须处于 0 至 1 之间")
	}
	if in.SoilMoisturePercent < 0 || in.SoilMoisturePercent > 100 {
		return nil, Invalid("SOIL_MOISTURE", "土壤含水率必须处于 0 至 100 之间")
	}
	if len(in.EvidenceRefs) == 0 {
		return nil, Invalid("EVIDENCE", "每个方位至少提供一项证据引用")
	}
	for _, ref := range in.EvidenceRefs {
		if strings.TrimSpace(ref) == "" || len(ref) > 240 {
			return nil, Invalid("EVIDENCE", "证据引用不能为空且不得超过 240 个字符")
		}
	}
	if err := ValidateActor(in.RecordedBy); err != nil {
		return nil, err
	}
	recommended := c.TrunkDiameterCM*5 + float64(in.CriticalRootCount)*2
	if in.ExposedRootRatio > 0.35 {
		recommended *= 1.15
	}
	recommended = math.Round(recommended*10) / 10
	survey := RootSurvey{SurveyID: in.SurveyID, CaseID: c.CaseID, Sector: in.Sector, ProbeDepthCM: in.ProbeDepthCM, CriticalRootCount: in.CriticalRootCount, ExposedRootRatio: in.ExposedRootRatio, SoilMoisturePercent: in.SoilMoisturePercent, EvidenceRefs: append([]string(nil), in.EvidenceRefs...), RecommendedRootBallRadiusCM: recommended, RecordedBy: in.RecordedBy, RecordedAt: in.Now.UTC()}
	coverage := len(c.Surveys)+1 == len(RequiredSectors())
	evidenceCount := 0
	for _, existing := range c.Surveys {
		if len(existing.EvidenceRefs) > 0 {
			evidenceCount++
		}
	}
	evidenceCount++
	completeness := float64(evidenceCount) / float64(len(RequiredSectors()))
	e, err := NewEvent(EventRootSurveyRecorded, c.CaseID, in.RecordedBy, in.Now, RootSurveyRecordedData{Survey: survey, CoverageComplete: coverage, EvidenceCompleteness: completeness})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

func (c *RelocationCase) SurveyCoverage() (float64, float64) {
	coverage := float64(len(c.Surveys)) / float64(len(RequiredSectors()))
	withEvidence := 0
	for _, sector := range RequiredSectors() {
		if survey, ok := c.Surveys[sector]; ok && len(survey.EvidenceRefs) > 0 {
			withEvidence++
		}
	}
	return coverage, float64(withEvidence) / float64(len(RequiredSectors()))
}

func (c *RelocationCase) RecommendedRootBallRadius() float64 {
	maximum := float64(0)
	for _, survey := range c.Surveys {
		if survey.RecommendedRootBallRadiusCM > maximum {
			maximum = survey.RecommendedRootBallRadiusCM
		}
	}
	return maximum
}
