import http from "k6/http";
import { check, sleep } from "k6";
import { Trend } from "k6/metrics";

// Custom metrics
const customDuration = new Trend("custom_duration");

export const options = {
  stages: [
    { duration: "1m", target: 10 },   // Ramp up to 10 users
    { duration: "1m", target: 30 },   // Increase to 30 users
    { duration: "1m", target: 50 },   // Increase to 50 users - stress level
    { duration: "2m", target: 50 },   // Maintain stress level for 2 minutes
    { duration: "30s", target: 0 },   // Ramp down to 0
  ],
  thresholds: {
    http_req_duration: ["p(95)<2000"], // 95% of requests should be below 2s
    http_req_failed: ["rate<0.01"],    // Less than 1% of requests should fail
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:3000";

export default function () {
  const startTime = Date.now();

  // Test 1: Homepage
  let response = http.get(BASE_URL);
  check(response, {
    "homepage status is 200": (r) => r.status === 200,
    "homepage response time ok": (r) => r.timings.duration < 5000,
  });
  sleep(1);

  // Test 2: Get breeds list
  response = http.get(`${BASE_URL}/api/dogs/breeds`);
  check(response, {
    "breeds status is 200": (r) => r.status === 200,
    "breeds has data": (r) => {
      try {
        return Object.keys(JSON.parse(r.body).message).length > 0;
      } catch {
        return false;
      }
    },
  });
  sleep(1);

  // Test 3: Get random dog image
  response = http.get(`${BASE_URL}/api/dogs`);
  check(response, {
    "random dog status is 200": (r) => r.status === 200,
  });
  sleep(2);

  // Test 4: Get specific breed
  const breeds = ["husky", "corgi", "retriever", "bulldog", "poodle", "labrador"];
  const randomBreed = breeds[Math.floor(Math.random() * breeds.length)];
  response = http.get(`${BASE_URL}/api/dogs?breed=${randomBreed}`);
  check(response, {
    "specific breed status is 200": (r) => r.status === 200,
  });

  const endTime = Date.now();
  customDuration.add(endTime - startTime);
  
  sleep(1);
}
