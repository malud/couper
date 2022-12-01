package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"

	"github.com/avenga/couper/internal/test"
)

// TODO: relocate while refactoring integration tests

func TestDefinitions_Jobs(t *testing.T) {
	type testcase struct {
		name       string
		fileName   string
		origin     http.Handler
		wantErr    bool
		wantFields logrus.Fields
	}

	const basePath = "testdata/definitions"

	for _, tc := range []testcase{
		{"without label", "01_job.hcl", http.HandlerFunc(nil), true, nil},
		{"without interval", "02_job.hcl", http.HandlerFunc(nil), true, nil},
		{"variable reference", "03_job.hcl", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			payload := map[string]string{
				"prop1": "val1",
				"prop2": "val2",
			}
			b, _ := json.Marshal(payload)

			switch req.Method {
			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")
				w.Write(b)
			case http.MethodPost:
				r, _ := io.ReadAll(req.Body)
				if !bytes.Equal(b, r) {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
		}), false, logrus.Fields{"custom": logrus.Fields{
			"status_a": float64(http.StatusOK),
			"status_b": float64(http.StatusOK),
		}}},
	} {
		t.Run(tc.name, func(st *testing.T) {
			origin := httptest.NewServer(tc.origin)
			defer origin.Close()

			helper := test.New(st)

			shutdown, hook, err := newCouperWithTemplate(filepath.Join(basePath, tc.fileName), helper, map[string]interface{}{
				"origin": origin.URL,
			})

			if (err != nil) != tc.wantErr {
				st.Fatalf("want error: %v, got: %v", tc.wantErr, err)
			} else if tc.wantErr && err != nil {
				return
			}

			defer shutdown()

			time.Sleep(time.Second / 4)

			for _, entry := range hook.AllEntries() {
				if entry.Data["type"] == "job" {
					for k := range tc.wantFields {
						if diff := cmp.Diff(entry.Data[k], tc.wantFields[k]); diff != "" {
							st.Errorf("expected log fields %q:\n%v", k, diff)
						}
					}
					continue
				}
				if entry.Data["status"] != 200 {
					st.Errorf("expected status OK, got: %v", entry.Data["status"])
				}
			}

		})
	}
}
