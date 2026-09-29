package controllers

import (
	"net/http"

	"auth/internal/utils"
)

func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieHostRefresh)
	if err != nil {
		utils.JSONResponse(w, http.StatusUnauthorized, false, nil, "Refresh token missing")
		return
	}

	resp, err := c.authService.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		c.clearAuthCookies(w)
		utils.JSONResponse(w, http.StatusUnauthorized, false, nil, "Invalid or expired refresh token")
		return
	}

	c.setAuthCookies(w, resp.Token, resp.RefreshToken)
	utils.JSONResponse(w, http.StatusOK, true, resp, nil)
}
