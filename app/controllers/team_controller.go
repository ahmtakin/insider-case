package controllers

import (
	"encoding/json"
	"insider-case/app/services"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ITeamController interface {
	GetTeamsByLeagueID(w http.ResponseWriter, r *http.Request)
}
type TeamController struct {
	service services.ITeamService
}

func NewTeamController(service services.ITeamService) *TeamController {
	return &TeamController{service: service}
}
func (tc *TeamController) GetTeamsByLeagueID(w http.ResponseWriter, r *http.Request) {
	leagueIDStr := mux.Vars(r)["leagueID"]
	leagueID, err := strconv.ParseUint(leagueIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid league ID", http.StatusBadRequest)
		return
	}

	teams, err := tc.service.GetTeamsByLeagueID(uint(leagueID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}
