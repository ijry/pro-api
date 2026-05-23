package payment

import "testing"

func TestHolderFrom_Empty_ReturnsNewHolder(t *testing.T) {
	h := HolderFrom(nil)
	if h == nil {
		t.Fatal("want non-nil")
	}
	if h.Manual != nil || h.Redeem != nil {
		t.Fatal("want empty fields")
	}
}

func TestHolderFrom_ExistingHolder_ReturnsIt(t *testing.T) {
	original := &Holder{Manual: "M", Redeem: "R"}
	h := HolderFrom(original)
	if h != original {
		t.Fatal("want same holder")
	}
}

func TestHolderFrom_WrongType_ReturnsNewHolder(t *testing.T) {
	h := HolderFrom("not-a-holder")
	if h == nil {
		t.Fatal("want non-nil")
	}
	if h.Manual != nil || h.Redeem != nil {
		t.Fatal("want empty fields")
	}
}

func TestHolder_FieldsAccessible(t *testing.T) {
	h := &Holder{}
	h.Manual = "manual-svc"
	h.Redeem = "redeem-svc"
	if h.Manual != "manual-svc" || h.Redeem != "redeem-svc" {
		t.Fatal("fields not assignable")
	}
}
