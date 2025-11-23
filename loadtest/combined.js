import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  scenarios: {
    mixes: {
      executor: 'constant-vus',
      vus: 20,
      duration: '1m'
    }
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
    http_req_duration: ['p(95)<300']
  }
};

const BASE = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  // 60% GET /users/getReview
  if (Math.random() < 0.6) {
    let res = http.get(`${BASE}/users/getReview?user_id=u002`);
    check(res, {"team get status is 200": (r) => r.status === 200});
  } else if (Math.random() < 0.8) { // 20% POST create
    const prId = `pr-${__VU}-${__ITER}-${Date.now()}`;
    let res = http.post(`${BASE}/pullRequest/create`,
      JSON.stringify({ pull_request_id: prId, pull_request_name: "feat", author_id: "u020" }),
      { headers: { "Content-Type": "application/json" } });
    check(res, {
        "create pr 201 or 409 or 404": (r) => r.status === 201 || r.status === 404 || r.status === 409
    });
  } else { // 20% reassign
    let res = http.post(`${BASE}/pullRequest/reassign`,
      JSON.stringify({ pull_request_id: "pr011", old_user_id: "u019" }),
      { headers: { "Content-Type": "application/json" } });
    check(res, {
        "reassign ok or 404 or 409": (r) => r.status === 200 || r.status === 404 || r.status === 409
    });
  }

  sleep(0.1);
}
