package pricing

import (
	"testing"
)

func TestModelMatch_Exact(t *testing.T) {
	if !modelMatch("gpt-4o", "gpt-4o") {
		t.Fatal("expected match")
	}
	if modelMatch("gpt-4o", "gpt-4o-mini") {
		t.Fatal("expected no match")
	}
}

func TestModelMatch_WildcardSuffix(t *testing.T) {
	if !modelMatch("gpt-4*", "gpt-4o") {
		t.Fatal("expected match")
	}
	if !modelMatch("gpt-4*", "gpt-4-turbo") {
		t.Fatal("expected match")
	}
	if modelMatch("gpt-4*", "gpt-3.5") {
		t.Fatal("expected no match")
	}
}

func TestModelMatch_AllStar(t *testing.T) {
	if !modelMatch("*", "anything") {
		t.Fatal("'*' should match all")
	}
}

func TestMatch_NoRule_ReturnsNil(t *testing.T) {
	s, _ := newTestService(t, false)
	if r := s.matchRule("gpt-4o", 1); r != nil {
		t.Fatalf("expected nil, got %+v", r)
	}
}

func TestMatch_GlobalOnly(t *testing.T) {
	s, _ := newTestService(t, false)
	_, err := s.Create(ctx(), CreateInput{
		Scope:      ScopeGlobal,
		InputRatio: ratioPtr(1.0), OutputRatio: ratioPtr(2.0),
		Priority: 100,
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	r := s.matchRule("gpt-4o", 1)
	if r == nil || r.Scope != ScopeGlobal {
		t.Fatalf("expected global rule, got %+v", r)
	}
}

func TestMatch_GroupBeatsGlobal(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope:      ScopeGlobal,
		InputRatio: ratioPtr(1.0), OutputRatio: ratioPtr(2.0),
		Priority: 100,
	}, 0)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeGroup, GroupID: idPtr(7),
		InputRatio: ratioPtr(0.5),
		Priority:   100,
	}, 0)
	r := s.matchRule("gpt-4o", 7)
	if r == nil || r.Scope != ScopeGroup {
		t.Fatalf("expected group rule for group=7, got %+v", r)
	}
	// group_id 不同 → 退回 global
	r2 := s.matchRule("gpt-4o", 99)
	if r2 == nil || r2.Scope != ScopeGlobal {
		t.Fatalf("expected fallback to global, got %+v", r2)
	}
}

func TestMatch_ModelBeatsGroup(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeGroup, GroupID: idPtr(1), InputRatio: ratioPtr(0.5), Priority: 100,
	}, 0)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0), Priority: 100,
	}, 0)
	r := s.matchRule("gpt-4o", 1)
	if r == nil || r.Scope != ScopeModel {
		t.Fatalf("expected model rule, got %+v", r)
	}
}

func TestMatch_GroupModelBeatsAll(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0), Priority: 100,
	}, 0)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeGroupModel, GroupID: idPtr(2), Model: strPtr("gpt-4o"),
		InputRatio: ratioPtr(0.3), Priority: 100,
	}, 0)
	r := s.matchRule("gpt-4o", 2)
	if r == nil || r.Scope != ScopeGroupModel {
		t.Fatalf("expected group_model rule, got %+v", r)
	}
}

func TestMatch_LowerPriorityNumberWins(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(1.0), Priority: 100,
	}, 0)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0), Priority: 50,
	}, 0)
	r := s.matchRule("gpt-4o", 0)
	if r == nil || r.Priority != 50 {
		t.Fatalf("expected priority=50 rule, got %+v", r)
	}
}

func TestMatch_SamePriority_HigherIDWins(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(1.0), Priority: 100,
	}, 0)
	r2, _ := s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0), Priority: 100,
	}, 0)
	r := s.matchRule("gpt-4o", 0)
	if r == nil || r.ID != r2.ID {
		t.Fatalf("expected r.ID=%d, got %+v", r2.ID, r)
	}
}

func TestMatch_WildcardModel_Matches(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4*"), InputRatio: ratioPtr(2.0), Priority: 100,
	}, 0)
	if r := s.matchRule("gpt-4-turbo", 0); r == nil {
		t.Fatal("expected match via wildcard")
	}
	if r := s.matchRule("gpt-3.5", 0); r != nil {
		t.Fatal("expected no match")
	}
}

func TestMatch_DisabledRule_Skipped(t *testing.T) {
	s, _ := newTestService(t, false)
	r, _ := s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0), Priority: 100,
	}, 0)
	// 直接改成 disabled
	st := RuleStatusDisabled
	_, err := s.Update(ctx(), r.ID, UpdatePatch{Status: &st}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.matchRule("gpt-4o", 0); got != nil {
		t.Fatalf("expected no match (rule disabled), got %+v", got)
	}
}
