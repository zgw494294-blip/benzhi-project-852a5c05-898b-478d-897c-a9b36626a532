package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type RevisePlanInput struct {
	PlanID               string
	CutBoundaryCM        float64
	RootBallRadiusCM     float64
	RootBallDepthCM      float64
	SupportMethod        string
	MoistureMethod       string
	MaxTransportMinutes  int
	MonitoringThresholds map[string]float64
	PreparedBy           string
	Now                  time.Time
}

func (c *RelocationCase) RevisePlan(in RevisePlanInput) ([]Event, error) {
	if c.Status != StatusPlanning {
		return nil, Conflict("仅方案编制状态可修订保护方案")
	}
	coverage, evidence := c.SurveyCoverage()
	if coverage < 1 || evidence < 1 {
		return nil, Conflict("四个方位勘查及证据完整后方可编制方案")
	}
	if in.PlanID == "" {
		return nil, Invalid("PLAN_ID", "方案编号不能为空")
	}
	if in.CutBoundaryCM < c.TrunkDiameterCM*2 || in.CutBoundaryCM > 3000 {
		return nil, Invalid("CUT_BOUNDARY", "断根边界超出允许范围")
	}
	if in.RootBallRadiusCM < c.TrunkDiameterCM*2 || in.RootBallRadiusCM > 3000 {
		return nil, Invalid("ROOT_BALL_RADIUS", "土球半径超出允许范围")
	}
	if in.RootBallDepthCM < 40 || in.RootBallDepthCM > 500 {
		return nil, Invalid("ROOT_BALL_DEPTH", "土球深度必须处于 40 至 500 厘米之间")
	}
	if len(strings.TrimSpace(in.SupportMethod)) < 3 || len(in.SupportMethod) > 300 {
		return nil, Invalid("SUPPORT_METHOD", "支撑措施说明长度无效")
	}
	if len(strings.TrimSpace(in.MoistureMethod)) < 3 || len(in.MoistureMethod) > 300 {
		return nil, Invalid("MOISTURE_METHOD", "保湿措施说明长度无效")
	}
	if in.MaxTransportMinutes < 10 || in.MaxTransportMinutes > 1440 {
		return nil, Invalid("TRANSPORT_LIMIT", "运输时限必须处于 10 至 1440 分钟之间")
	}
	requiredThresholds := []string{"rootVibrationMM", "soilMoistureMinPercent", "tiltDegrees"}
	for _, key := range requiredThresholds {
		value, ok := in.MonitoringThresholds[key]
		if !ok || value <= 0 {
			return nil, Invalid("MONITORING_THRESHOLDS", "缺少有效监测阈值 "+key)
		}
	}
	if err := ValidateActor(in.PreparedBy); err != nil {
		return nil, err
	}
	revision := c.CurrentPlanRevision + 1
	digestInput := struct {
		Revision           int
		Cut, Radius, Depth float64
		Support, Moisture  string
		Transport          int
		Thresholds         map[string]float64
	}{revision, in.CutBoundaryCM, in.RootBallRadiusCM, in.RootBallDepthCM, in.SupportMethod, in.MoistureMethod, in.MaxTransportMinutes, in.MonitoringThresholds}
	b, err := json.Marshal(digestInput)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	plan := ProtectionPlan{PlanID: in.PlanID, CaseID: c.CaseID, Revision: revision, CutBoundaryCM: in.CutBoundaryCM, RootBallRadiusCM: in.RootBallRadiusCM, RootBallDepthCM: in.RootBallDepthCM, SupportMethod: in.SupportMethod, MoistureMethod: in.MoistureMethod, MaxTransportMinutes: in.MaxTransportMinutes, MonitoringThresholds: cloneThresholds(in.MonitoringThresholds), InputDigest: hex.EncodeToString(sum[:]), PreparedBy: in.PreparedBy, PreparedAt: in.Now.UTC()}
	e, err := NewEvent(EventProtectionPlanRevised, c.CaseID, in.PreparedBy, in.Now, ProtectionPlanRevisedData{Plan: plan})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

func cloneThresholds(source map[string]float64) map[string]float64 {
	copyMap := make(map[string]float64, len(source))
	for key, value := range source {
		copyMap[key] = value
	}
	return copyMap
}
