import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 10,
  rate: 5,
  duration: "10s"
};

const BASE = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  let r1 = http.get(`${BASE}/team/get?team_name=Frontend%20Team`);
  check(r1, {"team get status is 200": (r) => r.status === 200});

  let r2 = http.get(`${BASE}/users/getReview?user_id=u001`);
  check(r2, {"getReview status is 200 or 404": (r) => r.status === 200 || r.status === 404});

  sleep(1);
}
