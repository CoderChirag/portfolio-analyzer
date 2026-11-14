package query_builder

import (
	"coderchirag/folio-tracker/lib/db"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type QueryType string

const (
	SelectQuery QueryType = "select"
	UpdateQuery QueryType = "update"
	InsertQuery QueryType = "insert"
)

func (q QueryType) IsValid() bool {
	switch q {
	case SelectQuery, UpdateQuery, InsertQuery:
		return true
	}
	return false
}

type Op string

const (
	EqualOp            Op = "="
	NotEqualOp         Op = "!="
	LessThanOp         Op = "<"
	GreaterThanOp      Op = ">"
	LessThanEqualOp    Op = "<="
	GreaterThanEqualOp Op = ">="
	LikeOp             Op = "LIKE"
	InOp               Op = "IN"
	AndOp              Op = "AND"
	OrOp               Op = "OR"
	IsNullOp           Op = "IS NULL"
	IsNotNullOp        Op = "IS NOT NULL"
)

func (o Op) IsValid() bool {
	switch o {
	case EqualOp, NotEqualOp, LessThanOp, GreaterThanOp, LessThanEqualOp, GreaterThanEqualOp, LikeOp, InOp, AndOp, OrOp, IsNullOp, IsNotNullOp:
		return true
	}
	return false
}

func (o Op) IsComparisonOp() bool {
	switch o {
	case EqualOp, NotEqualOp, LessThanOp, GreaterThanOp, LessThanEqualOp, GreaterThanEqualOp, LikeOp, InOp:
		return true
	}
	return false
}

func (o Op) IsLogicalOp() bool {
	switch o {
	case AndOp, OrOp:
		return true
	}
	return false
}

func (o Op) IsNoArgOp() bool {
	switch o {
	case IsNullOp, IsNotNullOp:
		return true
	}
	return false
}

type column struct {
	name string
}

type cond struct {
	col   *column
	cond1 *cond
	cond2 *cond
	op    Op
	value any
}

type QueryBuilder[Model any] struct {
	Model Model
	db    db.DBTX

	tablename string
	modelname string
	columns   []string
	cols      map[any]string

	queryType QueryType

	selection []string
	selectAll bool
	where     *cond

	err []error
}

func New[Model any](db db.DBTX, tablename string, queryType QueryType) (*QueryBuilder[Model], error) {
	if !queryType.IsValid() {
		return nil, fmt.Errorf("query_builder.New: invalid queryType: %s", queryType)
	}
	qb := &QueryBuilder[Model]{
		db:        db,
		tablename: tablename,
		queryType: queryType,
		columns:   make([]string, 0),
		cols:      map[any]string{},
	}

	rv := reflect.ValueOf(&qb.Model).Elem()
	if rv.Kind() == reflect.Pointer {
		rv.Set(reflect.New(rv.Type().Elem()))
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query_builder.New: model %s is not a struct", rv.Type())
	}

	qb.modelname = rv.Type().Name()
	if tablename == "" {
		qb.tablename = strings.ToLower(rv.Type().Name()) + "s"
	}
	if err := qb.indexColumns(rv); err != nil {
		return nil, fmt.Errorf("query_builder.New: %w", err)
	}

	return qb, nil
}

func (qb *QueryBuilder[Model]) indexColumns(rv reflect.Value) error {
	t := rv.Type()
	for i := range t.NumField() {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		fv := rv.Field(i)
		if sf.Anonymous && fv.Kind() == reflect.Struct {
			if err := qb.indexColumns(fv); err != nil {
				return err
			}
			continue
		}

		col := sf.Tag.Get("db")
		if col == "-" {
			continue
		}
		if col == "" {
			col = strings.ToLower(sf.Name)
		}

		if sf.Type.Size() == 0 {
			return fmt.Errorf("%s.%s: zero-size fields can't be columns", t.Name(), sf.Name)
		}

		qb.cols[fv.Addr().Interface()] = col
		qb.columns = append(qb.columns, col)
	}
	return nil
}

func (qb *QueryBuilder[Model]) Select(fields ...*column) *QueryBuilder[Model] {
	if len(qb.err) != 0 {
		return qb
	}

	if qb.selectAll {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Select: cannot chain Select() with SelectAll()"))
		return qb
	}

	for _, field := range fields {
		qb.selection = append(qb.selection, field.name)
	}

	return qb
}

func (qb *QueryBuilder[Model]) SelectAll() *QueryBuilder[Model] {
	if len(qb.err) != 0 || qb.selectAll {
		return qb
	}

	if len(qb.selection) != 0 {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.SelectAll: cannot chain SelectAll() with Select()"))
		return qb
	}

	qb.selection = append(qb.selection, qb.columns...)
	qb.selectAll = true
	return qb
}

func (qb *QueryBuilder[Model]) Where(cond *cond) *QueryBuilder[Model] {
	if len(qb.err) != 0 {
		return qb
	}

	qb.where = cond

	return qb
}

func (qb *QueryBuilder[Model]) ExecContext(ctx context.Context, dest any) (sql.Result, error) {
	query, args, err := qb.Build()
	if err != nil {
		return nil, fmt.Errorf("QueryBuilder.ExecContext: %w", err)
	}

	switch qb.queryType {
	case SelectQuery:
		slog.Debug("QueryBuilder.ExecContext: executing query", "query", query)
		return nil, qb.db.GetContext(ctx, dest, query, args...)
	default:
		return nil, fmt.Errorf("QueryBuilder.ExecContext: invalid queryType: %s", qb.queryType)
	}

}

func (qb *QueryBuilder[Model]) Build() (string, []any, error) {
	if len(qb.err) != 0 {
		return "", nil, fmt.Errorf("QueryBuilder.Build: %w", errors.Join(qb.err...))
	}

	switch qb.queryType {
	case SelectQuery:
		if len(qb.selection) == 0 {
			return "", nil, fmt.Errorf("QueryBuilder.Build: No SELECT field added")
		}
		q, a, err := qb.buildSelectQuery()
		if err != nil {
			return "", nil, fmt.Errorf("QueryBuilder.Build: %w", err)
		}
		return q, a, nil
	default:
		return "", nil, fmt.Errorf("QueryBuilder.Build: invalid queryType: %s", qb.queryType)
	}
}

func (qb *QueryBuilder[Model]) buildSelectQuery() (string, []any, error) {
	query := "SELECT "
	for i, f := range qb.selection {
		if i != len(qb.selection)-1 {
			query += fmt.Sprintf("%s,", f)
		} else {
			query += fmt.Sprintf("%s ", f)
		}
	}

	query += fmt.Sprintf("FROM %s", qb.tablename)

	q, args, err := qb.buildWhereQuery()
	if err != nil {
		return "", nil, fmt.Errorf("QueryBuilder.buildSelectQuery: %w", err)
	}
	query += q

	return query, args, nil
}

func (qb *QueryBuilder[Model]) buildWhereQuery() (string, []any, error) {
	var args []any
	query := ""
	if qb.where != nil {
		query += " WHERE "
		condString, a, err := qb.buildConditions(qb.where)
		if err != nil {
			return "", nil, fmt.Errorf("QueryBuilder.buildWhereQuery: %w", err)
		}
		query += condString
		args = a
	} else {
		args = make([]any, 0)
	}

	var err error
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, fmt.Errorf("QueryBuilder.buildWhereQuery: %w", err)
	}

	return query, args, nil
}

