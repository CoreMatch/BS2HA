package bsdump

import "testing"

func TestParseCreateAndInsert(t *testing.T) {
	sql := `
CREATE TABLE ` + "`users`" + ` (
  ` + "`uid`" + ` int NOT NULL,
  ` + "`username`" + ` varchar(255) DEFAULT NULL,
  ` + "`password`" + ` varchar(255) NOT NULL
);
INSERT INTO ` + "`users`" + ` VALUES (1,'Alice','$2y$10$hash');
`

	dump, err := Parse(sql)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	rows := dump.Tables["users"]
	if len(rows) != 1 {
		t.Fatalf("expected 1 user row, got %d", len(rows))
	}

	if got, _ := rows[0].Get("username"); got != "Alice" {
		t.Fatalf("username = %q, want %q", got, "Alice")
	}
	if got, _ := rows[0].Get("password"); got != "$2y$10$hash" {
		t.Fatalf("password = %q", got)
	}
}
