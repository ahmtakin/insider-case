package routes

import (
	"insider-case/app/controllers"
	"insider-case/app/repository"
	"insider-case/app/services"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	leagueController := controllers.NewLeagueController(
		services.NewLeagueService(
			repository.NewLeagueRepository(
				repository.NewTeamRepository(),
				repository.NewMatchRepository(),
				repository.NewTeamStatsRepository(),
			),
			services.NewMatchService(
				repository.NewMatchRepository(),
				repository.NewTeamRepository(),
				repository.NewTeamStatsRepository(),
			),
		),
	)
	teamController := controllers.NewTeamController(
		services.NewTeamService(
			repository.NewTeamRepository(),
			repository.NewTeamStatsRepository(),
		),
	)
	matchController := controllers.NewMatchController(
		services.NewMatchService(
			repository.NewMatchRepository(),
			repository.NewTeamRepository(),
			repository.NewTeamStatsRepository(),
		),
	)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/leagues", leagueController.CreateLeague).Methods("POST")

	api.HandleFunc("/teams/{leagueID}", teamController.GetTeamsByLeagueID).Methods("GET")
	api.HandleFunc("/matches/{leagueID}/{week}", matchController.GetMatchesByLeagueIDAndWeek).Methods("GET")
	api.HandleFunc("/matches/{leagueID}", matchController.GetMatchesByLeagueID).Methods("GET")
	api.HandleFunc("/leagues/simulate-week", leagueController.SimulateWeek).Methods("GET")
}
