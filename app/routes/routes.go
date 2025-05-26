package routes

import (
	"insider-case/app/controllers"
	"insider-case/app/repository"
	"insider-case/app/services"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	leagueRepo := repository.NewLeagueRepository(repository.NewTeamRepository(), repository.NewMatchRepository(), repository.NewTeamStatsRepository())
	leagueService := services.NewLeagueService(leagueRepo)
	leagueController := controllers.NewLeagueController(leagueService)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/leagues", leagueController.CreateLeague).Methods("POST")
}
