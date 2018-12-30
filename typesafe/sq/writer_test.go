package sq

import "testing"

func TestAppendAndReplace(t *testing.T) {
	tests := []struct {
		input string
		exp1  string // mysql
		exp2  string // postgres
	}{
		{
			`sample`,
			`sample`,
			`sample`,
		},
		{
			`foo = ? AND bar = ?`,
			`foo = ? AND bar = ?`,
			`foo = $1 AND bar = $2`,
		},
		{
			`"foo" = ? AND "bar" = ?`,
			"`foo` = ? AND `bar` = ?",
			`"foo" = $1 AND "bar" = $2`,
		},
		{
			`INSERT INTO "user"("id", "name") VALUES (?,?)`,
			"INSERT INTO `user`(`id`, `name`) VALUES (?,?)",
			`INSERT INTO "user"("id", "name") VALUES ($1,$2)`,
		},
		{
			`$.deleted_at IS NULL`,
			`schema.deleted_at IS NULL`,
			`schema.deleted_at IS NULL`,
		},
		{
			`INSERT INTO $."user"("id", "name") VALUES (?,?)`,
			"INSERT INTO schema.`user`(`id`, `name`) VALUES (?,?)",
			`INSERT INTO schema."user"("id", "name") VALUES ($1,$2)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			{
				var b []byte
				output := appendAndReplace(b, nil, '`', '?', tt.input, "schema")
				if string(output) != tt.exp1 {
					t.Errorf("\nExpect: %s\nOutput: %s\n", tt.exp1, output)
				}
			}
			{
				var b []byte
				var c int64
				output := appendAndReplace(b, &c, '"', '$', tt.input, "schema")
				if string(output) != tt.exp2 {
					t.Errorf("\nExpect: %s\nOutput: %s\n", tt.exp2, output)
				}
			}
		})
	}
}
