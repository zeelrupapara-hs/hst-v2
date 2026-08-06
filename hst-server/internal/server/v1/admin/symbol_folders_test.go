package admin

import "testing"

func TestValidateFolderSegment(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"letters", "Forex", false},
		{"digits", "Test123", false},
		{"allowed punct", "Symbols.c", false},
		{"empty", "  ", true},
		{"backslash", "bad\\name", true},
		{"space", "bad name", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFolderSegment(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateFolderSegment(%q) err=%v wantErr=%v", tc.in, err, tc.wantErr)
			}
		})
	}
}

func TestJoinSymbolFolder(t *testing.T) {
	got, err := joinSymbolFolder("Forex", "Majors")
	if err != nil {
		t.Fatal(err)
	}
	if got != `Forex\Majors` {
		t.Fatalf("got %q", got)
	}

	got, err = joinSymbolFolder("", "Synthetic")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Synthetic" {
		t.Fatalf("got %q", got)
	}
}
