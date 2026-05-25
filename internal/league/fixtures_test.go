package league

import (
	"testing"

	"github.com/mtunaswe/insider-league/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFixtures(t *testing.T) {
	teams := []domain.Team{
		{ID: 1, Name: "Chelsea", Strength: 85},
		{ID: 2, Name: "Manchester City", Strength: 80},
		{ID: 3, Name: "Arsenal", Strength: 70},
		{ID: 4, Name: "Liverpool", Strength: 60},
	}

	fixtures := GenerateFixtures(teams)

	t.Run("12 matches total", func(t *testing.T) {
		assert.Len(t, fixtures, 12)
	})

	t.Run("6 weeks with 2 matches each", func(t *testing.T) {
		weekCounts := make(map[int]int)
		for _, m := range fixtures {
			weekCounts[m.Week]++
		}
		require.Len(t, weekCounts, 6)
		for w := 1; w <= 6; w++ {
			assert.Equal(t, 2, weekCounts[w], "week %d should have 2 matches", w)
		}
	})

	t.Run("each unordered pair appears exactly twice", func(t *testing.T) {
		type pair struct{ a, b int }
		pairCounts := make(map[pair]int)
		for _, m := range fixtures {
			p := pair{m.HomeTeamID, m.AwayTeamID}
			if p.a > p.b {
				p.a, p.b = p.b, p.a
			}
			pairCounts[p]++
		}
		// C(4,2) = 6 unique pairs, each appearing twice
		assert.Len(t, pairCounts, 6)
		for p, count := range pairCounts {
			assert.Equal(t, 2, count, "pair (%d,%d) should appear exactly twice", p.a, p.b)
		}
	})

	t.Run("each ordered pair appears exactly once (home/away)", func(t *testing.T) {
		type orderedPair struct{ home, away int }
		seen := make(map[orderedPair]bool)
		for _, m := range fixtures {
			op := orderedPair{m.HomeTeamID, m.AwayTeamID}
			assert.False(t, seen[op], "ordered pair (%d,%d) should appear only once", op.home, op.away)
			seen[op] = true
		}
	})

	t.Run("no team plays itself", func(t *testing.T) {
		for _, m := range fixtures {
			assert.NotEqual(t, m.HomeTeamID, m.AwayTeamID)
		}
	})

	t.Run("all matches start unplayed", func(t *testing.T) {
		for _, m := range fixtures {
			assert.False(t, m.Played)
		}
	})
}
