package fetcher

import "time"

// Job is the normalised representation of a single listing,
// independent of which source it came from.

type Job struct {
	Title			string
	Company 		string
	Source  		string 
	URL				string 
	Description 	string 
	Location		string 
	Remote 			bool
	PostedAt		time.Time
	Hash			string
	RawJSON			[] byte

}