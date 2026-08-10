package goapitosdk

import "testing"

func TestParsePlanTier(t *testing.T) {
	if ParsePlanTier("paid+") != PlanTierPaidPlus {
		t.Fatal("paid+")
	}
	if !PlanAtLeast("paid", "free") {
		t.Fatal("atLeast")
	}
	if PlanAtLeast("free", "paid") {
		t.Fatal("free < paid")
	}
}

func TestCanFromSnapshot(t *testing.T) {
	snap := &EffectivePermissionsSnapshot{
		PlanClamped: true,
		APIPermissions: map[string]*EffectiveModelPermission{
			"student": {Read: "all", Create: "all", Update: "all", Delete: "none"},
			"staff":   {Read: "all", Create: "none", Update: "none", Delete: "none"},
		},
	}
	if !CanFromSnapshot(snap, "student", "list") {
		t.Fatal("student list")
	}
	if CanFromSnapshot(snap, "staff", "create") {
		t.Fatal("staff create should be denied")
	}
}

func TestIsPlanQuotaError(t *testing.T) {
	if !IsPlanQuotaError(fmtError("plan quota exceeded: max_records.student")) {
		t.Fatal("quota")
	}
}

type fmtError string

func (e fmtError) Error() string { return string(e) }
