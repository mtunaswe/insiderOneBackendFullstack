package domain

import "sort"

func ComputeStandings(teams []Team, matches []Match) []StandingsRow {
	rowMap := make(map[int]*StandingsRow, len(teams))
	for _, t := range teams {
		rowMap[t.ID] = &StandingsRow{
			TeamID:   t.ID,
			TeamName: t.Name,
		}
	}

	for _, m := range matches {
		if !m.Played || m.HomeGoals == nil || m.AwayGoals == nil {
			continue
		}

		home := rowMap[m.HomeTeamID]
		away := rowMap[m.AwayTeamID]
		hg, ag := *m.HomeGoals, *m.AwayGoals

		home.P++
		away.P++
		home.GF += hg
		home.GA += ag
		away.GF += ag
		away.GA += hg

		switch {
		case hg > ag:
			home.W++
			home.Pts += PointsWin
			away.L++
		case hg < ag:
			away.W++
			away.Pts += PointsWin
			home.L++
		default:
			home.D++
			away.D++
			home.Pts += PointsDraw
			away.Pts += PointsDraw
		}
	}

	rows := make([]StandingsRow, 0, len(teams))
	for _, r := range rowMap {
		r.GD = r.GF - r.GA
		rows = append(rows, *r)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Pts != rows[j].Pts {
			return rows[i].Pts > rows[j].Pts
		}
		if rows[i].GD != rows[j].GD {
			return rows[i].GD > rows[j].GD
		}
		return rows[i].GF > rows[j].GF
	})

	return rows
}
