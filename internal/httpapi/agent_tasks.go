package httpapi

import (
	"net/http"
	"strings"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type agentTaskListResponse struct {
	Tasks     []agentTaskDTO `json:"tasks"`
	RequestID string         `json:"request_id"`
}

type agentTaskDTO struct {
	ID            string `json:"id"`
	ThreadID      string `json:"thread_id"`
	RunID         string `json:"run_id"`
	Role          string `json:"role"`
	Goal          string `json:"goal"`
	Status        string `json:"status"`
	ResultSummary string `json:"result_summary,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func (s *Server) handleThreadAgentTasks(w http.ResponseWriter, r *http.Request, threadID string, suffix string) {
	if suffix != "agent-tasks" {
		writeAPIError(w, productdata.NewError(productdata.CodeThreadNotFound, "Thread not found."))
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	tasks, ok := s.product.(productdata.AgentTaskService)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Agent task service is unavailable."))
		return
	}
	items, err := tasks.ListAgentTasks(r.Context(), identity.LocalDevIdentity(), productdata.ListAgentTasksInput{ThreadID: threadID, Limit: intQuery(r, "limit", 20)})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agentTaskListResponse{Tasks: agentTaskDTOs(items), RequestID: diagnostics.NewRequestID()})
}

func agentTaskDTOs(tasks []productdata.AgentTask) []agentTaskDTO {
	items := make([]agentTaskDTO, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, agentTaskDTO{
			ID:            task.ID,
			ThreadID:      task.ThreadID,
			RunID:         task.RunID,
			Role:          task.Role,
			Goal:          task.Goal,
			Status:        string(task.Status),
			ResultSummary: strings.TrimSpace(task.ResultSummary),
			CreatedAt:     task.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     task.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	return items
}
