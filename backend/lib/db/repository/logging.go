package repository

import (
	"context"
	"log/slog"

	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

func LogQuery(ctx context.Context, op string, stmt mysqldialect.Statement) {
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		slog.DebugContext(ctx, "sql query", "op", op, "sql", stmt.DebugSql())
	}
}
