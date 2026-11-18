package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/api/proto/authpb"
)

// RegisterInternalRoutes mounts HTTP JSON equivalents of internal gRPC directory operations
// for Cloud Run service-to-service communication (AUTH_SERVICE_URL).
func (c *GrpcAuthController) RegisterInternalRoutes(r chi.Router) {
	r.Get("/participants/{id}", c.httpGetParticipant)
	r.Get("/companies/{id}", c.httpGetCompany)
	r.Post("/users/active-by-roles", c.httpGetActiveUserIDsByRoles)
	r.Get("/users/{id}/device-tokens", c.httpGetUserDeviceTokens)
	r.Post("/users/{id}/device-tokens/revoke", c.httpRevokeDeviceTokens)
	r.Get("/campaigns/recipients", c.httpGetCampaignRecipients)
	r.Get("/users/{id}/email-verified", c.httpCheckUserEmailVerified)
	r.Get("/users/{id}/contact", c.httpGetUserContact)
	r.Post("/action-links", c.httpIssueActionLink)
}

func (c *GrpcAuthController) httpGetParticipant(w http.ResponseWriter, r *http.Request) {
	resp, err := c.GetParticipant(r.Context(), &authpb.GetParticipantRequest{
		ParticipantId: chi.URLParam(r, "id"),
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpGetCompany(w http.ResponseWriter, r *http.Request) {
	resp, err := c.GetCompany(r.Context(), &authpb.GetCompanyRequest{
		CompanyId: chi.URLParam(r, "id"),
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpGetActiveUserIDsByRoles(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RoleIDs []string `json:"role_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeGrpcError(w, status.Error(codes.InvalidArgument, "invalid request body"))
		return
	}
	resp, err := c.GetActiveUserIDsByRoles(r.Context(), &authpb.GetActiveUserIDsByRolesRequest{
		RoleIds: body.RoleIDs,
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpGetUserDeviceTokens(w http.ResponseWriter, r *http.Request) {
	resp, err := c.GetUserDeviceTokens(r.Context(), &authpb.GetUserDeviceTokensRequest{
		UserId: chi.URLParam(r, "id"),
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpRevokeDeviceTokens(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tokens []string `json:"tokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeGrpcError(w, status.Error(codes.InvalidArgument, "invalid request body"))
		return
	}
	resp, err := c.RevokeDeviceTokens(r.Context(), &authpb.RevokeDeviceTokensRequest{
		UserId: chi.URLParam(r, "id"),
		Tokens: body.Tokens,
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpGetCampaignRecipients(w http.ResponseWriter, r *http.Request) {
	resp, err := c.GetCampaignRecipients(r.Context(), &authpb.GetCampaignRecipientsRequest{})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpCheckUserEmailVerified(w http.ResponseWriter, r *http.Request) {
	resp, err := c.CheckUserEmailVerified(r.Context(), &authpb.CheckUserEmailVerifiedRequest{
		UserId: chi.URLParam(r, "id"),
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpGetUserContact(w http.ResponseWriter, r *http.Request) {
	resp, err := c.GetUserContact(r.Context(), &authpb.GetUserContactRequest{
		UserId: chi.URLParam(r, "id"),
	})
	writeGrpcJSON(w, resp, err)
}

func (c *GrpcAuthController) httpIssueActionLink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RequestID string `json:"request_id"`
		UserID    string `json:"user_id"`
		Type      string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeGrpcError(w, status.Error(codes.InvalidArgument, "invalid request body"))
		return
	}
	resp, err := c.IssueActionLink(r.Context(), &authpb.IssueActionLinkRequest{
		RequestId: body.RequestID,
		UserId:    body.UserID,
		Type:      body.Type,
	})
	writeGrpcJSON(w, resp, err)
}

func writeGrpcJSON(w http.ResponseWriter, payload any, err error) {
	if err != nil {
		writeGrpcError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeGrpcError(w http.ResponseWriter, err error) {
	httpStatus := http.StatusInternalServerError
	switch status.Code(err) {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.Unimplemented:
		httpStatus = http.StatusNotImplemented
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": status.Convert(err).Message()})
}
