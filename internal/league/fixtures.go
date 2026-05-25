package league

import "github.com/mtunaswe/insider-league/internal/domain"

func GenerateFixtures(teams []domain.Team) []domain.Match {
	n := len(teams)
	var matches []domain.Match
	week := 1

	// First half: each pair plays once (home/away assigned by round-robin circle method)
	// With 4 teams we get 3 rounds × 2 matches = 6 matches in weeks 1-3
	for round := 0; round < n-1; round++ {
		roundMatches := circleRound(teams, round)
		for _, m := range roundMatches {
			m.Week = week
			matches = append(matches, m)
		}
		week++
	}

	// Second half: reverse home/away for weeks 4-6
	firstHalfCount := len(matches)
	for i := 0; i < firstHalfCount; i++ {
		m := matches[i]
		matches = append(matches, domain.Match{
			Week:       m.Week + (n - 1),
			HomeTeamID: m.AwayTeamID,
			AwayTeamID: m.HomeTeamID,
			Played:     false,
		})
	}

	return matches
}

func circleRound(teams []domain.Team, round int) []domain.Match {
	n := len(teams)
	ids := make([]int, n)
	ids[0] = teams[0].ID
	for i := 1; i < n; i++ {
		idx := 1 + (i-1+round)%(n-1)
		ids[i] = teams[idx].ID
	}

	var matches []domain.Match
	for i := 0; i < n/2; i++ {
		home := ids[i]
		away := ids[n-1-i]
		matches = append(matches, domain.Match{
			HomeTeamID: home,
			AwayTeamID: away,
			Played:     false,
		})
	}
	return matches
}
