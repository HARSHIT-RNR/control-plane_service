package helpers

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgUUID converts uuid.UUID to pgtype.UUID
func UUIDToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

// StringToPgUUID converts string UUID to pgtype.UUID
func StringToPgUUID(id string) (pgtype.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{
		Bytes: uid,
		Valid: true,
	}, nil
}

// PgUUIDToUUID converts pgtype.UUID to uuid.UUID
func PgUUIDToUUID(pgUUID pgtype.UUID) (uuid.UUID, error) {
	if !pgUUID.Valid {
		return uuid.Nil, nil
	}
	return uuid.FromBytes(pgUUID.Bytes[:])
}

// PgUUIDToString converts pgtype.UUID to string
func PgUUIDToString(pgUUID pgtype.UUID) string {
	if !pgUUID.Valid {
		return ""
	}
	uid, _ := uuid.FromBytes(pgUUID.Bytes[:])
	return uid.String()
}

// StringPtrToPgText converts *string to pgtype.Text
func StringPtrToPgText(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

// StringToPgText converts string to pgtype.Text
func StringToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

// PgTextToString converts pgtype.Text to string
func PgTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// UUIDPtrToPgUUID converts *uuid.UUID to pgtype.UUID
func UUIDPtrToPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{
		Bytes: *id,
		Valid: true,
	}
}
