package league

import "errors"

var (
	ErrSeasonComplete = errors.New("season is already complete")
	ErrMatchNotPlayed = errors.New("match has not been played yet")
)
