import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

// Custom metric to track error rate
const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "2m", target: 20 }, // Ramp up to 20 users over 2 minutes
    { duration: "5m", target: 20 }, // Stay at 20 users for 5 minutes
    { duration: "2m", target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    http_req_failed: ["rate<0.01"],   // Less than 1% of requests should fail
    errors: ["rate<0.05"],             // Error rate should be below 5%
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:3000";

export default function () {
  // Test 1: Homepage
  let response = http.get(BASE_URL);
  check(response, {
    "homepage status is 200": (r) => r.status === 200,
    "homepage loads quickly": (r) => r.timings.duration < 2000,
  }) || errorRate.add(1);
  sleep(2);

  // Test 2: Get breeds list
  response = http.get(`${BASE_URL}/api/dogs/breeds`);
  check(response, {
    "breeds status is 200": (r) => r.status === 200,
    "breeds list is not empty": (r) => {
      try {
        return Object.keys(JSON.parse(r.body).message).length > 0;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);
  sleep(1);

  // Test 3: Get random dog image
  response = http.get(`${BASE_URL}/api/dogs`);
  check(response, {
    "random dog status is 200": (r) => r.status === 200,
    "random dog has message": (r) => {
      try {
        return JSON.parse(r.body).message !== undefined;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);
  sleep(3);

  // Test 4: Get specific breed
  const breeds = ["husky", "corgi", "retriever", "bulldog", "poodle"];
  const randomBreed = breeds[Math.floor(Math.random() * breeds.length)];
  response = http.get(`${BASE_URL}/api/dogs?breed=${randomBreed}`);
  check(response, {
    "specific breed status is 200": (r) => r.status === 200,
    "specific breed has message": (r) => {
      try {
        return JSON.parse(r.body).message !== undefined;
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);
  sleep(2);
}
