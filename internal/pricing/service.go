package pricing

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ijry/pro-api/internal/audit"
	"github.com/ijry/pro-api/pkg/apierr"
	"gorm.io/gorm"
)

// List 列规则(分页 + 过滤)。
func (s *service) List(ctx context.Context, filter ListFilter) ([]*Rule, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Size <= 0 {
		filter.Size = 20
	}
	if filter.Size > 100 {
		filter.Size = 100
	}
	q := s.db.WithContext(ctx).Model(&Rule{})
	if filter.Scope != "" {
		q = q.Where("scope = ?", filter.Scope)
	}
	if filter.GroupID != nil {
		q = q.Where("group_id = ?", *filter.GroupID)
	}
	if filter.Model != nil {
		q = q.Where("model = ?", *filter.Model)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apierr.New(apierr.CodeDatabase, err.Error())
	}
	var rows []*Rule
	if err := q.Order("priority ASC, id DESC").
		Offset((filter.Page - 1) * filter.Size).
		Limit(filter.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, apierr.New(apierr.CodeDatabase, err.Error())
	}
	return rows, total, nil
}

// Get 按 id 取(不存在返回 CodeNotFound)。
func (s *service) Get(ctx context.Context, id int64) (*Rule, error) {
	if id <= 0 {
		return nil, apierr.New(apierr.CodeInvalidParam, "invalid id")
	}
	var r Rule
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apierr.New(apierr.CodeNotFound, "pricing rule not found")
		}
		return nil, apierr.New(apierr.CodeDatabase, err.Error())
	}
	return &r, nil
}

