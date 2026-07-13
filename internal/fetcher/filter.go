package fetcher

import (
	"strings"
	"time"
)

// FilterCriteria hold what we filter against (derived from the user the user profile)
type FilterCriteria struct {
	Keywords [] string  //target roles + skills, lowercased
	MaxAge time.Duration // jobs older than this are dropped
}

// Relevant reports whether a job passes the cheap filter: recent enough AND mentions at least one keyword.

func (fc FilterCriteria) Relevant (j Job) bool {
	if !fc.RecentEnough(j) {
		return false
	}

	return fc.MatchesKeyword(j)
}

func (fc FilterCriteria) RecentEnough(j Job) bool {
	if j.PostedAt.IsZero() {
		return true // unknown date: don't drop it on stalenes alone
	}
	return time.Since(j.PostedAt) <= fc.MaxAge
}

func (fc FilterCriteria) MatchesKeyword(j Job) bool {
	haystack := strings.ToLower(j.Title)

	for _, kw := range fc.Keywords {
		if kw == ""{
			continue
		}
		if strings.Contains(haystack, kw){
			return  true
		}
	}
	return false
}

// Filter returns only jobs that pass the criteria
func Filter(jobs []Job, fc FilterCriteria) []Job {
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		if fc.Relevant(j) {
			out = append(out, j)
		}
	}
	return out
}