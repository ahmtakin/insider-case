package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"insider-case/app/dto"
	"insider-case/app/services"
)

type LeagueController struct {
	service services.ILeagueService
}

func NewLeagueController(service services.ILeagueService) *LeagueController {
	return &LeagueController{service: service}
}

func (lc *LeagueController) CreateLeague(w http.ResponseWriter, r *http.Request) {
	var req dto.LeagueCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := lc.service.InitializeLeague(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (lc *LeagueController) SimulateWeek(w http.ResponseWriter, r *http.Request) {
	leagueIDStr := r.URL.Query().Get("leagueID")
	if leagueIDStr == "" {
		http.Error(w, "leagueID is required", http.StatusBadRequest)
		return
	}

	leagueID, err := strconv.ParseUint(leagueIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	week, err := lc.service.SimulateWeek(uint(leagueID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(week)
}
