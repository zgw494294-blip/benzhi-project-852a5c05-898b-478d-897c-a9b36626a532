package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type VerifySiteInput struct {
	WorkZoneReady        bool
	MachineryAccessReady bool
	TemporaryCareReady   bool
	WeatherWindowSafe    bool
	Notes                string
	VerifiedBy           string
	Now                  time.Time
}

func (c *RelocationCase) VerifySite(in VerifySiteInput) ([]Event, error) {
	if c.Status != StatusSiteVerification {
		return nil, Conflict("当前案卷尚不可进行现场核验")
	}
	if len(c.OpenBlockerIDs()) > 0 {
		return nil, Conflict("仍有阻断项未关闭")
	}
	if !in.WorkZoneReady || !in.MachineryAccessReady || !in.TemporaryCareReady || !in.WeatherWindowSafe {
		return nil, Invalid("SITE_CHECKLIST", "作业区、机械通道、临时养护和天气窗口必须全部通过")
	}
	if len(strings.TrimSpace(in.Notes)) > 1000 {
		return nil, Invalid("SITE_NOTES", "现场核验备注不得超过 1000 个字符")
	}
	if err := ValidateActor(in.VerifiedBy); err != nil {
		return nil, err
	}
	plan := c.Plans[c.CurrentPlanRevision]
	if plan.PreparedBy == in.VerifiedBy {
		return nil, Invalid("ROLE_SEPARATION", "现场核验人不得与方案编制人相同")
	}
	verification := SiteVerification{WorkZoneReady: in.WorkZoneReady, MachineryAccessReady: in.MachineryAccessReady, TemporaryCareReady: in.TemporaryCareReady, WeatherWindowSafe: in.WeatherWindowSafe, Notes: in.Notes, VerifiedBy: in.VerifiedBy, VerifiedAt: in.Now.UTC()}
	b, err := json.Marshal(verification)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	e, err := NewEvent(EventSiteVerificationPassed, c.CaseID, in.VerifiedBy, in.Now, SiteVerificationPassedData{Verification: verification, PlanRevision: c.CurrentPlanRevision, ChecklistDigest: hex.EncodeToString(sum[:])})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}
