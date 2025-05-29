// script.js

let teamMap = {};
let currentLeague = null;
let currentWeek = 1;
let allMatches = [];
let teamCount = 0;


// Dynamic team inputs based on count
const teamInputsContainer = document.getElementById('teams-inputs');
const teamCountInput = document.getElementById('team-count');
teamCountInput.addEventListener('input', generateTeamInputs);


function generateTeamInputs() {
    teamInputsContainer.innerHTML = '';
    const count = parseInt(teamCountInput.value);
    for (let i = 0; i < count; i++) {
        teamInputsContainer.innerHTML += `
            <div class="team-input">
                <label>Team Name: <input type="text" class="team-name" value="Team ${i + 1}" required></label>
                <label>Strength: <input type="number" class="team-strength" value="1000" min="1000" max="3000" required></label>
            </div>`;
    }
}

// Initial call
generateTeamInputs();

// Create league
const leagueForm = document.getElementById('league-form');
leagueForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    const name = document.getElementById('league-name').value;
    const teamNames = document.querySelectorAll('.team-name');
    const teamStrengths = document.querySelectorAll('.team-strength');
    teamCount = teamNames.length;

    const teams = Array.from(teamNames).map((el, idx) => ({
        name: el.value,
        strength: parseInt(teamStrengths[idx].value)
    }));

    const res = await fetch('https://insider-case.onrender.com/api/leagues', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ name, team_count: teams.length, teams })
    });

    const data = await res.json();
    currentLeague = data;
    currentWeek = data.curr_week;
    teamMap = {};
    data.teams.forEach(team => teamMap[team.id] = team.name);


    allMatches = data.matches;
    console.log(allMatches);

    document.getElementById('setup-section').classList.add('hidden');
    document.getElementById('dashboard').classList.remove('hidden');

    updateDashboard(data);
});

function updateDashboard(data, champion = null) {
    renderTeamStats(data.teams);
    renderChampionPredictions(data.teams);
    renderMatches(data.matches);
    renderNextWeekMatches(allMatches, champion); 
}

function renderTeamStats(teams) {
    const table = document.getElementById('team-stats-table');
    table.innerHTML = '<tr><th>Team</th><th>P</th><th>Pts</th><th>W</th><th>L</th><th>D</th><th>GF</th><th>GA</th><th>GD</th></tr>';

    // Sort teams by points in descending order
    const sortedTeams = [...teams].sort((a, b) => (b.stats?.points || 0) - (a.stats?.points || 0));

    sortedTeams.forEach(team => {
        const stats = team.stats || {};
        table.innerHTML += `<tr>
            <td>${team.name}</td>
            <td>${stats.played || 0}</td>
            <td>${stats.points || 0}</td>
            <td>${stats.won || 0}</td>
            <td>${stats.lost || 0}</td>
            <td>${stats.draw || 0}</td>
            <td>${stats.goals_for || 0}</td>
            <td>${stats.goals_against || 0}</td>
            <td>${stats.goal_diff || 0}</td>
        </tr>`;
    });
}

function renderChampionPredictions(teams) {
    const container = document.getElementById('champion-predictions');
    container.innerHTML = '';
    teams.forEach(team => {
        const estimationPercent = (team.stats.estimation * 100).toFixed(2);
        container.innerHTML += `<div class="prediction">${team.name}: ${estimationPercent}%</div>`;
    });
}

function renderMatches(matches) {
    const container = document.getElementById('all-matches');
    container.innerHTML = '';
    matches.forEach(match => {
        container.innerHTML += `<div>
            Week ${match.week} - ${teamMap[match.home_team]} vs ${teamMap[match.away_team]}:
            ${match.played ? `${match.home_score} - ${match.away_score}` : 'Not Played'}
        </div>`;
    });
}

