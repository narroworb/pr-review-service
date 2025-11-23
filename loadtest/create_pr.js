import http from "k6/http";
import { check } from "k6";

export let options = {
  vus: 10,          
  duration: "30s",
  rate: 5,
  thresholds: {
    http_req_failed: ["rate<0.001"],
    http_req_duration: ["p(95)<300"]
  }
};

const BASE = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  const prId = `pr-${__VU}-${__ITER}-${Date.now()}`;
  const payload = JSON.stringify({
    pull_request_id: prId,
    pull_request_name: "Add search",
    author_id: "u010"
  });

  let res = http.post(`${BASE}/pullRequest/create`, payload, { headers: { "Content-Type": "application/json" } });
  check(res, {
    "create pr 201 or 409 or 404": (r) => r.status === 201 || r.status === 404 || r.status === 409
  });
}
