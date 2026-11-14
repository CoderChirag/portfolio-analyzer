package repository

import "coderchirag/folio-tracker/lib/db/repository/internal/jetgen/table"

// UseSchema sets the schema prefix for every generated table; "" means the connection's database.
func UseSchema(name string) {
	table.UseSchema(name)
}
