package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AdminAction string

const (
	AdminActionCreateHomework  AdminAction = "create_homework"
	AdminActionUpdateHomework  AdminAction = "update_homework"
	AdminActionDeleteHomework  AdminAction = "delete_homework"
	AdminActionPublishHomework AdminAction = "publish_homework"
	AdminActionSetDeadline     AdminAction = "set_deadline"
	AdminActionDeleteDeadline  AdminAction = "delete_deadline"

	AdminActionCreateTask AdminAction = "create_task"
	AdminActionUpdateTask AdminAction = "update_task"
	AdminActionDeleteTask AdminAction = "delete_task"
	AdminActionScoreTask  AdminAction = "score_task"

	AdminActionAssignRole        AdminAction = "assign_role"
	AdminActionRevokeRole        AdminAction = "revoke_role"
	AdminActionGrantPermission   AdminAction = "grant_permission"
	AdminActionRevokePermission  AdminAction = "revoke_permission"
	AdminActionPromoteSuperAdmin AdminAction = "promote_super_admin"
	AdminActionDemoteSuperAdmin  AdminAction = "demote_super_admin"
)

type AdminOutcome string

const (
	AdminOutcomeSuccess AdminOutcome = "success"
	AdminOutcomeError   AdminOutcome = "error"
)

type AdminMetrics struct {
	ActionTotal *prometheus.CounterVec
}

func newAdminMetrics(reg prometheus.Registerer) *AdminMetrics {
	factory := promauto.With(reg)

	return &AdminMetrics{
		ActionTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "admin",
				Name:      "action_total",
				Help:      "Total number of administrative actions.",
			},
			[]string{"action", "outcome"},
		),
	}
}

func (m *AdminMetrics) IncAction(action AdminAction, outcome AdminOutcome) {
	m.ActionTotal.WithLabelValues(string(action), string(outcome)).Inc()
}
