// Copyright 2026 The 42w.shop Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"
	"time"

	audit_model "gitea.dev/models/audit"
	api "gitea.dev/modules/structs"
	"gitea.dev/routers/api/v1/utils"
	"gitea.dev/services/context"
)

// ListAuditLogs returns audit logs matching the given filters
func ListAuditLogs(ctx *context.APIContext) {
	opts := audit_model.FindAuditLogOptions{
		ListOptions: utils.GetListOptions(ctx),
	}

	if action := ctx.FormString("action"); action != "" {
		opts.Action = audit_model.AuditAction(action)
	}
	if userID := ctx.FormInt64("user_id"); userID > 0 {
		opts.UserID = userID
	}
	if userName := ctx.FormString("user_name"); userName != "" {
		opts.UserName = userName
	}
	if repoID := ctx.FormInt64("repo_id"); repoID > 0 {
		opts.RepoID = repoID
	}
	if repoName := ctx.FormString("repo_name"); repoName != "" {
		opts.RepoName = repoName
	}
	if ip := ctx.FormString("ip"); ip != "" {
		opts.IP = ip
	}
	if since := ctx.FormString("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			opts.Since = t
		}
	}
	if until := ctx.FormString("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			opts.Until = t
		}
	}

	logs, count, err := audit_model.FindAuditLogs(ctx, opts)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	apiLogs := make([]*api.AuditLog, len(logs))
	for i, log := range logs {
		apiLogs[i] = toAPIAuditLog(log)
	}

	ctx.SetTotalCountHeader(count)
	ctx.JSON(http.StatusOK, apiLogs)
}

// GetAuditStats returns counts of audit logs grouped by action
func GetAuditStats(ctx *context.APIContext) {
	opts := audit_model.FindAuditLogOptions{}

	if since := ctx.FormString("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			opts.Since = t
		}
	}
	if until := ctx.FormString("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			opts.Until = t
		}
	}
	if repoName := ctx.FormString("repo_name"); repoName != "" {
		opts.RepoName = repoName
	}

	results, err := audit_model.CountAuditLogsByAction(ctx, opts)
	if err != nil {
		ctx.APIErrorInternal(err)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func toAPIAuditLog(log *audit_model.AuditLog) *api.AuditLog {
	return &api.AuditLog{
		ID:        log.ID,
		Action:    string(log.Action),
		UserID:    log.UserID,
		UserName:  log.UserName,
		RepoID:    log.RepoID,
		RepoName:  log.RepoName,
		OrgID:     log.OrgID,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Detail:    log.Detail,
		Created:   log.Created.AsTimePtr().Format(time.RFC3339),
	}
}
