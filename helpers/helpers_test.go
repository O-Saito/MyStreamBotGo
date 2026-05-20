package helpers

import (
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		slice  []string
		target string
		want   bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not_found", []string{"a", "b", "c"}, "d", false},
		{"empty_slice", []string{}, "a", false},
		{"nil_slice", []string(nil), "a", false},
		{"single_match", []string{"a"}, "a", true},
		{"single_no_match", []string{"a"}, "b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.slice, tt.target); got != tt.want {
				t.Errorf("Contains(%v, %q) = %v, want %v", tt.slice, tt.target, got, tt.want)
			}
		})
	}
}

func TestFind(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		predicate := func(s *string) bool { return *s == "b" }
		got, ok := Find(slice, predicate)
		if !ok || *got != "b" {
			t.Errorf("Find() = (%q, %v), want (%q, true)", *got, ok, "b")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		slice := []string{"a", "b", "c"}
		predicate := func(s *string) bool { return *s == "d" }
		got, ok := Find(slice, predicate)
		if ok {
			t.Errorf("Find() = (%q, true), want (zero, false)", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		slice := []string{}
		predicate := func(s *string) bool { return *s == "a" }
		got, ok := Find(slice, predicate)
		if ok {
			t.Errorf("Find() = (%q, true), want (zero, false)", got)
		}
	})
}

func TestTernary(t *testing.T) {
	tests := []struct {
		name string
		cond bool
		a    int
		b    int
		want int
	}{
		{"true", true, 1, 2, 1},
		{"false", false, 1, 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ternary(tt.cond, tt.a, tt.b); got != tt.want {
				t.Errorf("Ternary(%v, %d, %d) = %d, want %d", tt.cond, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestTernaryString(t *testing.T) {
	t.Run("returns_a_when_true", func(t *testing.T) {
		got := Ternary(true, "yes", "no")
		if got != "yes" {
			t.Errorf("Ternary(true, \"yes\", \"no\") = %q, want \"yes\"", got)
		}
	})

	t.Run("returns_b_when_false", func(t *testing.T) {
		got := Ternary(false, "yes", "no")
		if got != "no" {
			t.Errorf("Ternary(false, \"yes\", \"no\") = %q, want \"no\"", got)
		}
	})
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("correct_length", func(t *testing.T) {
		got := GenerateRandomString(32)
		if len(got) != 32 {
			t.Errorf("GenerateRandomString(32) length = %d, want 32", len(got))
		}
	})

	t.Run("different_each_call", func(t *testing.T) {
		s1 := GenerateRandomString(32)
		s2 := GenerateRandomString(32)
		if s1 == s2 {
			t.Errorf("GenerateRandomString(32) returned same string twice")
		}
	})

	t.Run("valid_charset", func(t *testing.T) {
		charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
		s := GenerateRandomString(100)
		for _, c := range s {
			found := false
			for _, cc := range charset {
				if c == cc {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GenerateRandomString(100) contains invalid char %c", c)
			}
		}
	})
}

func TestGenerateCodeChallenge(t *testing.T) {
	t.Run("short", func(t *testing.T) {
		got := GenerateCodeChallenge("abc")
		if len(got) != 43 {
			t.Errorf("GenerateCodeChallenge(%q) len = %d, want 43", "abc", len(got))
		}
	})

	t.Run("long", func(t *testing.T) {
		got := GenerateCodeChallenge("this_is_a_test_verifier_string")
		if len(got) != 43 {
			t.Errorf("GenerateCodeChallenge(%q) len = %d, want 43", "this_is_a_test_verifier_string", len(got))
		}
	})

	t.Run("empty_produces_base64", func(t *testing.T) {
		got := GenerateCodeChallenge("")
		if len(got) == 0 {
			t.Errorf("GenerateCodeChallenge(%q) should produce output", "")
		}
	})

	t.Run("valid_base64url", func(t *testing.T) {
		got := GenerateCodeChallenge("test_verifier")
		for _, c := range got {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				t.Errorf("GenerateCodeChallenge() contains invalid Base64URL char: %c", c)
			}
		}
	})
}
