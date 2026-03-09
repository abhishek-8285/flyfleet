package admin_support

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// Organization Management
	r.Route("/orgs", func(r chi.Router) {
		r.Post("/", h.createOrg)
		r.Get("/", h.listOrgs)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getOrg)
			r.Post("/suspend", h.suspendOrg)
			r.Post("/activate", h.activateOrg)
			r.Post("/users", h.addOrgUser)
			r.Get("/users", h.listOrgUsers)
		})
	})

	// Support Tickets
	r.Route("/tickets", func(r chi.Router) {
		r.Post("/", h.createTicket)
		r.Get("/", h.listTickets)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getTicket)
			r.Post("/reply", h.replyTicket)
			r.Post("/assign", h.assignTicket)
			r.Post("/close", h.closeTicket)
			r.Post("/attachments", h.addTicketAttachment)
			r.Get("/attachments", h.listTicketAttachments)
		})
	})

	// Incidents
	r.Route("/incidents", func(r chi.Router) {
		r.Post("/", h.declareIncident)
		r.Get("/", h.listIncidents)
		r.Route("/{id}", func(r chi.Router) {
			r.Post("/update", h.updateIncident)
			r.Post("/resolve", h.resolveIncident)
		})
	})

	// System Controls
	r.Route("/system", func(r chi.Router) {
		r.Post("/flag", h.setFlag)
		r.Get("/flags", h.listFlags)
		r.Post("/maintenance", h.startMaintenance)
	})

	// RBAC Management
	r.Route("/rbac", func(r chi.Router) {
		r.Get("/permissions", h.listPermissions)
		r.Get("/roles/{role}/permissions", h.getRolePermissions)
		r.Post("/roles/{role}/permissions", h.setRolePermissions)
	})

	// Impersonation
	r.Route("/impersonate", func(r chi.Router) {
		r.Post("/start", h.startImpersonation)
		r.Post("/stop", h.stopImpersonation)
	})

	// DR & Approvals
	r.Route("/dr", func(r chi.Router) {
		r.Post("/failover/request", h.requestFailover)
		r.Post("/failover/approve/{id}", h.approveFailover)
		r.Post("/failover/reject/{id}", h.rejectFailover)
		r.Get("/approvals/pending", h.listPendingApprovals)
	})

	// Cross-Domain Search
	r.Get("/search", h.searchAdminData)

	return r
}

// ─── Organization Handlers ───────────────────────────────────────────────────

