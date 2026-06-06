import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiLatency = new Trend('api_latency');

// Test configuration
export const options = {
  stages: [
    // Ramp up
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '5m', target: 200 },   // Stay at 200 users
    { duration: '2m', target: 300 },   // Ramp up to 300 users
    { duration: '5m', target: 300 },   // Stay at 300 users
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],     // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],         // Less than 0.1% errors
    errors: ['rate<0.1'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8090';
const AUTH_TOKEN = __ENV.AUTH_TOKEN;

// Test data
const TEST_DATA = {
  collections: [],
  records: [],
};

export function setup() {
  // Create test collection
  const collectionRes = http.post(
    `${BASE_URL}/api/collections`,
    JSON.stringify({
      name: `load_test_collection_${Date.now()}`,
      schema: [
        { name: 'title', type: 'text', required: true },
        { name: 'description', type: 'text' },
        { name: 'price', type: 'number' },
        { name: 'count', type: 'number' },
      ],
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${AUTH_TOKEN}`,
      },
    }
  );

  const collection = JSON.parse(collectionRes.body);
  
  // Create initial records
  for (let i = 0; i < 1000; i++) {
    http.post(
      `${BASE_URL}/api/collections/${collection.id}/records`,
      JSON.stringify({
        title: `Product ${i}`,
        description: `Description for product ${i}`,
        price: Math.random() * 1000,
        count: Math.floor(Math.random() * 100),
      }),
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${AUTH_TOKEN}`,
        },
      }
    );
  }

  return { collectionId: collection.id };
}

export default function(data) {
  group('API Load Test', () => {
    const collectionId = data.collectionId;
    
    group('List Collections', () => {
      const res = http.get(`${BASE_URL}/api/collections`, {
        headers: { 'Authorization': `Bearer ${AUTH_TOKEN}` },
      });
      
      const success = check(res, {
        'list collections status is 200': (r) => r.status === 200,
        'list collections response time < 500ms': (r) => r.timings.duration < 500,
      });
      
      errorRate.add(!success);
      apiLatency.add(res.timings.duration);
    });

    group('Create Record', () => {
      const payload = JSON.stringify({
        title: `Test Product ${Math.random()}`,
        description: 'Load test record',
        price: Math.random() * 1000,
        count: Math.floor(Math.random() * 100),
      });

      const res = http.post(
        `${BASE_URL}/api/collections/${collectionId}/records`,
        payload,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${AUTH_TOKEN}`,
          },
        }
      );

      const success = check(res, {
        'create record status is 201': (r) => r.status === 201,
        'create record response time < 500ms': (r) => r.timings.duration < 500,
      });

      errorRate.add(!success);
      apiLatency.add(res.timings.duration);
    });

    group('List Records', () => {
      const page = Math.floor(Math.random() * 10) + 1;
      const res = http.get(
        `${BASE_URL}/api/collections/${collectionId}/records?page=${page}&perPage=30`,
        {
          headers: { 'Authorization': `Bearer ${AUTH_TOKEN}` },
        }
      );

      const success = check(res, {
        'list records status is 200': (r) => r.status === 200,
        'list records response time < 500ms': (r) => r.timings.duration < 500,
      });

      errorRate.add(!success);
      apiLatency.add(res.timings.duration);
    });

    group('Get Record', () => {
      const recordId = `record_${Math.floor(Math.random() * 1000)}`;
      const res = http.get(
        `${BASE_URL}/api/collections/${collectionId}/records/${recordId}`,
        {
          headers: { 'Authorization': `Bearer ${AUTH_TOKEN}` },
        }
      );

      // 404 is OK for random IDs, just check performance
      const success = check(res, {
        'get record response time < 200ms': (r) => r.timings.duration < 200,
      });

      errorRate.add(!success);
      apiLatency.add(res.timings.duration);
    });
  });

  sleep(1);
}

export function teardown(data) {
  // Cleanup: Delete test collection
  http.del(
    `${BASE_URL}/api/collections/${data.collectionId}`,
    null,
    {
      headers: { 'Authorization': `Bearer ${AUTH_TOKEN}` },
    }
  );
}