function renderNextWeekMatches(matches, champion) {
    const container = document.getElementById('next-matches');
    container.innerHTML = '';
    if (matches.length > 0 && currentWeek < 2 * (teamCount - 1) + 1) {
        const nextMatches = matches.slice(currentWeek * 2 - 2, currentWeek * 2);
        container.innerHTML = `<h3>Next Matches (Week ${currentWeek})</h3>`;
        for (const match of nextMatches) {
            container.innerHTML += `<div>
                Week ${currentWeek} - ${teamMap[match.home_team]} vs ${teamMap[match.away_team]}
            </div>`;
        }
       
    }else if (currentWeek == 2*(teamCount-1)+1) {
        if (champion) {
            container.innerHTML = `<h3>${currentLeague.name} has finished!!!</h3>`;
            container.innerHTML += `<div class="champion-display">
                Current Champion: ${champion.name}
            </div>`;
        }
    } 
    else {

        container.innerHTML = '<div>No matches scheduled for next week.</div>';
    }
}

document.getElementById('simulate-week').addEventListener('click', async () => {
    const res = await fetch('https://insider-case.onrender.com/api/leagues/simulate-week', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ leagueID: currentLeague.id })
    });
    const data = await res.json();
    currentWeek = data.week;
    updateDashboard({ ...currentLeague, matches: data.matches, teams: mergeStats(currentLeague.teams, data.team_stats) }, data.champion);
});

function mergeStats(teams, stats) {
    return teams.map(team => {
        const newStats = stats.find(s => s.team_id === team.id);
        return { ...team, stats: newStats };
    });
}

document.getElementById('play-all').addEventListener('click', async () => {
    const res = await fetch('https://insider-case.onrender.com/api/leagues/play-remaining-matches', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ leagueID: currentLeague.id })
    });
    const allWeeks = await res.json();
    const allMatches = allWeeks.flatMap(w => w.matches);
    const lastStats = allWeeks[allWeeks.length - 1].team_stats;
    currentWeek = allWeeks[allWeeks.length - 1].week+1;
    const champion = allWeeks[allWeeks.length - 1].champion;
    updateDashboard({ ...currentLeague, matches: allMatches, teams: mergeStats(currentLeague.teams, lastStats) }, champion);
});

document.getElementById('manual-week').addEventListener('click', () => {
    const modal = document.getElementById('manual-week-modal');
    const container = document.getElementById('match-results-container');
    const currentWeekDisplay = document.getElementById('current-week-display');

    // Show the modal
    modal.classList.remove('hidden');

    // Set the current week display
    currentWeekDisplay.textContent = currentWeek;

    // Clear previous inputs
    container.innerHTML = '';

    // Get matches for the current week
    const matchesForWeek = allMatches.filter(match => match.week === currentWeek);

    // Generate input fields for each match
    matchesForWeek.forEach(match => {
        const matchInput = document.createElement('div');
        matchInput.innerHTML = `
            <label>${teamMap[match.home_team]} vs ${teamMap[match.away_team]}:</label>
            <input type="number" placeholder="Home Score" data-match-id="${match.id}" data-home-team="${match.home_team}" data-away-team="${match.away_team}" class="home-score" required>
            <input type="number" placeholder="Away Score" class="away-score">
        `;
        container.appendChild(matchInput);
    });
});

document.getElementById('close-modal').addEventListener('click', () => {
    const modal = document.getElementById('manual-week-modal');
    modal.classList.add('hidden');
});

document.getElementById('manual-week-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const inputs = document.querySelectorAll('#match-results-container div');
    const results = Array.from(inputs).map(input => {
        const homeScore = input.querySelector('.home-score').value;
        const awayScore = input.querySelector('.away-score').value;
        const matchId = input.querySelector('.home-score').dataset.matchId;
        const homeTeamId = input.querySelector('.home-score').dataset.homeTeam;
        const awayTeamId = input.querySelector('.home-score').dataset.awayTeam;

        return {
            league_id: currentLeague.id,
            week: currentWeek,
            match_id: parseInt(matchId),
            home_team_id: parseInt(homeTeamId),
            away_team_id: parseInt(awayTeamId),
            home_score: parseInt(homeScore),
            away_score: parseInt(awayScore)
        };
    });

    const res = await fetch('https://insider-case.onrender.com/api/leagues/user-play-week', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(results)
    });

    if (res.ok) {
        const data = await res.json();
        currentWeek = data.week;
        const champion = data.champion || null;
        updateDashboard({ ...currentLeague, matches: data.matches, teams: mergeStats(currentLeague.teams, data.team_stats) }, champion);
        document.getElementById('manual-week-modal').classList.add('hidden');
    } else {
        alert('Failed to submit match results.');
    }
});
