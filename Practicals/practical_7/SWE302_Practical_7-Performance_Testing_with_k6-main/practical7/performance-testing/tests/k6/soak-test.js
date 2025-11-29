import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Counter, Trend } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const successfulRequests = new Counter("successful_requests");
const memoryTrend = new Trend("memory_indicator");

export const options = {
  stages: [
    { duration: "2m", target: 15 },   // Ramp up to 15 users over 2 minutes
    { duration: "26m", target: 15 },  // Stay at 15 users for 26 minutes (total 30 min)
    { duration: "2m", target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ["p(95)<1000"], // 95% of requests should be below 1s
    http_req_failed: ["rate<0.02"],    // Less than 2% of requests should fail
    errors: ["rate<0.05"],             // Error rate should be below 5%
    http_req_waiting: ["p(95)<800"],   // Server response time should be consistent
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:3000";

export default function () {
  // Test 1: Homepage
  let response = http.get(BASE_URL);
  if (check(response, {
    "homepage status is 200": (r) => r.status === 200,
    "homepage loads in time": (r) => r.timings.duration < 2000,
  })) {
    successfulRequests.add(1);
  } else {
    errorRate.add(1);
  }
  
  // Track response time as memory indicator
  memoryTrend.add(response.timings.duration);
  sleep(2);

  // Test 2: Get breeds list
  response = http.get(`${BASE_URL}/api/dogs/breeds`);
  if (check(response, {
    "breeds status is 200": (r) => r.status === 200,
    "breeds list is not empty": (r) => {
      try {
        return Object.keys(JSON.parse(r.body).message).length > 0;
      } catch (e) {
        return false;
      }
    },
  })) {
    successfulRequests.add(1);
  } else {
    errorRate.add(1);
  }
  sleep(1);

  // Test 3: Get random dog image
  response = http.get(`${BASE_URL}/api/dogs`);
  if (check(response, {
    "random dog status is 200": (r) => r.status === 200,
    "random dog has message": (r) => {
      try {
        return JSON.parse(r.body).message !== undefined;
      } catch (e) {
        return false;
      }
    },
  })) {
    successfulRequests.add(1);
  } else {
    errorRate.add(1);
  }
  sleep(3);

  // Test 4: Get specific breed
  const breeds = ["husky", "corgi", "retriever", "bulldog", "poodle"];
  const randomBreed = breeds[Math.floor(Math.random() * breeds.length)];
  response = http.get(`${BASE_URL}/api/dogs?breed=${randomBreed}`);
  if (check(response, {
    "specific breed status is 200": (r) => r.status === 200,
    "specific breed has message": (r) => {
      try {
        return JSON.parse(r.body).message !== undefined;
      } catch (e) {
        return false;
      }
    },
  })) {
    successfulRequests.add(1);
  } else {
    errorRate.add(1);
  }
  sleep(2);
}
