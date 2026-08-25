package domain

import (
	"strings"
	"time"
)

type SubmitRemediationInput struct {
	FindingID   string
	Evidence    []string
	SubmittedBy string
	Now         time.Time
}

func (c *RelocationCase) SubmitRemediation(in SubmitRemediationInput) ([]Event, error) {
	if c.Status != StatusRemediation {
		return nil, Conflict("仅待整改状态可提交整改证据")
	}
	finding, ok := c.Findings[in.FindingID]
	if !ok {
		return nil, NotFound("发现项", in.FindingID)
	}
	if finding.Severity != SeverityBlocker {
		return nil, Conflict("警告项无需阻断整改复核")
	}
	if finding.Status == FindingAccepted {
		return nil, Conflict("已接受的整改项不可再次提交")
	}
	if len(in.Evidence) == 0 || len(in.Evidence) > 20 {
		return nil, Invalid("REMEDIATION_EVIDENCE", "整改证据应包含 1 至 20 项")
	}
	for _, evidence := range in.Evidence {
		if len(strings.TrimSpace(evidence)) < 3 || len(evidence) > 500 {
			return nil, Invalid("REMEDIATION_EVIDENCE", "整改证据说明长度无效")
		}
	}
	if err := ValidateActor(in.SubmittedBy); err != nil {
		return nil, err
	}
	e, err := NewEvent(EventRemediationSubmitted, c.CaseID, in.SubmittedBy, in.Now, RemediationSubmittedData{FindingID: in.FindingID, Evidence: append([]string(nil), in.Evidence...), SubmittedBy: in.SubmittedBy})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

type ReviewRemediationInput struct {
	FindingID  string
	ReviewedBy string
	Decision   ReviewDecision
	Now        time.Time
}

func (c *RelocationCase) ReviewRemediation(in ReviewRemediationInput) ([]Event, error) {
	if c.Status != StatusRemediation {
		return nil, Conflict("仅待整改状态可复核整改项")
	}
	finding, ok := c.Findings[in.FindingID]
	if !ok {
		return nil, NotFound("发现项", in.FindingID)
	}
	if finding.Status != FindingSubmitted {
		return nil, Conflict("发现项尚无待复核的整改证据")
	}
	if err := ValidateActor(in.ReviewedBy); err != nil {
		return nil, err
	}
	if finding.SubmittedBy == in.ReviewedBy {
		return nil, Invalid("ROLE_SEPARATION", "整改提交人与复核人必须不同")
	}
	if in.Decision != DecisionAccept && in.Decision != DecisionReject {
		return nil, Invalid("REVIEW_DECISION", "复核结论必须为 ACCEPT 或 REJECT")
	}
	allClosed := in.Decision == DecisionAccept
	if allClosed {
		for id, other := range c.Findings {
			if other.Severity == SeverityBlocker && id != in.FindingID && other.Status != FindingAccepted {
				allClosed = false
				break
			}
		}
	}
	e, err := NewEvent(EventRemediationReviewed, c.CaseID, in.ReviewedBy, in.Now, RemediationReviewedData{FindingID: in.FindingID, ReviewedBy: in.ReviewedBy, Decision: in.Decision, ReviewedAt: in.Now.UTC(), AllBlockersClosed: allClosed})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}

func (c *RelocationCase) OpenBlockerIDs() []string {
	ids := make([]string, 0)
	for id, finding := range c.Findings {
		if finding.Severity == SeverityBlocker && finding.Status != FindingAccepted {
			ids = append(ids, id)
		}
	}
	return ids
}
