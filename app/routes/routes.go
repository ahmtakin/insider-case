package routes

import (
	"insider-case/app/controllers"
	"insider-case/app/repository"
	"insider-case/app/services"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	teamRepo := repository.NewTeamRepository()
	matchRepo := repository.NewMatchRepository()
	teamStatsRepo := repository.NewTeamStatsRepository()
	weeklyLogRepo := repository.NewWeeklyLogRepository(teamStatsRepo)
	leagueRepo := repository.NewLeagueRepository(teamRepo, matchRepo, teamStatsRepo)
	matchService := services.NewMatchService(matchRepo, teamRepo, teamStatsRepo)

	leagueController := controllers.NewLeagueController(
		services.NewLeagueService(
			leagueRepo,
			matchService,
			teamStatsRepo,
			weeklyLogRepo,
			teamRepo,
		),
	)
	teamController := controllers.NewTeamController(
		services.NewTeamService(
			teamRepo,
			teamStatsRepo,
		),
	)
	matchController := controllers.NewMatchController(
		matchService,
	)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/leagues", leagueController.CreateLeague).Methods("POST")

	api.HandleFunc("/teams/{leagueID}", teamController.GetTeamsByLeagueID).Methods("GET")
	api.HandleFunc("/matches/{leagueID}/{week}", matchController.GetMatchesByLeagueIDAndWeek).Methods("GET")
	api.HandleFunc("/matches/{leagueID}", matchController.GetMatchesByLeagueID).Methods("GET")
	api.HandleFunc("/leagues/simulate-week", leagueController.SimulateWeek).Methods("POST")
	api.HandleFunc("/leagues/play-remaining-matches", leagueController.PlayRemainingMatches).Methods("POST")
	api.HandleFunc("/leagues/user-play-week", leagueController.UserPlayWeek).Methods("POST")
	api.HandleFunc("/leagues/championship-estimations", leagueController.GetChampionshipEstimations).Methods("GET")

	r.PathPrefix("/api").Handler(enableCORS(api))

	fileServer := uiFileServer()
	r.PathPrefix("/").Handler(fileServer)

}

func uiFileServer() http.Handler {
	fs := http.FileServer(http.Dir("app/ui"))

	// Wrap the file server to set correct MIME types
	fileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set MIME types for different file extensions
		switch filepath.Ext(r.URL.Path) {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".html":
			w.Header().Set("Content-Type", "text/html")
		}
		fs.ServeHTTP(w, r)
	})
	return fileServer
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace '*' with specific domain in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