func (h *Handler) createOrg(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		PlanType string `json:"plan_type"`
		Region   string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org, err := h.service.CreateOrganization(r.Context(), req.Name, req.PlanType, req.Region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) listOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.service.ListOrganizations(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(orgs)
}

func (h *Handler) getOrg(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	org, err := h.service.GetOrganization(r.Context(), id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) suspendOrg(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	// ActorID would normally come from context (middleware extracted from token)
	actorID := uuid.New() 

	if err := h.service.SuspendOrganization(r.Context(), id, actorID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activateOrg(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	actorID := uuid.New()

	if err := h.service.ActivateOrganization(r.Context(), id, actorID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addOrgUser(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		Email string           `json:"email"`
		Role  OrganizationRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.AddOrganizationUser(r.Context(), orgID, req.Email, req.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) listOrgUsers(w http.ResponseWriter, r *http.Request) {
	orgID, _ := uuid.Parse(chi.URLParam(r, "id"))
	users, err := h.service.ListOrganizationUsers(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

// ─── Support Ticket Handlers ─────────────────────────────────────────────────

func (h *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID    uuid.UUID      `json:"organization_id"`
		Subject  string         `json:"subject"`
		Priority TicketPriority `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// createdBy would come from context
	createdBy := uuid.New()

	ticket, err := h.service.CreateTicket(r.Context(), req.OrgID, req.Subject, req.Priority, createdBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("organization_id")
	var orgID *uuid.UUID
	if orgIDStr != "" {
		id, err := uuid.Parse(orgIDStr)
		if err == nil {
			orgID = &id
		}
	}

	tickets, err := h.service.ListTickets(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tickets)
}

func (h *Handler) getTicket(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	ticket, err := h.service.GetTicket(r.Context(), id)
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(ticket)
}

func (h *Handler) replyTicket(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	senderID := uuid.New()
	if err := h.service.ReplyToTicket(r.Context(), id, senderID, req.Message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) assignTicket(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		AdminID uuid.UUID `json:"admin_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.AssignTicket(r.Context(), ticketID, req.AdminID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) closeTicket(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	if err := h.service.ResolveTicket(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addTicketAttachment(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		FileURL string `json:"file_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadedBy := uuid.New() // Would come from auth context
	att, err := h.service.AddTicketAttachment(r.Context(), ticketID, req.FileURL, uploadedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(att)
}

func (h *Handler) listTicketAttachments(w http.ResponseWriter, r *http.Request) {
	ticketID, _ := uuid.Parse(chi.URLParam(r, "id"))
	atts, err := h.service.ListTicketAttachments(r.Context(), ticketID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(atts)
}

// ─── Incident Handlers ───────────────────────────────────────────────────────

func (h *Handler) declareIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Severity    IncidentSeverity `json:"severity"`
		Description string           `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inc, err := h.service.DeclareIncident(r.Context(), req.Severity, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inc)
}

func (h *Handler) listIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.service.ListIncidents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(incidents)
}

func (h *Handler) updateIncident(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateIncident(r.Context(), id, req.Message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resolveIncident(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	if err := h.service.ResolveIncident(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── System Control Handlers ─────────────────────────────────────────────────

func (h *Handler) setFlag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	actorID := uuid.New()
	if err := h.service.SetSystemFlag(r.Context(), req.Key, req.Value, actorID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.service.ListSystemFlags(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(flags)
}

func (h *Handler) startMaintenance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service     string `json:"service"`
		DurationSec int    `json:"duration_seconds"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	window, err := h.service.StartMaintenance(r.Context(), req.Service, time.Duration(req.DurationSec)*time.Second, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(window)
}

// ─── Impersonation Handlers ──────────────────────────────────────────────────

func (h *Handler) startImpersonation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adminID := uuid.New()
	session, err := h.service.StartImpersonation(r.Context(), adminID, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *Handler) stopImpersonation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID uuid.UUID `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.StopImpersonation(r.Context(), req.SessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── RBAC Handlers ───────────────────────────────────────────────────────────

func (h *Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	// Logic to list all possible permissions
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode([]string{"org.manage", "ticket.manage", "incident.manage", "system.override", "dr.execute"})
}

func (h *Handler) getRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := chi.URLParam(r, "role")
	perms, err := h.service.GetRolePermissions(r.Context(), role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(perms)
}

func (h *Handler) setRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := chi.URLParam(r, "role")
	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.SetRolePermissions(r.Context(), role, req.Permissions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Disaster Recovery & Two-Person Approvals ───────────────────────────────

func (h *Handler) requestFailover(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Region string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	requestedBy := uuid.New() // Normally from auth context
	app, err := h.service.RequestFailover(r.Context(), req.Region, requestedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

func (h *Handler) approveFailover(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	approvedBy := uuid.New() // Normally from auth context

	if err := h.service.ApproveFailover(r.Context(), id, approvedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) rejectFailover(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(chi.URLParam(r, "id"))
	rejectedBy := uuid.New() // Normally from auth context

	if err := h.service.RejectFailover(r.Context(), id, rejectedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listPendingApprovals(w http.ResponseWriter, r *http.Request) {
	apps, err := h.service.ListPendingApprovals(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(apps)
}

// ─── Cross-Domain Search ───────────────────────────────────────────────────

func (h *Handler) searchAdminData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	results, err := h.service.SearchAdminData(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}
