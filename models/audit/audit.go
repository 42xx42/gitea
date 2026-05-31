// Copyright 2026 The 42w.shop Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package audit

import (
	"context"
	"encoding/json"
	"time"

	"gitea.dev/models/db"
	"gitea.dev/modules/timeutil"
)

// AuditAction describes the type of audited action
type AuditAction string

const (
	AuditClone        AuditAction = "clone"
	AuditPush         AuditAction = "push"
	AuditPull         AuditAction = "pull"
	AuditFork         AuditAction = "fork"
	AuditTransfer     AuditAction = "transfer"
	AuditRepoCreate   AuditAction = "repo_create"
	AuditRepoDelete   AuditAction = "repo_delete"
	AuditIssueCreate  AuditAction = "issue_create"
	AuditPRCreate     AuditAction = "pr_create"
	AuditReview       AuditAction = "review"
	AuditLogin        AuditAction = "login"
	AuditTokenCreate  AuditAction = "token_create"
	AuditPermChange   AuditAction = "perm_change"
	AuditArchive      AuditAction = "archive"
	AuditCollaborator AuditAction = "collaborator"
)

// AuditLog represents a single audit log entry
type AuditLog struct {
	ID        int64              `xorm:"pk autoincr"`
	Action    AuditAction        `xorm:"INDEX NOT NULL"`
	UserID    int64              `xorm:"INDEX"`
	UserName  string             `xorm:"INDEX"`
	RepoID    int64              `xorm:"INDEX"`
	RepoName  string             `xorm:"INDEX"`
	OrgID     int64              `xorm:"INDEX"`
	IP        string             `xorm:"INDEX"`
	UserAgent string             `xorm:"TEXT"`
	Detail    string             `xorm:"LONGTEXT"` // JSON blob for extra info
	Created   timeutil.TimeStamp `xorm:"INDEX created"`
}

func init() {
	db.RegisterModel(new(AuditLog))
}

// AuditDetail is a helper to marshal detail fields into JSON
type AuditDetail map[string]any

