package domain

import (
	"strings"
	"time"
)

type IssueCredentialInput struct {
	CredentialID  string
	SerialNumber  uint64
	EventSequence uint64
	ContentDigest string
	IssuedBy      string
	Now           time.Time
}

func (c *RelocationCase) IssueCredential(in IssueCredentialInput) ([]Event, error) {
	if c.Status != StatusFrozen {
		return nil, Conflict("仅已冻结案卷可签发放行凭据")
	}
	if c.Credential != nil {
		return nil, Conflict("案卷已经签发放行凭据")
	}
	if in.CredentialID == "" || in.SerialNumber == 0 || in.EventSequence == 0 {
		return nil, Invalid("CREDENTIAL", "凭据编号、序号和事件边界不能为空")
	}
	if len(in.ContentDigest) != 64 || strings.Trim(in.ContentDigest, "0123456789abcdef") != "" {
		return nil, Invalid("CONTENT_DIGEST", "内容摘要必须为小写 SHA-256 十六进制值")
	}
	if err := ValidateActor(in.IssuedBy); err != nil {
		return nil, err
	}
	credential := ClearanceCredential{CredentialID: in.CredentialID, CaseID: c.CaseID, SerialNumber: in.SerialNumber, FrozenPlanRevision: c.FrozenPlanRevision, SiteChecklistDigest: c.SiteChecklistDigest, EventSequence: in.EventSequence, ContentDigest: in.ContentDigest, IssuedBy: in.IssuedBy, IssuedAt: in.Now.UTC()}
	e, err := NewEvent(EventClearanceCredentialIssued, c.CaseID, in.IssuedBy, in.Now, ClearanceCredentialIssuedData{Credential: credential})
	if err != nil {
		return nil, err
	}
	return []Event{e}, nil
}
