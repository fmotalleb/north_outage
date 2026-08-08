package database

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/fmotalleb/north_outage/internal/otel"
	"gorm.io/gorm"
)

const dbSpanKey = "otel:span"

// tracingPlugin wraps every GORM operation (create/query/update/delete/row/raw)
// in an OTel span so database work shows up in traces. If the query carries a
// context (via .WithContext), the span links to it; otherwise it is a root span.
type tracingPlugin struct{}

func (tracingPlugin) Name() string { return "otel:tracing" }

func (tracingPlugin) Initialize(db *gorm.DB) error {
	cb := db.Callback()
	for _, o := range []struct {
		name, before, after string
		register            func() error
	}{
		{"create", "gorm:create", "gorm:after_create", func() error {
			return registerOp(cb.Create(), "create", "gorm:create", "gorm:after_create", beforeCallback("create"), afterCallback)
		}},
		{"query", "gorm:query", "gorm:after_query", func() error {
			return registerOp(cb.Query(), "query", "gorm:query", "gorm:after_query", beforeCallback("query"), afterCallback)
		}},
		{"update", "gorm:update", "gorm:after_update", func() error {
			return registerOp(cb.Update(), "update", "gorm:update", "gorm:after_update", beforeCallback("update"), afterCallback)
		}},
		{"delete", "gorm:delete", "gorm:after_delete", func() error {
			return registerOp(cb.Delete(), "delete", "gorm:delete", "gorm:after_delete", beforeCallback("delete"), afterCallback)
		}},
		{"row", "gorm:row", "gorm:after_row", func() error {
			return registerOp(cb.Row(), "row", "gorm:row", "gorm:after_row", beforeCallback("row"), afterCallback)
		}},
		{"raw", "gorm:raw", "gorm:after_raw", func() error {
			return registerOp(cb.Raw(), "raw", "gorm:raw", "gorm:after_raw", beforeCallback("raw"), afterCallback)
		}},
	} {
		o := o
		if err := o.register(); err != nil {
			return err
		}
	}
	return nil
}

// registerOp attaches a before/after span pair to a GORM callback processor.
// The processor and callback types are unexported in gorm, so they are
// inferred through the generic constraint instead of being named.
func registerOp[P interface {
	Before(string) R
	After(string) R
	Register(string, func(*gorm.DB)) error
}, R interface {
	Register(string, func(*gorm.DB)) error
}](p P, name, before, after string, beforeFn, afterFn func(*gorm.DB)) error {
	if err := p.Before(before).Register("otel:before_"+name, beforeFn); err != nil {
		return err
	}
	return p.After(after).Register("otel:after_"+name, afterFn)
}

func beforeCallback(op string) func(*gorm.DB) {
	return func(tx *gorm.DB) {
		ctx := tx.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}
		spanName := "gorm." + op
		if table := tx.Statement.Table; table != "" {
			spanName += " " + table
		}
		_, span := otel.Tracer("north_outage.database").Start(ctx, spanName)
		tx.Statement.Settings.Store(dbSpanKey, span)
	}
}

func afterCallback(tx *gorm.DB) {
	v, ok := tx.Statement.Settings.Load(dbSpanKey)
	if !ok {
		return
	}
	span, ok := v.(trace.Span)
	if !ok {
		return
	}
	defer span.End()
	if table := tx.Statement.Table; table != "" {
		span.SetAttributes(attribute.String("gorm.table", table))
	}
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		span.RecordError(tx.Error)
		span.SetStatus(codes.Error, tx.Error.Error())
	}
}
