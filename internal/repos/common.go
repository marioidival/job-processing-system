package repos

import (
	"database/sql"
	"github.com/google/uuid"
	"strings"
	"time"
)

func ToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func FromString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullableString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func FromNullableString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{String: "", Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func FromUUID(id uuid.UUID) uuid.NullUUID {
	if id == uuid.Nil {
		return uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func ToNullableUUID(id uuid.NullUUID) *uuid.UUID {
	if id.Valid {
		return &id.UUID
	}
	return nil
}

func FromNullableUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil || *id == uuid.Nil {
		return uuid.NullUUID{UUID: uuid.Nil, Valid: false}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func FromTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}

func ToNullableTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func FromNullableTime(t *time.Time) sql.NullTime {
	if t == nil || t.IsZero() {
		return sql.NullTime{Time: time.Time{}, Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func FromInt16(i int16) sql.NullInt16 {
	return sql.NullInt16{Int16: i, Valid: true}
}

func ToNullableInt16(i sql.NullInt16) *int16 {
	if i.Valid {
		return &i.Int16
	}
	return nil
}

func FromInt32(i int32) sql.NullInt32 {
	return sql.NullInt32{Int32: i, Valid: i != 0}
}

func ToNullableInt32(i sql.NullInt32) *int32 {
	if i.Valid {
		return &i.Int32
	}
	return nil
}

func FromNullableInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Int32: 0, Valid: false}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func FromBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func ToBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

func ConvertNullableUUID(u *uuid.UUID) string {
	if u != nil {
		return u.String()
	}
	return ""
}

func ToInt32(i sql.NullInt32) int32 {
	if i.Valid {
		return i.Int32
	}
	return 0
}

func PrepareQueryForFullTextSearch(query string) string {
	terms := strings.Split(query, " ")
	queryTerms := ""
	for i, t := range terms {
		if t == "" {
			continue
		}
		queryTerms = queryTerms + t + ":*"
		if i != len(terms)-1 {
			queryTerms = queryTerms + " & "
		}
	}
	return queryTerms
}
