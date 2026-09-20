package main

import "testing"

// The socket form is easy to break by "simplifying" buildDSN back to a single
// Sprintf: an absolute host in the URL authority parses without error and then
// fails at connect time, on Cloud Run only, with a misleading message.
func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "tcp host keeps the requested sslmode",
			host: "10.0.0.5",
			want: "postgres://medilog:p%40ss@10.0.0.5:5432/medilog?sslmode=require",
		},
		{
			// Cloud SQL mounts the instance as a socket directory; lib/pq only
			// reads that from the host query parameter, and cannot negotiate
			// TLS over it.
			name: "cloudsql socket moves host into the query and disables tls",
			host: "/cloudsql/proj:us-central1:medilog",
			want: "postgres://medilog:p%40ss@/medilog?host=%2Fcloudsql%2Fproj%3Aus-central1%3Amedilog&port=5432&sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDSN(tt.host, "5432", "medilog", "p@ss", "medilog", "require")
			if got != tt.want {
				t.Errorf("buildDSN(%q)\n got: %s\nwant: %s", tt.host, got, tt.want)
			}
		})
	}
}
