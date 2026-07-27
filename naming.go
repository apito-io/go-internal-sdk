package goapitosdk

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	camelPieceRe  = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	canonicalIDRe = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
)

var singularKeepAsIs = map[string]struct{}{
	"news": {}, "data": {}, "media": {}, "analytics": {}, "series": {}, "species": {},
}

func splitCamelPieces(piece string) []string {
	spaced := camelPieceRe.ReplaceAllString(piece, `${1} ${2}`)
	parts := strings.Fields(spaced)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		clean := strings.ToLower(strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return r
			}
			return -1
		}, p))
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func camelFromCanonical(canonical string) string {
	parts := strings.Split(canonical, "_")
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(p))
		} else {
			b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
		}
	}
	return b.String()
}

func pascalFromCanonical(canonical string) string {
	parts := strings.Split(canonical, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + strings.ToLower(p[1:]))
	}
	return b.String()
}

func pascalFromAnyModelID(modelID string) string {
	if modelID == "" {
		return ""
	}
	if strings.Contains(modelID, "_") {
		return pascalFromCanonical(modelID)
	}
	segs := splitCamelPieces(modelID)
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(strings.ToUpper(s[:1]) + strings.ToLower(s[1:]))
	}
	return b.String()
}

// ApitoSingularResourceName matches refine-apito / flutter_admin_sdk naming.
func ApitoSingularResourceName(name string) string {
	t := strings.TrimSpace(name)
	if strings.HasSuffix(t, "ListCount") {
		t = t[:len(t)-len("ListCount")]
	} else if strings.HasSuffix(t, "List") {
		t = t[:len(t)-len("List")]
	}
	t = strings.TrimSpace(t)
	if t == "" {
		return ""
	}
	if strings.Contains(t, "_") {
		return camelFromCanonical(t)
	}
	segs := splitCamelPieces(t)
	if len(segs) == 0 {
		return strings.ToLower(t)
	}
	var b strings.Builder
	for i, s := range segs {
		if i == 0 {
			b.WriteString(strings.ToLower(s))
		} else {
			b.WriteString(strings.ToUpper(s[:1]) + strings.ToLower(s[1:]))
		}
	}
	return b.String()
}

// ApitoMultipleResourceName returns the list GraphQL field name.
func ApitoMultipleResourceName(name string) string {
	return ApitoSingularResourceName(name) + "List"
}

func listGraphQLTypeName(modelID string) string {
	return pascalFromAnyModelID(ApitoSingularResourceName(modelID)) + "List"
}

// ApitoGraphQLComposedTypeName builds composed payload type names.
func ApitoGraphQLComposedTypeName(modelID, suffix string) string {
	singular := ApitoSingularResourceName(modelID)
	suf := strings.TrimPrefix(suffix, "_")
	sufParts := strings.Split(suf, "_")

	var modelSegs []string
	if strings.Contains(singular, "_") {
		modelSegs = strings.Split(singular, "_")
	} else {
		for _, s := range splitCamelPieces(singular) {
			modelSegs = append(modelSegs, strings.ToLower(s))
		}
	}
	var extra []string
	for _, chunk := range sufParts {
		if chunk == "" {
			continue
		}
		for _, x := range splitCamelPieces(chunk) {
			extra = append(extra, strings.ToLower(x))
		}
	}
	all := append(modelSegs, extra...)
	var parts []string
	for _, p := range all {
		if p == "" {
			continue
		}
		parts = append(parts, strings.ToUpper(p[:1])+strings.ToLower(p[1:]))
	}
	return strings.Join(parts, "_")
}

// ApitoSingularGraphQLTypeName returns PascalCase singular type name.
func ApitoSingularGraphQLTypeName(resource string) string {
	return pascalFromAnyModelID(ApitoSingularResourceName(resource))
}

// ApitoListGraphQLTypeName returns PascalCase list type name.
func ApitoListGraphQLTypeName(resource string) string {
	return listGraphQLTypeName(resource)
}

func apitoStoredSnakeModelID(resource string) string {
	singular := ApitoSingularResourceName(resource)
	if strings.Contains(singular, "_") {
		return singular
	}
	return strings.Join(splitCamelPieces(singular), "_")
}

// ApitoWhereInputType returns list where input type name.
func ApitoWhereInputType(resource string) string {
	return strings.ToUpper(listGraphQLTypeName(resource) + "_Input_Where_Payload")
}

// ApitoSortInputType returns list sort input type name.
func ApitoSortInputType(resource string) string {
	return strings.ToUpper(listGraphQLTypeName(resource) + "_Input_Sort_Payload")
}

// ApitoListCountWhereInputType returns list count where input type.
func ApitoListCountWhereInputType(resource string) string {
	return strings.ToUpper(ApitoGraphQLComposedTypeName(resource, "List_Count") + "_Input_Where_Payload")
}

// ApitoConnectionFilterConditionType returns connection filter enum name.
func ApitoConnectionFilterConditionType(resource string) string {
	return strings.ToUpper(apitoStoredSnakeModelID(resource) + "_Connection_Filter_Condition")
}

// CanonicalizeModelName normalizes admin input to snake_case singular.
// Parity with open-core utility/apito_naming.go: already-canonical ids
// (e.g. indication, practitioner) are accepted without run-on rejection.
func CanonicalizeModelName(raw string) (string, error) {
	t := strings.TrimSpace(raw)
	if t == "" {
		return "", errInvalidModelName
	}
	// Already-canonical snake_case: singularize last segment only.
	if canonicalIDRe.MatchString(t) {
		parts := strings.Split(t, "_")
		last := parts[len(parts)-1]
		if _, ok := singularKeepAsIs[last]; !ok && strings.HasSuffix(last, "s") && !strings.HasSuffix(last, "ss") {
			parts[len(parts)-1] = last[:len(last)-1]
		}
		out := strings.Join(parts, "_")
		if !canonicalIDRe.MatchString(out) {
			return "", errInvalidModelName
		}
		return out, nil
	}
	// simplified: split on _ and camel, singularize last segment
	normalized := strings.ReplaceAll(t, "-", "_")
	chunks := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '_' || r == ' ' || r == '\t'
	})
	var segments []string
	for _, chunk := range chunks {
		for _, p := range splitCamelPieces(chunk) {
			segments = append(segments, p)
		}
	}
	if len(segments) == 0 {
		return "", errInvalidModelName
	}
	last := segments[len(segments)-1]
	if _, ok := singularKeepAsIs[last]; !ok && strings.HasSuffix(last, "s") && !strings.HasSuffix(last, "ss") {
		segments[len(segments)-1] = last[:len(last)-1]
	}
	out := strings.Join(segments, "_")
	if !canonicalIDRe.MatchString(out) {
		return "", errInvalidModelName
	}
	return out, nil
}

var errInvalidModelName = &namingError{msg: "invalid model name"}

type namingError struct{ msg string }

func (e *namingError) Error() string { return e.msg }