func (qb *QueryBuilder[Model]) buildConditions(cond *cond) (string, []any, error) {
	if cond.cond1 == nil && cond.cond2 == nil {
		if cond.op.IsLogicalOp() {
			return "", nil, fmt.Errorf("QueryBuilder.buildConditions: terminal cond node cannot contain logical operator")
		}
		if cond.op.IsNoArgOp() {
			return fmt.Sprintf("%s %s", cond.col.name, string(cond.op)), []any{}, nil
		}
		return fmt.Sprintf("%s %s ?", cond.col.name, string(cond.op)), []any{cond.value}, nil
	}

	if !cond.op.IsLogicalOp() {
		return "", nil, fmt.Errorf("QueryBuilder.buildConditions: non-terminal cond node needs to have logical operator")
	}

	cond1, args1, err := qb.buildConditions(cond.cond1)
	if err != nil {
		return "", nil, err
	}
	q := fmt.Sprintf("(%s) %s ", cond1, string(cond.op))
	cond2, args2, err := qb.buildConditions(cond.cond2)
	if err != nil {
		return "", nil, err
	}
	q += fmt.Sprintf("(%s)", cond2)

	args := make([]any, 0, len(args1)+len(args2))
	args = append(args, args1...)
	args = append(args, args2...)

	return q, args, nil
}

func (qb *QueryBuilder[Model]) Col(field any) (col *column) {
	if len(qb.err) != 0 {
		return &column{name: ""}
	}

	// Map lookup panics if field's dynamic type isn't hashable,
	// e.g. qb.Model.Tags ([]string) passed without &
	defer func() {
		if r := recover(); r != nil {
			qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Col: invalid field. Pass field pointer like \"&qb.Model.Name.\""))
		}
		col = &column{name: ""}
	}()

	c, ok := qb.cols[field]
	if !ok {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Col: invalid field '%s' (type: %T), not a field pointer of 'models.%s' (did you forget '&'?)", field, field, qb.modelname))
		return &column{name: ""}
	}

	return &column{name: c}
}

func (qb *QueryBuilder[Model]) Cond(field *column, op Op, value any) *cond {
	if !op.IsComparisonOp() {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Cond: invalid operation %s", string(op)))
	}

	if value == nil && !op.IsNoArgOp() {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Cond: value cannot be nil for a non noArgOp"))
	}

	return &cond{
		col:   field,
		op:    op,
		value: value,
	}
}

func (qb *QueryBuilder[Model]) And(cond1 *cond, cond2 *cond) *cond {
	return &cond{
		op:    AndOp,
		cond1: cond1,
		cond2: cond2,
	}
}

func (qb *QueryBuilder[Model]) Or(cond1 *cond, cond2 *cond) *cond {
	return &cond{
		op:    OrOp,
		cond1: cond1,
		cond2: cond2,
	}
}
