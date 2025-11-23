package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

var now = time.Now().Format("2006-01-02-15-04-05")

func postJSON(t *testing.T, url string, payload interface{}) *http.Response {
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	assert.NoError(t, err)
	return resp
}

func getJSON(t *testing.T, url string) *http.Response {
	resp, err := http.Get(url)
	assert.NoError(t, err)
	return resp
}

func TestCreateTeamAndUsers(t *testing.T) {
	payload := map[string]interface{}{
		"team_name": "e2e_team_" + now,
		"members": []map[string]interface{}{
			{"user_id": "u1_" + now, "username": "Alice", "is_active": true},
			{"user_id": "u2_" + now, "username": "Bob", "is_active": true},
		},
	}

	resp := postJSON(t, baseURL+"/team/add", payload)
	assert.Equal(t, 201, resp.StatusCode)

	resp = getJSON(t, baseURL+"/team/get?team_name=e2e_team_"+now)
	assert.Equal(t, 200, resp.StatusCode)

	var teamResp map[string]map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&teamResp)
	assert.Equal(t, "e2e_team_"+now, teamResp["team"]["team_name"])
	assert.Equal(t, 2, len(teamResp["team"]["members"].([]interface{})))
}

func TestCreatePR(t *testing.T) {
	payload := map[string]string{
		"pull_request_id":   "pr-e2e-1_" + now,
		"pull_request_name": "Add E2E feature",
		"author_id":         "u1_" + now,
	}

	resp := postJSON(t, baseURL+"/pullRequest/create", payload)
	assert.Equal(t, 201, resp.StatusCode)

	var prResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&prResp)
	pr := prResp["pr"].(map[string]interface{})
	assigned := pr["assigned_reviewers"].([]interface{})
	assert.True(t, len(assigned) <= 2)
}

func TestMergePR(t *testing.T) {
	payload := map[string]string{"pull_request_id": "pr-e2e-1_" + now}
	resp := postJSON(t, baseURL+"/pullRequest/merge", payload)
	assert.Equal(t, 200, resp.StatusCode)

	reassignPayload := map[string]string{
		"pull_request_id": "pr-e2e-1_" + now,
		"old_user_id":     "u2_" + now,
	}
	resp = postJSON(t, baseURL+"/pullRequest/reassign", reassignPayload)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestStatsPR(t *testing.T) {
	resp := getJSON(t, baseURL+"/stats/pullRequests")
	assert.Equal(t, 200, resp.StatusCode)

	var teamResp map[string]map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&teamResp)
	assert.Equal(t, float64(1), teamResp["statistic_count_reviewers"]["pr-e2e-1_"+now])
}
