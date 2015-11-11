package tally

import (
	"strconv"

	"appengine"
	"appengine/datastore"

	"hunt"
	"puzzle"
	"team"
)

// normalize paper/non-paper separately
// 10 points per other team in the game per puzzle type

type VotePoints struct {
	Fun float32
	Presentation float32
	Ingredients float32
	Score float32
}

type FinalScore struct {
	Name string
	SolvePoints []int
	TotalSolvePoints int
	NonPaperPoints VotePoints
	PaperPoints VotePoints
	TotalVotingPoints float32
	FinalScore float32
}

const (
	finalScoreKind = "finalscore"
)

func BuildFinalTally(c appengine.Context, h *hunt.Hunt) {
	teams := team.All(c, h)
	puzzles := puzzle.All(c, h, nil)
	teamMap := map[string]*FinalScore{}
	for _, t := range teams {
		fs := &FinalScore{Name: t.Name}
		for _, p := range puzzles {
			if solve := puzzle.GetSolve(c, h, p, t); solve != nil {
				fs.SolvePoints = append(fs.SolvePoints, solve.Points)
				fs.TotalSolvePoints += solve.Points
			} else {
				fs.SolvePoints = append(fs.SolvePoints, 0)
			}
		}
		teamMap[t.Name] = fs
	}
	puzzleMap := map[int]*VotePoints{}
	for _, p := range puzzles {
		for _, t := range teams {
			if p.Team.Equal(t.Key) {
				if p.Paper {
					puzzleMap[p.Number] = &teamMap[t.Name].PaperPoints
				} else {
					puzzleMap[p.Number] = &teamMap[t.Name].NonPaperPoints
				}
			}
		}
	}

	numVotes := len(puzzles) - 2
	for _, t := range teams {
		var votes []float32
		if len(t.Survey) != 3 * numVotes {
			for i := 0; i < 3 * numVotes; i++ {
				votes = append(votes, 1.0)
			}
		} else {
			for _, c := range t.Survey {
				points, _ := strconv.Atoi(string(c))
				if points < 1 {
					points = 1
				}
				if points > 5 {
					points = 5
				}
				votes = append(votes, float32(points))
			}
		}
		normalize(votes[0:len(votes)/2])
		normalize(votes[len(votes)/2:len(votes)])
		voteIndex := 0
		for _, p := range puzzles {
			if p.Team.Equal(t.Key) {
				continue
			}
			puzzleMap[p.Number].Fun += votes[voteIndex]
			puzzleMap[p.Number].Presentation += votes[voteIndex+1]
			puzzleMap[p.Number].Ingredients += votes[voteIndex+2]
			voteIndex += 3
		}
	}

	for _, fs := range teamMap {
		fs.TotalVotingPoints = fs.NonPaperPoints.Fun + fs.NonPaperPoints.Presentation + fs.NonPaperPoints.Ingredients + fs.PaperPoints.Fun + fs.PaperPoints.Presentation + fs.PaperPoints.Ingredients
		fs.FinalScore = float32(fs.TotalSolvePoints) + fs.TotalVotingPoints
		k := datastore.NewIncompleteKey(c, finalScoreKind, h.Key)
		k, err := datastore.Put(c, k, fs)
		if err != nil {
			c.Errorf("Error: %v", err)
		}
	}
}

func Get(c appengine.Context, h *hunt.Hunt) []*FinalScore {
	var fs []*FinalScore
	_, err := datastore.NewQuery(finalScoreKind).Ancestor(h.Key).GetAll(c, &fs)
	if err != nil {
		c.Errorf("Error: %v", err)
	}
	return fs
}

func normalize(arr []float32) {
	var sum float32
	for i := range arr {
		if i % 3 == 0 {
			arr[i] = arr[i] * 2
		}
		sum += arr[i]
	}
	// Average 10 pts per puzzle
	target := float32(len(arr) / 3 * 10)
	ratio := target / sum
	for i := range arr {
		arr[i] = arr[i] * ratio
	}
}
