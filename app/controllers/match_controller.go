package controllers

import (
	"encoding/json"
	"insider-case/app/services"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type MatchController struct {
	service services.IMatchService
}

func NewMatchController(service services.IMatchService) *MatchController {
	return &MatchController{service: service}
}
func (mc *MatchController) GetMatchesByLeagueIDAndWeek(w http.ResponseWriter, r *http.Request) {
	leagueIDStr := mux.Vars(r)["leagueID"]
	weekStr := mux.Vars(r)["week"]

	leagueID, err := strconv.ParseUint(leagueIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil {
		http.Error(w, "Invalid week number", http.StatusBadRequest)
		return
	}

	matches, err := mc.service.GetMatchesByLeagueIdAndWeek(uint(leagueID), week)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
func (mc *MatchController) GetMatchesByLeagueID(w http.ResponseWriter, r *http.Request) {
	leagueIDStr := mux.Vars(r)["leagueID"]

	leagueID, err := strconv.ParseUint(leagueIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	matches, err := mc.service.GetMatchesByLeagueId(uint(leagueID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
