// load_test.js - CORRECTED
import http from 'k6/http';
import { check } from 'k6';

// ✅ Options must be at TOP LEVEL (not inside functions)
export const options = {
  stages: [
    { duration: '5s', target: 20 },
    { duration: '10s', target: 20 },
  ],
  thresholds: {
    // ✅ 50 = 50 milliseconds for p95 latency
    http_req_duration: ['p(95)<50'],
    
    // ✅ Allow up to 99.9% "failures" since 429s are expected
    http_req_failed: ['rate<0.999'],
  },
};

// ✅ Default function must NOT contain exports
export default function () {
  const params = { 
    headers: { 'X-User-ID': 'test-user-001' } 
  };
  
  const res = http.get('http://localhost:8080/request', params);
  
  // ✅ Checks go here (not exports)
  check(res, {
    'valid response': (r) => r.status === 200 || r.status === 429,
    'no server errors': (r) => r.status < 500,  // Only 5xx are real failures
  });
}