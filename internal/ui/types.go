package ui

import "strings"

// Types retorna a lista de tipos suportados (ordem mostrada no wizard).
func Types() []CommitType {
	return commitTypes
}

// FindType procura por key case-insensitive.
func FindType(key string) (CommitType, bool) {
	for _, ct := range commitTypes {
		if equalFold(ct.Key, key) {
			return ct, true
		}
	}
	return CommitType{}, false
}

// equalFold evita importar strings só para EqualFold.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return toLower(a) == toLower(b)
	}
	return toLower(a) == toLower(b)
}

func toLower(s string) string {
	return strings.ToLower(s)
}
