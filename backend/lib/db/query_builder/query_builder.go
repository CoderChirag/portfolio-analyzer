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
)

func (o Op) IsValid() bool {
	switch o {
	case EqualOp, NotEqualOp, LessThanOp, GreaterThanOp, LessThanEqualOp, GreaterThanEqualOp, LikeOp, InOp, AndOp, OrOp:
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
	query     string
	args      []any

	selection []string
	selectAll bool
	where     *cond

	err []error
}

func New[Model any](db db.DBTX, tablename string, queryType QueryType) (*QueryBuilder[Model], error) {
	if !queryType.IsValid() {
		return nil, fmt.Errorf("query_builder.New: invalid queryType: %s.", queryType)
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
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query_builder.New: model %s is not a struct.", rv.Type())
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
			col = sf.Name
		}

		if sf.Type.Size() == 0 {
			return fmt.Errorf("%s.%s: zero-size fields can't be columns.", t.Name(), sf.Name)
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
	if _, _, err := qb.Build(); err != nil {
		return nil, fmt.Errorf("QueryBuilder.ExecContext: %w", err)
	}

	switch qb.queryType {
	case SelectQuery:
		slog.Debug("QueryBuilder.ExecContext: executing query", "query", qb.query)
		return nil, qb.db.GetContext(ctx, dest, qb.query, qb.args...)
	default:
		return nil, fmt.Errorf("QueryBuilder.ExecContext: invalid queryType: %s.", qb.queryType)
	}

}

func (qb *QueryBuilder[Model]) Build() (string, []any, error) {
	if len(qb.err) != 0 {
		return "", nil, fmt.Errorf("QueryBuilder.Build: %w", errors.Join(qb.err...))
	}

	switch qb.queryType {
	case SelectQuery:
		q, a := qb.buildSelectQuery()
		return q, a, nil
	default:
		return "", nil, fmt.Errorf("QueryBuilder.Build: invalid queryType: %s.", qb.queryType)
	}
}

func (qb *QueryBuilder[Model]) buildSelectQuery() (string, []any) {
	qb.query = "SELECT "
	for i, f := range qb.selection {
		if i != len(qb.selection)-1 {
			qb.query += fmt.Sprintf("%s,", f)
		} else {
			qb.query += fmt.Sprintf("%s ", f)
		}
	}

	qb.query += fmt.Sprintf("FROM %s WHERE ", qb.tablename)
	condString, args := qb.buildConditions(qb.where)
	qb.query += condString
	qb.args = args

	return qb.query, qb.args
}

func (qb *QueryBuilder[Model]) buildConditions(cond *cond) (string, []any) {
	if cond.cond1 == nil && cond.cond2 == nil {
		return fmt.Sprintf("%s %s ?", cond.col.name, string(cond.op)), []any{cond.value}
	}

	cond1, args1 := qb.buildConditions(cond.cond1)
	q := fmt.Sprintf("(%s) %s ", cond1, string(cond.op))
	cond2, args2 := qb.buildConditions(cond.cond2)
	q += fmt.Sprintf("(%s)", cond2)

	args := make([]any, 0, len(args1)+len(args2))
	args = append(args, args1...)
	args = append(args, args2...)

	return q, args
}

func (qb *QueryBuilder[Model]) Col(field any) *column {
	if len(qb.err) != 0 {
		return &column{name: ""}
	}

	// Map lookup panics if field's dynamic type isn't hashable,
	// e.g. qb.Model.Tags ([]string) passed without &
	defer func() {
		if r := recover(); r != nil {
			qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Col: invalid field. Pass field pointer like \"&qb.Model.Name.\""))
		}
	}()

	col, ok := qb.cols[field]
	if !ok {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Col: invalid field '%s' (type: %T), not a field pointer of 'models.%s' (did you forget '&'?)", field, field, qb.modelname))
	}

	return &column{name: col}
}

func (qb *QueryBuilder[Model]) Cond(field *column, op Op, value any) *cond {
	if !op.IsValid() {
		qb.err = append(qb.err, fmt.Errorf("QueryBuilder.Cond: invalid operation %s", string(op)))
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
