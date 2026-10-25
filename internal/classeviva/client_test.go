package classeviva

import "testing"

func TestStudentIDFromUID(t *testing.T) {
	tests := []struct {
		name string
		uid  string
		want string
	}{
		{
			name: "UID with final letter",
			uid:  "S12345678A",
			want: "12345678",
		},
		{
			name: "numeric suffix remains compatible",
			uid:  "S123456789",
			want: "123456789",
		},
		{
			name: "lowercase prefix with final letter",
			uid:  "s12345678a",
			want: "12345678",
		},
		{
			name: "other UID formats retain legacy conversion",
			uid:  "P12AB34",
			want: "12AB34",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := studentIDFromUID(test.uid); got != test.want {
				t.Fatalf("studentIDFromUID(%q) = %q, want %q", test.uid, got, test.want)
			}
		})
	}
}
