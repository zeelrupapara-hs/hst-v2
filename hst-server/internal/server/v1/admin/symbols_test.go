package admin

import "testing"

func TestValidateSessionsOvernightWindows(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 2, Open: 710, Close: 1440},
		{Type: 0, Day: 2, Open: 10, Close: 470},
		{Type: 1, Day: 2, Open: 5, Close: 1440},
	}
	if err := validateSessions(sessions); err != nil {
		t.Fatalf("expected overnight windows to pass validation, got: %v", err)
	}
}

func TestValidateSessionsOverlapRejected(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 1, Open: 100, Close: 500},
		{Type: 0, Day: 1, Open: 400, Close: 800},
	}
	if err := validateSessions(sessions); err == nil {
		t.Fatal("expected overlapping sessions to fail validation")
	}
}

func TestValidateSessionsOpenMustBeLessThanClose(t *testing.T) {
	sessions := []CrtSymbolSession{
		{Type: 0, Day: 0, Open: 600, Close: 600},
	}
	if err := validateSessions(sessions); err == nil {
		t.Fatal("expected open >= close to fail validation")
	}
}
