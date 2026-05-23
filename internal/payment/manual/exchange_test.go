package manual

import "testing"

func TestComputeQuota_Standard100Yuan_AtRate7(t *testing.T) {
	// 100 元 / 7.0 * 500000 = 7142857.14... → 7142857
	got := ComputeQuota(1_000_000, 7.0, 500_000)
	if got != 7_142_857 {
		t.Fatalf("want 7142857, got %d", got)
	}
}

func TestComputeQuota_LargeAmount_NoOverflow(t *testing.T) {
	// 10000 元 / 7.0 * 500000 = 714285714.28... → 714285714
	got := ComputeQuota(100_000_000, 7.0, 500_000)
	if got != 714_285_714 {
		t.Fatalf("want 714285714, got %d", got)
	}
}

func TestComputeQuota_Zero_ReturnsZero(t *testing.T) {
	if got := ComputeQuota(0, 7.0, 500_000); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestComputeQuota_ZeroRate_ReturnsZero(t *testing.T) {
	if got := ComputeQuota(1_000_000, 0, 500_000); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestComputeQuota_NegativeRate_ReturnsZero(t *testing.T) {
	if got := ComputeQuota(1_000_000, -1, 500_000); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestComputeQuota_ZeroBase_ReturnsZero(t *testing.T) {
	if got := ComputeQuota(1_000_000, 7.0, 0); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestComputeQuota_FloorOnFractional(t *testing.T) {
	// 0.0999 元 / 7.0 * 500000 = 7135.71... → 7135 (floor)
	got := ComputeQuota(999, 7.0, 500_000)
	if got != 7135 {
		t.Fatalf("want 7135, got %d", got)
	}
}

func TestComputeQuota_TinyAmount_FloorsToZero(t *testing.T) {
	// 1 厘 / 7.0 * 500000 = 7.14... → 7
	got := ComputeQuota(1, 7.0, 500_000)
	if got != 7 {
		t.Fatalf("want 7, got %d", got)
	}
}
