package logadapter

import (
	"regexp"
	"strings"

	"go.uber.org/zap/zapcore"
)

var (
	// flc_ prefixed keys/secrets
	apiKeyRegex = regexp.MustCompile(`flc_[a-zA-Z0-9]+`)
)

type scrubberCore struct {
	zapcore.Core
}

// NewScrubberCore wraps an existing zapcore.Core and masks sensitive strings in messages and fields.
func NewScrubberCore(next zapcore.Core) zapcore.Core {
	return &scrubberCore{Core: next}
}

func (s *scrubberCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// 1. Mask sensitive data in the message (immutable entry)
	ent.Message = maskSensitive(ent.Message)

	// 2. Mask sensitive data in fields (avoid mutating pooled slice)
	var scrubbedFields []zapcore.Field
	for i, f := range fields {
		if f.Type == zapcore.StringType && strings.Contains(f.String, "flc_") {
			if scrubbedFields == nil {
				// Allocate only if we actually find a secret, preserving zero-alloc for clean logs
				scrubbedFields = make([]zapcore.Field, len(fields))
				copy(scrubbedFields, fields)
			}
			scrubbedFields[i].String = maskSensitive(f.String)
		}
	}

	if scrubbedFields != nil {
		return s.Core.Write(ent, scrubbedFields)
	}
	return s.Core.Write(ent, fields)
}

func (s *scrubberCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(ent.Level) {
		return ce.AddCore(ent, s)
	}
	return ce
}

func (s *scrubberCore) With(fields []zapcore.Field) zapcore.Core {
	var scrubbedFields []zapcore.Field
	for i, f := range fields {
		if f.Type == zapcore.StringType && strings.Contains(f.String, "flc_") {
			if scrubbedFields == nil {
				scrubbedFields = make([]zapcore.Field, len(fields))
				copy(scrubbedFields, fields)
			}
			scrubbedFields[i].String = maskSensitive(f.String)
		}
	}

	if scrubbedFields != nil {
		return &scrubberCore{Core: s.Core.With(scrubbedFields)}
	}
	return &scrubberCore{Core: s.Core.With(fields)}
}

func maskSensitive(input string) string {
	if !strings.Contains(input, "flc_") {
		return input
	}
	return apiKeyRegex.ReplaceAllStringFunc(input, func(match string) string {
		if len(match) <= 8 {
			return "flc_****"
		}
		return match[:8] + "****"
	})
}
