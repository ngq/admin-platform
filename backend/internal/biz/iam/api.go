package iam

import "time"

// APIResource represents an API route that can be bound to menu/function permissions.
type APIResource struct {
	ID             int64
	Group          string
	Path           string
	Method         string
	Description    string
	MenuID         int64
	ActionType     string
	PermissionMode string
	Ignored        int8
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
