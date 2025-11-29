import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Counter } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const totalRequests = new Counter("total_requests");

export const options = {
  stages: [
    { duration: "10s", target: 10 },   // Ramp up to 10 users (normal load)
    { duration: "1m", target: 50 },    // Spike to 50 users - Reduced for laptop capability
    { duration: "10s", target: 10 },   // Drop back to 10 users
    { duration: "10s", target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<3000"], // 95% of requests should be below 3s during spike - More realistic
    http_req_failed: ["rate<0.05"],    // Less than 5% of requests should fail
    errors: ["rate<0.1"],              // Error rate should be below 10%
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:3000";

export default function () {
  // Test 1: Homepage
  let response = http.get(BASE_URL);
  totalRequests.add(1);
  check(response, {
    "homepage status is 200": (r) => r.status === 200,
  }) || errorRate.add(1);
  sleep(1);

  // Test 2: Get random dog
  response = http.get(`${BASE_URL}/api/dogs`);
  totalRequests.add(1);
  check(response, {
    "random dog status is 200": (r) => r.status === 200,
  }) || errorRate.add(1);
  sleep(1);

  // Test 3: Get breeds
  response = http.get(`${BASE_URL}/api/dogs/breeds`);
  totalRequests.add(1);
  check(response, {
    "breeds status is 200": (r) => r.status === 200,
  }) || errorRate.add(1);
  sleep(1);
}
