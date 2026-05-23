package pricing

import (
	"errors"
	"testing"

	"github.com/ijry/pro-api/pkg/apierr"
)

func TestService_Create_ValidatesScope(t *testing.T) {
	s, _ := newTestService(t, false)
	// global with group_id 不允许
	_, err := s.Create(ctx(), CreateInput{
		Scope: ScopeGlobal, GroupID: idPtr(1), InputRatio: ratioPtr(1.0),
	}, 0)
	if err == nil {
		t.Fatal("expected error for global+group_id")
	}
	// model with group_id 不允许
	_, err = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("x"), GroupID: idPtr(1), InputRatio: ratioPtr(1.0),
	}, 0)
	if err == nil {
		t.Fatal("expected error for model+group_id")
	}
	// group_model 缺 model
	_, err = s.Create(ctx(), CreateInput{
		Scope: ScopeGroupModel, GroupID: idPtr(1), InputRatio: ratioPtr(1.0),
	}, 0)
	if err == nil {
		t.Fatal("expected error for group_model w/o model")
	}
	// scope unknown
	_, err = s.Create(ctx(), CreateInput{
		Scope: "weird", InputRatio: ratioPtr(1.0),
	}, 0)
	if err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestService_Create_RejectsAllNilRatios(t *testing.T) {
	s, _ := newTestService(t, false)
	_, err := s.Create(ctx(), CreateInput{Scope: ScopeGlobal}, 0)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeInvalidParam {
		t.Fatalf("expected CodeInvalidParam, got %v", err)
	}
}

func TestService_Create_RejectsNegativeRatio(t *testing.T) {
	s, _ := newTestService(t, false)
	_, err := s.Create(ctx(), CreateInput{
		Scope: ScopeGlobal, InputRatio: ratioPtr(-1.0),
	}, 0)
	if err == nil {
		t.Fatal("expected error for negative ratio")
	}
}

func TestService_Create_RefreshesCache(t *testing.T) {
	s, _ := newTestService(t, false)
	if len(s.cache.All()) != 0 {
		t.Fatal("cache should be empty initially")
	}
	_, err := s.Create(ctx(), CreateInput{
		Scope: ScopeGlobal, InputRatio: ratioPtr(1.0),
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.cache.All()) != 1 {
		t.Fatalf("expected cache len=1, got %d", len(s.cache.All()))
	}
}

func TestService_Update_ClearInputRatio(t *testing.T) {
	s, _ := newTestService(t, false)
	r, _ := s.Create(ctx(), CreateInput{
		Scope:      ScopeGlobal,
		InputRatio: ratioPtr(1.0), OutputRatio: ratioPtr(2.0),
	}, 0)
	updated, err := s.Update(ctx(), r.ID, UpdatePatch{ClearInput: true}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if updated.InputRatio != nil {
		t.Fatalf("expected nil input_ratio, got %v", *updated.InputRatio)
	}
	if updated.OutputRatio == nil || *updated.OutputRatio != 2.0 {
		t.Fatalf("output_ratio should remain 2.0")
	}
}

func TestService_Update_NotFound(t *testing.T) {
	s, _ := newTestService(t, false)
	_, err := s.Update(ctx(), 99999, UpdatePatch{InputRatio: ratioPtr(2.0)}, 0)
	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Code != apierr.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", err)
	}
}

func TestService_Delete_RemovesAndRefreshes(t *testing.T) {
	s, _ := newTestService(t, false)
	r, _ := s.Create(ctx(), CreateInput{
		Scope: ScopeGlobal, InputRatio: ratioPtr(1.0),
	}, 0)
	if err := s.Delete(ctx(), r.ID, 0); err != nil {
		t.Fatal(err)
	}
	if len(s.cache.All()) != 0 {
		t.Fatal("cache should be empty after delete")
	}
	_, err := s.Get(ctx(), r.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestService_List_FilterByScope(t *testing.T) {
	s, _ := newTestService(t, false)
	_, _ = s.Create(ctx(), CreateInput{Scope: ScopeGlobal, InputRatio: ratioPtr(1.0)}, 0)
	_, _ = s.Create(ctx(), CreateInput{
		Scope: ScopeModel, Model: strPtr("gpt-4o"), InputRatio: ratioPtr(2.0),
	}, 0)
	items, total, err := s.List(ctx(), ListFilter{Scope: ScopeModel})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].Scope != ScopeModel {
		t.Fatalf("filter wrong: total=%d items=%v", total, items)
	}
}