// ToJSON marshals the detail to JSON string
func (d AuditDetail) ToJSON() string {
	if d == nil {
		return ""
	}
	b, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(ctx context.Context, action AuditAction, userID int64, userName string, repoID int64, repoName string, orgID int64, ip string, userAgent string, detail string) error {
	log := &AuditLog{
		Action:    action,
		UserID:    userID,
		UserName:  userName,
		RepoID:    repoID,
		RepoName:  repoName,
		OrgID:     orgID,
		IP:        ip,
		UserAgent: userAgent,
		Detail:    detail,
	}
	return db.Insert(ctx, log)
}

// LogClone is a convenience function for logging clone operations
func LogClone(ctx context.Context, userID int64, userName string, repoID int64, repoName string, ip string, userAgent string, isSSH bool) error {
	detail := AuditDetail{"is_ssh": isSSH}.ToJSON()
	return CreateAuditLog(ctx, AuditClone, userID, userName, repoID, repoName, 0, ip, userAgent, detail)
}

// LogPush is a convenience function for logging push operations
func LogPush(ctx context.Context, userID int64, userName string, repoID int64, repoName string, ip string, userAgent string, isSSH bool) error {
	detail := AuditDetail{"is_ssh": isSSH}.ToJSON()
	return CreateAuditLog(ctx, AuditPush, userID, userName, repoID, repoName, 0, ip, userAgent, detail)
}

// LogArchive is a convenience function for logging archive downloads
func LogArchive(ctx context.Context, userID int64, userName string, repoID int64, repoName string, ip string, userAgent string, archiveType string, ref string) error {
	detail := AuditDetail{"archive_type": archiveType, "ref": ref}.ToJSON()
	return CreateAuditLog(ctx, AuditArchive, userID, userName, repoID, repoName, 0, ip, userAgent, detail)
}

// LogFork is a convenience function for logging fork operations
func LogFork(ctx context.Context, userID int64, userName string, repoID int64, repoName string, orgID int64, forkRepoName string) error {
	detail := AuditDetail{"fork_to": forkRepoName}.ToJSON()
	return CreateAuditLog(ctx, AuditFork, userID, userName, repoID, repoName, orgID, "", "", detail)
}

// LogTransfer is a convenience function for logging transfer operations
func LogTransfer(ctx context.Context, userID int64, userName string, repoID int64, repoName string, oldOwner string, newOwner string) error {
	detail := AuditDetail{"from": oldOwner, "to": newOwner}.ToJSON()
	return CreateAuditLog(ctx, AuditTransfer, userID, userName, repoID, repoName, 0, "", "", detail)
}

// LogRepoCreate is a convenience function for logging repo creation
func LogRepoCreate(ctx context.Context, userID int64, userName string, repoID int64, repoName string, orgID int64) error {
	return CreateAuditLog(ctx, AuditRepoCreate, userID, userName, repoID, repoName, orgID, "", "", "")
}

// LogRepoDelete is a convenience function for logging repo deletion
func LogRepoDelete(ctx context.Context, userID int64, userName string, repoID int64, repoName string, orgID int64) error {
	return CreateAuditLog(ctx, AuditRepoDelete, userID, userName, repoID, repoName, orgID, "", "", "")
}

// LogLogin is a convenience function for logging login events
func LogLogin(ctx context.Context, userID int64, userName string, ip string, userAgent string, loginType string) error {
	detail := AuditDetail{"login_type": loginType}.ToJSON()
	return CreateAuditLog(ctx, AuditLogin, userID, userName, 0, "", 0, ip, userAgent, detail)
}

// LogPermChange is a convenience function for logging permission changes
func LogPermChange(ctx context.Context, userID int64, userName string, repoID int64, repoName string, targetUser string, newPerm string) error {
	detail := AuditDetail{"target_user": targetUser, "new_perm": newPerm}.ToJSON()
	return CreateAuditLog(ctx, AuditPermChange, userID, userName, repoID, repoName, 0, "", "", detail)
}

// LogCollaborator is a convenience function for logging collaborator changes
func LogCollaborator(ctx context.Context, userID int64, userName string, repoID int64, repoName string, targetUser string, action string) error {
	detail := AuditDetail{"target_user": targetUser, "collab_action": action}.ToJSON()
	return CreateAuditLog(ctx, AuditCollaborator, userID, userName, repoID, repoName, 0, "", "", detail)
}

// FindAuditLogOptions defines search options for audit logs
type FindAuditLogOptions struct {
	db.ListOptions
	Action   AuditAction
	UserID   int64
	UserName string
	RepoID   int64
	RepoName string
	IP       string
	Since    time.Time
	Until    time.Time
}

// ApplyTo applies the options to the session
func (opts FindAuditLogOptions) ApplyTo(sess db.Engine) db.Engine {
	if opts.Action != "" {
		sess = sess.Where("action = ?", opts.Action)
	}
	if opts.UserID > 0 {
		sess = sess.Where("user_id = ?", opts.UserID)
	}
	if opts.UserName != "" {
		sess = sess.Where("user_name = ?", opts.UserName)
	}
	if opts.RepoID > 0 {
		sess = sess.Where("repo_id = ?", opts.RepoID)
	}
	if opts.RepoName != "" {
		sess = sess.Where("repo_name = ?", opts.RepoName)
	}
	if opts.IP != "" {
		sess = sess.Where("ip = ?", opts.IP)
	}
	if !opts.Since.IsZero() {
		sess = sess.Where("created >= ?", timeutil.TimeStamp(opts.Since.Unix()))
	}
	if !opts.Until.IsZero() {
		sess = sess.Where("created <= ?", timeutil.TimeStamp(opts.Until.Unix()))
	}
	return sess
}

// FindAuditLogs returns audit logs matching the options
func FindAuditLogs(ctx context.Context, opts FindAuditLogOptions) ([]*AuditLog, int64, error) {
	sess := opts.ApplyTo(db.GetEngine(ctx).Desc("created"))

	var count int64
	count, err := db.GetEngine(ctx).Count(new(AuditLog))
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*AuditLog, 0, opts.PageSize)
	sess.Limit(opts.PageSize, (opts.Page-1)*opts.PageSize)
	if err := sess.Find(&logs); err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}

// ActionCount holds a count of audit logs for a given action
type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// CountAuditLogsByAction returns the count of audit logs grouped by action
func CountAuditLogsByAction(ctx context.Context, opts FindAuditLogOptions) ([]*ActionCount, error) {
	sess := opts.ApplyTo(db.GetEngine(ctx))
	results := make([]*ActionCount, 0, 20)
	if err := sess.SQL("SELECT action, count(*) as count FROM audit_log GROUP BY action").Find(&results); err != nil {
		return nil, err
	}
	return results, nil
}

// DeleteOldAuditLogs deletes audit logs older than the given duration
func DeleteOldAuditLogs(ctx context.Context, olderThan time.Duration) error {
	if olderThan <= 0 {
		return nil
	}
	_, err := db.GetEngine(ctx).Where("created < ?", time.Now().Add(-olderThan).Unix()).Delete(&AuditLog{})
	return err
}
