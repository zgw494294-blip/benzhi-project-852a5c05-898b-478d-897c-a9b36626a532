package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"heritage-tree-relocation-clearance/internal/application"
	"heritage-tree-relocation-clearance/internal/domain"
)

type selfcheckClient struct {
	baseURL string
	client  *http.Client
	counter int
}

// RunSelfCheck 通过真实 HTTP 连接执行包含阻断整改的完整放行流程。
func RunSelfCheck(ctx context.Context, baseURL string) error {
	client := &selfcheckClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 3 * time.Second}}
	if err := client.get(ctx, "/healthz", http.StatusOK, &map[string]any{}); err != nil {
		return fmt.Errorf("健康检查: %w", err)
	}
	caseID := fmt.Sprintf("SELF-CHECK-%d", time.Now().UnixNano())
	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	create := map[string]any{"caseID": caseID, "treeCode": "TREE-SC-001", "protectionGrade": "I", "species": "香樟", "trunkDiameterCM": 80, "crownRadiusM": 6.5, "plannedWindowStart": start, "plannedWindowEnd": start.Add(48 * time.Hour), "createdBy": "selfcheck-creator", "idempotencyKey": "selfcheck-create"}
	var mutation application.MutationResult
	if err := client.post(ctx, "/api/v1/relocation-cases", create, http.StatusCreated, &mutation); err != nil {
		return fmt.Errorf("创建案卷: %w", err)
	}
	if mutation.Status != domain.StatusPreparing || mutation.Version != 1 {
		return errors.New("创建案卷后的状态或版本不符合预期")
	}
	for index, sector := range domain.RequiredSectors() {
		payload := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": fmt.Sprintf("selfcheck-survey-%d", index), "surveyID": fmt.Sprintf("SC-SURVEY-%s", sector), "sector": sector, "probeDepthCM": 90, "criticalRootCount": 5, "exposedRootRatio": 0.2, "soilMoisturePercent": 30, "evidenceRefs": []string{"photo://selfcheck/" + sector}, "recordedBy": "selfcheck-surveyor"}
		if err := client.post(ctx, "/api/v1/relocation-cases/"+caseID+"/root-surveys", payload, http.StatusCreated, &mutation); err != nil {
			return fmt.Errorf("登记 %s 方位勘查: %w", sector, err)
		}
	}
	if mutation.Status != domain.StatusPlanning {
		return errors.New("勘查完整后未进入方案编制状态")
	}
	plan := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": "selfcheck-plan-01", "planID": "SC-PLAN-01", "cutBoundaryCM": 500, "rootBallRadiusCM": 300, "rootBallDepthCM": 120, "supportMethod": "四向钢索支撑并设置软质衬垫", "moistureMethod": "土球包覆保湿布并定时雾化", "maxTransportMinutes": 420, "monitoringThresholds": map[string]float64{"rootVibrationMM": 10, "soilMoistureMinPercent": 30, "tiltDegrees": 4}, "preparedBy": "selfcheck-planner"}
	if err := client.post(ctx, "/api/v1/relocation-cases/"+caseID+"/protection-plans", plan, http.StatusCreated, &mutation); err != nil {
		return fmt.Errorf("编制保护方案: %w", err)
	}
	review := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": "selfcheck-risk-review", "reviewedBy": "selfcheck-risk-reviewer"}
	if err := client.post(ctx, "/api/v1/relocation-cases/"+caseID+"/risk-reviews", review, http.StatusOK, &mutation); err != nil {
		return fmt.Errorf("执行风险审查: %w", err)
	}
	if mutation.Status != domain.StatusRemediation {
		return errors.New("风险规则未产生预期阻断状态")
	}
	var caseState domain.RelocationCase
	if err := client.get(ctx, "/api/v1/relocation-cases/"+caseID, http.StatusOK, &caseState); err != nil {
		return fmt.Errorf("查询待整改案卷: %w", err)
	}
	blockerIDs := make([]string, 0)
	for id, finding := range caseState.Findings {
		if finding.Severity == domain.SeverityBlocker {
			blockerIDs = append(blockerIDs, id)
		}
	}
	sort.Strings(blockerIDs)
	if len(blockerIDs) == 0 {
		return errors.New("自检方案未生成阻断发现项")
	}
	for index, findingID := range blockerIDs {
		submission := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": fmt.Sprintf("selfcheck-remediate-%d", index), "evidence": []string{"已补充专项措施和现场照片证据"}, "submittedBy": "selfcheck-remediator"}
		path := "/api/v1/relocation-cases/" + caseID + "/findings/" + findingID + "/remediations"
		if err := client.post(ctx, path, submission, http.StatusOK, &mutation); err != nil {
			return fmt.Errorf("提交整改 %s: %w", findingID, err)
		}
		decision := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": fmt.Sprintf("selfcheck-decision-%d", index), "reviewedBy": "selfcheck-remediation-reviewer", "decision": "ACCEPT"}
		path = "/api/v1/relocation-cases/" + caseID + "/findings/" + findingID + "/reviews"
		if err := client.post(ctx, path, decision, http.StatusOK, &mutation); err != nil {
			return fmt.Errorf("复核整改 %s: %w", findingID, err)
		}
	}
	if mutation.Status != domain.StatusSiteVerification {
		return errors.New("全部阻断项关闭后未进入现场核验")
	}
	site := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": "selfcheck-site-check", "workZoneReady": true, "machineryAccessReady": true, "temporaryCareReady": true, "weatherWindowSafe": true, "notes": "自检现场条件均满足", "verifiedBy": "selfcheck-site-reviewer"}
	if err := client.post(ctx, "/api/v1/relocation-cases/"+caseID+"/site-verifications", site, http.StatusOK, &mutation); err != nil {
		return fmt.Errorf("现场核验: %w", err)
	}
	if mutation.Status != domain.StatusFrozen {
		return errors.New("现场核验后方案未冻结")
	}
	issue := map[string]any{"expectedVersion": mutation.Version, "idempotencyKey": "selfcheck-credential", "issuedBy": "selfcheck-issuer"}
	var issueResponse struct {
		Credential domain.ClearanceCredential `json:"credential"`
		Mutation   application.MutationResult `json:"mutation"`
	}
	if err := client.post(ctx, "/api/v1/relocation-cases/"+caseID+"/credentials", issue, http.StatusCreated, &issueResponse); err != nil {
		return fmt.Errorf("签发凭据: %w", err)
	}
	if issueResponse.Mutation.Status != domain.StatusCleared || len(issueResponse.Credential.ContentDigest) != 64 {
		return errors.New("签发结果或内容摘要无效")
	}
	var queried domain.ClearanceCredential
	if err := client.get(ctx, "/api/v1/clearance-credentials/"+issueResponse.Credential.CredentialID, http.StatusOK, &queried); err != nil {
		return fmt.Errorf("查询凭据: %w", err)
	}
	if queried.ContentDigest != issueResponse.Credential.ContentDigest || queried.SerialNumber == 0 {
		return errors.New("查询所得凭据与签发结果不一致")
	}
	var audit application.AuditView
	if err := client.get(ctx, "/api/v1/relocation-cases/"+caseID+"/audit", http.StatusOK, &audit); err != nil {
		return fmt.Errorf("查询审计轨迹: %w", err)
	}
	if audit.Case.Status != domain.StatusCleared || audit.Credential == nil || len(audit.Timeline) < 10 {
		return errors.New("审计轨迹未覆盖完整放行流程")
	}
	var credentialAudit application.AuditView
	if err := client.get(ctx, "/api/v1/clearance-credentials/"+queried.CredentialID+"/audit", http.StatusOK, &credentialAudit); err != nil {
		return fmt.Errorf("按凭据查询审计轨迹: %w", err)
	}
	if credentialAudit.Case.CaseID != caseID || len(credentialAudit.Timeline) != len(audit.Timeline) {
		return errors.New("按凭据查询的审计轨迹不完整")
	}
	return nil
}

func (c *selfcheckClient) post(ctx context.Context, path string, body any, expectedStatus int, target any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	c.counter++
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", fmt.Sprintf("selfcheck-%03d", c.counter))
	return c.do(request, expectedStatus, target)
}

func (c *selfcheckClient) get(ctx context.Context, path string, expectedStatus int, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.counter++
	request.Header.Set("X-Request-ID", fmt.Sprintf("selfcheck-%03d", c.counter))
	return c.do(request, expectedStatus, target)
}

func (c *selfcheckClient) do(request *http.Request, expectedStatus int, target any) error {
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return err
	}
	if response.StatusCode != expectedStatus {
		return fmt.Errorf("HTTP %d，期望 %d，响应 %s", response.StatusCode, expectedStatus, strings.TrimSpace(string(payload)))
	}
	if target != nil && len(payload) > 0 {
		if err := json.Unmarshal(payload, target); err != nil {
			return fmt.Errorf("解析响应: %w", err)
		}
	}
	return nil
}
