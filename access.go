package goapitosdk

import (
	"fmt"
	"strings"
)

// PlanTierKnown are engine pro_tenants.plan_tier slugs.
const (
	PlanTierFree     = "free"
	PlanTierPaid     = "paid"
	PlanTierPaidPlus = "paid_plus"
	PlanTierUltra    = "ultra"
)

// ParsePlanTier normalizes unknown/empty plan strings to a known tier (default free).
func ParsePlanTier(raw string) string {
	key := strings.ToLower(strings.TrimSpace(raw))
	switch key {
	case "paid":
		return PlanTierPaid
	case "paid_plus", "paid+", "paidplus":
		return PlanTierPaidPlus
	case "ultra":
		return PlanTierUltra
	case "free", "":
		return PlanTierFree
	default:
		return PlanTierFree
	}
}

func planRank(tier string) int {
	switch ParsePlanTier(tier) {
	case PlanTierPaid:
		return 1
	case PlanTierPaidPlus:
		return 2
	case PlanTierUltra:
		return 3
	default:
		return 0
	}
}

// PlanAtLeast reports whether tier meets or exceeds min.
func PlanAtLeast(tier, min string) bool {
	return planRank(tier) >= planRank(min)
}

// ScopeAllows is true when a CRUD scope is not empty/none.
func ScopeAllows(scope string) bool {
	s := strings.ToLower(strings.TrimSpace(scope))
	return s != "" && s != "none"
}

func scopeForAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "list", "show", "read":
		return "read"
	case "create":
		return "create"
	case "edit", "update":
		return "update"
	case "delete":
		return "delete"
	default:
		return strings.ToLower(strings.TrimSpace(action))
	}
}

// CanFromSnapshot checks api_permissions for a Refine-style action.
func CanFromSnapshot(snap *EffectivePermissionsSnapshot, resource, action string) bool {
	if snap == nil {
		return false
	}
	if snap.IsAdmin && !snap.PlanClamped {
		return true
	}
	if snap.APIPermissions == nil {
		return false
	}
	model := snap.APIPermissions[resource]
	if model == nil {
		model = snap.APIPermissions[strings.ReplaceAll(resource, "-", "_")]
	}
	if model == nil {
		return false
	}
	switch scopeForAction(action) {
	case "read":
		return ScopeAllows(model.Read)
	case "create":
		return ScopeAllows(model.Create)
	case "update":
		return ScopeAllows(model.Update)
	case "delete":
		return ScopeAllows(model.Delete)
	default:
		return false
	}
}

// IsPlanQuotaError detects engine plan quota create failures.
func IsPlanQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "plan quota exceeded") || strings.Contains(msg, "max_records.")
}

// ParseEffectivePermissionsSnapshot maps a GraphQL data map to the typed snapshot.
func ParseEffectivePermissionsSnapshot(raw map[string]interface{}) (*EffectivePermissionsSnapshot, error) {
	if raw == nil {
		return nil, fmt.Errorf("nil myEffectivePermissions payload")
	}
	out := &EffectivePermissionsSnapshot{
		APIPermissions: map[string]*EffectiveModelPermission{},
		Quotas:         map[string]float64{},
		Usage:          map[string]float64{},
	}
	if v, ok := raw["plan_slug"].(string); ok {
		out.PlanSlug = v
	}
	if v, ok := raw["role_id"].(string); ok {
		out.RoleID = v
	}
	if v, ok := raw["plan_clamped"].(bool); ok {
		out.PlanClamped = v
	}
	if v, ok := raw["is_admin"].(bool); ok {
		out.IsAdmin = v
	}
	if perms, ok := raw["api_permissions"].(map[string]interface{}); ok {
		for k, vv := range perms {
			m, ok := vv.(map[string]interface{})
			if !ok {
				continue
			}
			ap := &EffectiveModelPermission{}
			if s, ok := m["read"].(string); ok {
				ap.Read = s
			}
			if s, ok := m["create"].(string); ok {
				ap.Create = s
			}
			if s, ok := m["update"].(string); ok {
				ap.Update = s
			}
			if s, ok := m["delete"].(string); ok {
				ap.Delete = s
			}
			if b, ok := m["grace"].(bool); ok {
				ap.Grace = b
			}
			out.APIPermissions[k] = ap
		}
	}
	if q, ok := raw["quotas"].(map[string]interface{}); ok {
		for k, vv := range q {
			switch n := vv.(type) {
			case float64:
				out.Quotas[k] = n
			case int:
				out.Quotas[k] = float64(n)
			}
		}
	}
	if u, ok := raw["usage"].(map[string]interface{}); ok {
		for k, vv := range u {
			switch n := vv.(type) {
			case float64:
				out.Usage[k] = n
			case int:
				out.Usage[k] = float64(n)
			}
		}
	}
	if arr, ok := raw["grace_models"].([]interface{}); ok {
		for _, it := range arr {
			if s, ok := it.(string); ok && strings.TrimSpace(s) != "" {
				out.GraceModels = append(out.GraceModels, s)
			}
		}
	}
	if arr, ok := raw["logic_executions"].([]interface{}); ok {
		for _, it := range arr {
			if s, ok := it.(string); ok && strings.TrimSpace(s) != "" {
				out.LogicExecutions = append(out.LogicExecutions, s)
			}
		}
	}
	return out, nil
}
