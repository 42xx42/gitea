// Copyright 2026 The 42w.shop Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

// AuditLog represents an audit log entry in the API
type AuditLog struct {
	ID        int64  `json:"id"`
	Action    string `json:"action"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	RepoID    int64  `json:"repo_id"`
	RepoName  string `json:"repo_name"`
	OrgID     int64  `json:"org_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Detail    string `json:"detail"`
	Created   string `json:"created"`
}
