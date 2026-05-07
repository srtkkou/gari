package gari

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build UPDATE SQL.
	deleteBuilder struct {
		table    *Table    // Pointer to table.
		query    string    // UPDATE SQL query.
		prepared *sql.Stmt // Prepared statement pointer.
		records  []*Record // Records to update.
		err      error     // Error.
	}
)

var (
	ErrDeletePrepare      = errors.New("gari.ErrDeletePrepare")
	ErrDeleteValuesEncode = errors.New("gari.ErrDeleteValuesEncode")
	ErrDeleteValuesDecode = errors.New("gari.ErrDeleteValuesDecode")
	ErrDeleteRecordsEmpty = errors.New("gari.ErrDeleteRecordsEmpty")
	ErrDeleteExec         = errors.New("gari.ErrDeleteExec")
	ErrDeleteRowsAffected = errors.New("gari.ErrDeleteRowsAffected")
	ErrDeleteRowCount     = errors.New("gari.ErrDeleteRowCount")
)

// Create new deleteBuilder instance.
func newDeleteBuilder(t *Table) *deleteBuilder {
	b := deleteBuilder{
		table:   t,
		records: make([]*Record, 0),
	}
	// Build DELETE SQL query.
	b.query = b.buildQuery()
	return &b
}

// Add values to dalete.
func (b *deleteBuilder) Values(ptrs ...any) *deleteBuilder {
	if b.err != nil {
		return b
	}
	for _, ptr := range ptrs {
		// Convert to msgpack.
		blob, err := msgpack.Marshal(ptr)
		if err != nil {
			b.err = errs.Wrap(ErrDeleteValuesEncode, errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			b.gari().errorLog(b.err.Error())
			return b
		}
		// Unmarshal msgpack to map[string]encoded.
		m, err := splitToMap(blob)
		if err != nil {
			b.err = errs.Wrap(ErrDeleteValuesDecode, errs.WithCause(err),
				errs.WithContext("ptr", ptr),
				errs.WithContext("msgpack", blob))
			b.gari().errorLog(b.err.Error())
			return b
		}
		// Build record.
		values := make([]*value, 1)
		values[0] = newValue(b.table.pkey)
		values[0].blob = m[b.table.pkey.fieldName]
		r := newRecord(values)
		b.records = append(b.records, r)
	}
	return b
}

// Execute UPDATE SQL statement.
func (b *deleteBuilder) Exec(ctx context.Context) error {
	defer func() {
		b.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if b.err != nil {
		return b.err
	}
	// Check if records are not empty.
	if len(b.records) == 0 {
		b.err = errs.Wrap(ErrDeleteRecordsEmpty,
			errs.WithContext("query", b.query))
		b.gari().errorLog(b.err.Error())
		return b.err
	}
	for _, r := range b.records {
		// TODO:BEFORE UPDATE
		// Execute prepared statement.
		args := r.args()
		startedAt := time.Now()
		result, err := b.prepared.ExecContext(ctx, args...)
		if err != nil {
			b.err = errs.Wrap(ErrDeleteExec, errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		b.gari().infoLog("Execute UPDATE SQL.",
			slog.String("query", b.query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			b.err = errs.Wrap(ErrDeleteRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		if count != 1 {
			b.err = errs.Wrap(ErrDeleteRowCount,
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		// TODO:AFTER UPDATE
	}
	return nil
}

// Close prepared statement.
func (b *deleteBuilder) closePreparedStmt() {
	if b.prepared != nil {
		b.prepared.Close()
	}
}

// Build SQL statement.
func (b *deleteBuilder) buildQuery() string {
	tokens := []string{"DELETE", "FROM"}
	// Add table name.
	quote := b.table.gari.columnQuote
	table := fmt.Sprintf("%s%s%s", quote, b.table.name, quote)
	tokens = append(tokens, table)
	// Add WHERE statement.
	tokens = append(tokens, "WHERE")
	pkey := fmt.Sprintf("%s%s%s", quote, b.table.pkey.name, quote)
	tokens = append(tokens, pkey, "=")
	ph := b.table.gari.placeholder(0) + ";"
	tokens = append(tokens, ph)
	return strings.Join(tokens, " ")
}

// Prepare UPDATE SQL statement.
func (b *deleteBuilder) prepareStmt() {
	startedAt := time.Now()
	var err error
	db := b.table.gari.db
	b.prepared, err = db.Prepare(b.query)
	if err != nil {
		b.err = errs.Wrap(ErrDeletePrepare, errs.WithCause(err),
			errs.WithContext("query", b.query),
			errs.WithContext("duration", time.Since(startedAt)))
		b.gari().errorLog(b.err.Error())
	}
	b.gari().infoLog("deleteBuilder.prepareStmt()",
		slog.String("query", b.query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (b *deleteBuilder) gari() *Gari {
	return b.table.gari
}
