package fetcher

import (
	"testing"
	"time"
)


func TestFilterRelevant(t *testing.T) {
	criteria := FilterCriteria{
		Keywords: []string{"go", "backend", "postgresql"},
		MaxAge: 30 * 24 * time.Hour, // 30 days
	}

	now := time.Now()

	tests := []struct{
		name string 
		job Job
		want bool
	}{
		{
			name: "recent job matching keyword",
			job:  Job{Title: "Backend Engineer", Description: "We use Go", PostedAt: now},
			want: true,
		},
		{
			name: "recent job with no keyword match",
			job:  Job{Title: "Senior Nurse", Description: "Healthcare role", PostedAt: now},
			want: false,
		},
		{
			name: "keyword match but too old",
			job:  Job{Title: "Go Developer", Description: "", PostedAt: now.Add(-60 * 24 * time.Hour)},
			want: false,
		},
		{
			name: "zero date is kept if keyword matches",
			job:  Job{Title: "Backend Engineer", Description: "", PostedAt: time.Time{}},
			want: true,
		},
		{
			name: "case-insensitive keyword match",
			job:  Job{Title: "GOLANG DEVELOPER", Description: "", PostedAt: now},
			want: true,
		},
		{
			name: "keyword in description only",
			job:  Job{Title: "Software Engineer", Description: "PostgreSQL expert needed", PostedAt: now},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := criteria.Relevant(tc.job)
			if got != tc.want {
				t.Errorf("Relevant(%q) = %v, want %v", tc.job.Title, got, tc.want)
			}
		})
	}
}