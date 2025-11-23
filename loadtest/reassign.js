import http from "k6/http";
import { check } from "k6";

export let options = {
  vus: 20,
  duration: "30s",
  rate: 5,
  thresholds: {
    http_req_failed: ["rate<0.001"],
    http_req_duration: ["p(95)<300"]
  }
};

const BASE = __ENV.BASE_URL || "http://localhost:8080";

const PR_IDS = ["pr002", "pr003", "pr004", "pr006", "pr008"];

export default function () {
  const prId = PR_IDS[__ITER % PR_IDS.length];
  const payload = JSON.stringify({ pull_request_id: prId, old_user_id: "u002" });

  let res = http.post(`${BASE}/pullRequest/reassign`, payload, { headers: { "Content-Type": "application/json" } });
  check(res, {
    "reassign ok or 404 or 409": (r) => r.status === 200 || r.status === 404 || r.status === 409
  });
}