// Create 新建规则。
func (s *service) Create(ctx context.Context, in CreateInput, actor int64) (*Rule, error) {
	if err := validateInput(&in); err != nil {
		return nil, err
	}
	now := s.clk.Now().UTC()
	r := &Rule{
		ID:             s.idgen.Generate(),
		Scope:          in.Scope,
		GroupID:        in.GroupID,
		Model:          in.Model,
		InputRatio:     in.InputRatio,
		OutputRatio:    in.OutputRatio,
		CachedRatio:    in.CachedRatio,
		ReasoningRatio: in.ReasoningRatio,
		Priority:       in.Priority,
		Status:         in.Status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if r.Priority == 0 {
		r.Priority = 100
	}
	if err := s.db.WithContext(ctx).Create(r).Error; err != nil {
		return nil, apierr.New(apierr.CodeDatabase, err.Error())
	}
	_ = s.Refresh(ctx)
	s.publishInvalidate(ctx)
	s.logAudit(ctx, "pricing_rule.create", r.ID, actor, r)
	return r, nil
}

// Update 部分字段更新。clear_* 把字段置 NULL。
func (s *service) Update(ctx context.Context, id int64, patch UpdatePatch, actor int64) (*Rule, error) {
	r, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	before := *r
	updates := map[string]any{}
	if patch.InputRatio != nil {
		updates["input_ratio"] = *patch.InputRatio
	}
	if patch.OutputRatio != nil {
		updates["output_ratio"] = *patch.OutputRatio
	}
	if patch.CachedRatio != nil {
		updates["cached_ratio"] = *patch.CachedRatio
	}
	if patch.ReasoningRatio != nil {
		updates["reasoning_ratio"] = *patch.ReasoningRatio
	}
	if patch.ClearInput {
		updates["input_ratio"] = nil
	}
	if patch.ClearOutput {
		updates["output_ratio"] = nil
	}
	if patch.ClearCached {
		updates["cached_ratio"] = nil
	}
	if patch.ClearReasoning {
		updates["reasoning_ratio"] = nil
	}
	if patch.Priority != nil {
		updates["priority"] = *patch.Priority
	}
	if patch.Status != nil {
		updates["status"] = *patch.Status
	}
	if len(updates) == 0 {
		return r, nil
	}
	for _, k := range []string{"input_ratio", "output_ratio", "cached_ratio", "reasoning_ratio"} {
		if v, ok := updates[k]; ok && v != nil {
			if f, ok2 := v.(float64); ok2 && f < 0 {
				return nil, apierr.New(apierr.CodeInvalidParam, "ratio must be >= 0")
			}
		}
	}
	updates["updated_at"] = s.clk.Now().UTC()
	if err := s.db.WithContext(ctx).Model(&Rule{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, apierr.New(apierr.CodeDatabase, err.Error())
	}
	updated, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = s.Refresh(ctx)
	s.publishInvalidate(ctx)
	s.logAuditUpdate(ctx, "pricing_rule.update", id, actor, &before, updated)
	return updated, nil
}

// Delete 物理删除。
func (s *service) Delete(ctx context.Context, id int64, actor int64) error {
	r, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&Rule{}).Error; err != nil {
		return apierr.New(apierr.CodeDatabase, err.Error())
	}
	_ = s.Refresh(ctx)
	s.publishInvalidate(ctx)
	s.logAudit(ctx, "pricing_rule.delete", id, actor, r)
	return nil
}

// logAudit 写一条 audit。
func (s *service) logAudit(ctx context.Context, action string, targetID int64, actor int64, payload any) {
	if s.audit == nil {
		return
	}
	var actorPtr *int64
	if actor != 0 {
		a := actor
		actorPtr = &a
	}
	t := targetID
	after, _ := json.Marshal(payload)
	_ = s.audit.Log(ctx, audit.Entry{
		ActorID:    actorPtr,
		Action:     action,
		TargetType: "pricing_rule",
		TargetID:   &t,
		After:      after,
	})
}

// logAuditUpdate 同 logAudit,但带 before/after 两个快照。
func (s *service) logAuditUpdate(ctx context.Context, action string, targetID int64, actor int64, before, after any) {
	if s.audit == nil {
		return
	}
	var actorPtr *int64
	if actor != 0 {
		a := actor
		actorPtr = &a
	}
	t := targetID
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	_ = s.audit.Log(ctx, audit.Entry{
		ActorID:    actorPtr,
		Action:     action,
		TargetType: "pricing_rule",
		TargetID:   &t,
		Before:     beforeJSON,
		After:      afterJSON,
	})
}

// validateInput 校验 CreateInput。
func validateInput(in *CreateInput) error {
	switch in.Scope {
	case ScopeGlobal:
		if in.GroupID != nil || in.Model != nil {
			return apierr.New(apierr.CodeInvalidParam, "global scope must not set group_id/model")
		}
	case ScopeGroup:
		if in.GroupID == nil || in.Model != nil {
			return apierr.New(apierr.CodeInvalidParam, "group scope requires group_id and forbids model")
		}
	case ScopeModel:
		if in.Model == nil || in.GroupID != nil {
			return apierr.New(apierr.CodeInvalidParam, "model scope requires model and forbids group_id")
		}
	case ScopeGroupModel:
		if in.GroupID == nil || in.Model == nil {
			return apierr.New(apierr.CodeInvalidParam, "group_model scope requires both group_id and model")
		}
	default:
		return apierr.New(apierr.CodeInvalidParam, "unknown scope: "+in.Scope)
	}
	if in.InputRatio == nil && in.OutputRatio == nil &&
		in.CachedRatio == nil && in.ReasoningRatio == nil {
		return apierr.New(apierr.CodeInvalidParam, "at least one ratio must be set")
	}
	for _, r := range []*float64{in.InputRatio, in.OutputRatio, in.CachedRatio, in.ReasoningRatio} {
		if r != nil && *r < 0 {
			return apierr.New(apierr.CodeInvalidParam, "ratio must be >= 0")
		}
	}
	if in.Status != RuleStatusEnabled && in.Status != RuleStatusDisabled {
		return apierr.New(apierr.CodeInvalidParam, "status only 0/1")
	}
	return nil
}
