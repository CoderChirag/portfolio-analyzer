// Command jetgen generates the Jet table builders and row models from the
// migrated database schema, mapping columns to the domain's Go types.
//
//	go run ./cmd/jetgen -dsn 'folio:folio@tcp(localhost:3306)/folio' -out ./lib/db/repository/internal
package jetgen

import (
	"coderchirag/folio-tracker/lib/config"
	"coderchirag/folio-tracker/lib/domain"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-jet/jet/v2/generator/metadata"
	"github.com/go-jet/jet/v2/generator/mysql"
	"github.com/go-jet/jet/v2/generator/template"
	mysqldialect "github.com/go-jet/jet/v2/mysql"
)

// skipTables are tables the app never queries.
var skipTables = map[string]bool{
	"schema_migrations": true, // owned by golang-migrate
}

// columnTypes maps "table.column" to a domain type, for columns whose SQL type
// alone doesn't say which Go type they hold.
var columnTypes = map[string]any{
	"data_sources.code":      domain.DataSourceCode(""),
	"asset_types.code":       domain.AssetTypeCode(""),
	"transaction_types.code": domain.TransactionTypeCode(""),

	"instrument_listings.exchange": domain.Exchange(""),
	"stock_trade_details.exchange": domain.Exchange(""),

	"imports.kind":              &domain.ImportKind{},
	"imports.status":            domain.ImportStatus(""),
	"import_files.content_type": domain.ContentType(""),
	"jobs.job_type":             domain.JobType(""),
	"jobs.status":               domain.JobStatus(""),
}

// Lookup-table foreign keys, by column name; the same in every table.
var lookupIDTypes = map[string]any{
	"data_source_id":      domain.DataSourceID(0),
	"asset_type_id":       domain.AssetTypeID(0),
	"transaction_type_id": domain.TransactionTypeID(0),
}

// Lookup tables' own primary keys.
var lookupPKTypes = map[string]any{
	"data_sources":      domain.DataSourceID(0),
	"asset_types":       domain.AssetTypeID(0),
	"transaction_types": domain.TransactionTypeID(0),
}

func Main(out string) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})))

	cfg, err := config.Load()
	if err != nil {
		fatal("config parsing failed", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		cfg.DbFolioUser, cfg.DbFolioPassword, cfg.DbHost, cfg.DbPort, cfg.DbName)

	tmpl := template.Default(mysqldialect.Dialect).
		UseSchema(func(schema metadata.Schema) template.Schema {
			return template.DefaultSchema(schema).
				UsePath("internal/jetgen"). // not the DB name, so output doesn't depend on the environment
				UseModel(template.DefaultModel().
					UsePath("model").
					UseTable(func(table metadata.Table) template.TableModel {
						if skipTables[table.Name] {
							return template.TableModel{Skip: true}
						}
						return template.DefaultTableModel(table).UseField(
							func(col metadata.Column) template.TableModelField {
								field := template.DefaultTableModelField(col)
								if t, ok := goType(table.Name, col); ok {
									field = field.UseType(t)
								}
								return field
							})
					})).UseSQLBuilder(
				template.DefaultSQLBuilder().
					UseTable(func(table metadata.Table) template.TableSQLBuilder {
						if skipTables[table.Name] {
							return template.TableSQLBuilder{Skip: true}
						}
						return template.DefaultTableSQLBuilder(table)
					}))
		})

	if err := mysql.GenerateDSN(dsn, out, tmpl); err != nil {
		fatal("jetgen", err)
	}
}

// goType picks the domain type for a column; ok is false to keep Jet's default.
func goType(table string, col metadata.Column) (template.Type, bool) {
	var v any
	switch {
	case columnTypes[table+"."+col.Name] != nil:
		v = columnTypes[table+"."+col.Name]
	case col.Name == "id" && lookupPKTypes[table] != nil:
		v = lookupPKTypes[table]
	case lookupIDTypes[col.Name] != nil:
		v = lookupIDTypes[col.Name]
	case isBinary(col) && (col.Name == "id" || strings.HasSuffix(col.Name, "_id")):
		v = domain.ID{} // every BINARY(16) key is a UUIDv7
	case strings.ToLower(col.DataType.Name) == "json":
		return template.NewType(domain.JSON(nil)), true
	default:
		return template.Type{}, false
	}
	t := template.NewType(v)
	if col.IsNullable {
		t.Name = "*" + t.Name
	}
	return t, true
}

func isBinary(col metadata.Column) bool {
	return strings.ToLower(col.DataType.Name) == "binary"
}

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
