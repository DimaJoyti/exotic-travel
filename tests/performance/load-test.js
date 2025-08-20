import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 }, // Ramp up to 10 users
    { duration: '5m', target: 10 }, // Stay at 10 users
    { duration: '2m', target: 20 }, // Ramp up to 20 users
    { duration: '5m', target: 20 }, // Stay at 20 users
    { duration: '2m', target: 50 }, // Ramp up to 50 users
    { duration: '5m', target: 50 }, // Stay at 50 users
    { duration: '2m', target: 0 },  // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% of requests must complete below 2s
    http_req_failed: ['rate<0.05'],    // Error rate must be below 5%
    errors: ['rate<0.05'],             // Custom error rate must be below 5%
  },
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const FRONTEND_URL = __ENV.FRONTEND_URL || 'http://localhost:3000';

// Test data
const testUsers = [
  { email: 'test1@example.com', password: 'password123' },
  { email: 'test2@example.com', password: 'password123' },
  { email: 'test3@example.com', password: 'password123' },
];

let authToken = '';

export function setup() {
  // Login to get auth token for authenticated requests
  const loginResponse = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: testUsers[0].email,
    password: testUsers[0].password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (loginResponse.status === 200) {
    const loginData = JSON.parse(loginResponse.body);
    return { authToken: loginData.tokens?.access_token || '' };
  }

  console.warn('Failed to login during setup, some tests may fail');
  return { authToken: '' };
}

export default function (data) {
  const authToken = data.authToken;
  
  // Test scenarios with different weights
  const scenarios = [
    { name: 'health_check', weight: 10, func: testHealthCheck },
    { name: 'authentication', weight: 15, func: testAuthentication },
    { name: 'campaigns', weight: 25, func: testCampaigns },
    { name: 'content_generation', weight: 20, func: testContentGeneration },
    { name: 'integrations', weight: 15, func: testIntegrations },
    { name: 'analytics', weight: 10, func: testAnalytics },
    { name: 'frontend', weight: 5, func: testFrontend },
  ];

  // Select scenario based on weight
  const totalWeight = scenarios.reduce((sum, s) => sum + s.weight, 0);
  const random = Math.random() * totalWeight;
  let currentWeight = 0;
  
  for (const scenario of scenarios) {
    currentWeight += scenario.weight;
    if (random <= currentWeight) {
      scenario.func(authToken);
      break;
    }
  }

  sleep(1); // Wait 1 second between iterations
}

function testHealthCheck() {
  const response = http.get(`${BASE_URL}/health`);
  
  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 500ms': (r) => r.timings.duration < 500,
    'health check has correct status': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === 'healthy';
      } catch {
        return false;
      }
    },
  });

  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  requestCount.add(1);
}

function testAuthentication() {
  const user = testUsers[Math.floor(Math.random() * testUsers.length)];
  
  // Test login
  const loginResponse = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: user.email,
    password: user.password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  const loginSuccess = check(loginResponse, {
    'login status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'login response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!loginSuccess);
  responseTime.add(loginResponse.timings.duration);
  requestCount.add(1);

  // If login successful, test token refresh
  if (loginResponse.status === 200) {
    try {
      const loginData = JSON.parse(loginResponse.body);
      if (loginData.tokens?.refresh_token) {
        const refreshResponse = http.post(`${BASE_URL}/api/v1/auth/refresh`, JSON.stringify({
          refresh_token: loginData.tokens.refresh_token,
        }), {
          headers: { 'Content-Type': 'application/json' },
        });

        const refreshSuccess = check(refreshResponse, {
          'refresh status is 200': (r) => r.status === 200,
          'refresh response time < 1000ms': (r) => r.timings.duration < 1000,
        });

        errorRate.add(!refreshSuccess);
        responseTime.add(refreshResponse.timings.duration);
        requestCount.add(1);
      }
    } catch (e) {
      console.warn('Failed to parse login response:', e);
    }
  }
}

function testCampaigns(authToken) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Test get campaigns
  const getCampaignsResponse = http.get(`${BASE_URL}/api/v1/marketing/campaigns`, { headers });
  
  const getCampaignsSuccess = check(getCampaignsResponse, {
    'get campaigns status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'get campaigns response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!getCampaignsSuccess);
  responseTime.add(getCampaignsResponse.timings.duration);
  requestCount.add(1);

  // Test create campaign (if authenticated)
  if (getCampaignsResponse.status === 200) {
    const campaignData = {
      name: `Load Test Campaign ${Date.now()}`,
      type: 'social',
      budget: 1000,
      start_date: new Date().toISOString(),
      target_audience: {
        age_range: '25-45',
        interests: ['travel', 'adventure'],
      },
    };

    const createCampaignResponse = http.post(
      `${BASE_URL}/api/v1/marketing/campaigns`,
      JSON.stringify(campaignData),
      { headers }
    );

    const createCampaignSuccess = check(createCampaignResponse, {
      'create campaign status is 200 or 201': (r) => r.status === 200 || r.status === 201,
      'create campaign response time < 2000ms': (r) => r.timings.duration < 2000,
    });

    errorRate.add(!createCampaignSuccess);
    responseTime.add(createCampaignResponse.timings.duration);
    requestCount.add(1);
  }
}

function testContentGeneration(authToken) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  const contentRequest = {
    type: 'social_post',
    platform: 'facebook',
    topic: 'exotic travel destinations',
    tone: 'exciting',
    length: 'medium',
    target_audience: {
      age_range: '25-45',
      interests: ['travel', 'adventure'],
    },
  };

  const response = http.post(
    `${BASE_URL}/api/v1/content/generate`,
    JSON.stringify(contentRequest),
    { headers }
  );

  const success = check(response, {
    'content generation status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'content generation response time < 5000ms': (r) => r.timings.duration < 5000,
  });

  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  requestCount.add(1);
}

function testIntegrations(authToken) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Test get integrations
  const getIntegrationsResponse = http.get(`${BASE_URL}/api/v1/marketing/integrations`, { headers });
  
  const success = check(getIntegrationsResponse, {
    'get integrations status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'get integrations response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!success);
  responseTime.add(getIntegrationsResponse.timings.duration);
  requestCount.add(1);
}

function testAnalytics(authToken) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };

  // Test analytics dashboard
  const dashboardResponse = http.get(`${BASE_URL}/api/v1/analytics/dashboard`, { headers });
  
  const success = check(dashboardResponse, {
    'analytics dashboard status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'analytics dashboard response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  errorRate.add(!success);
  responseTime.add(dashboardResponse.timings.duration);
  requestCount.add(1);
}

function testFrontend() {
  // Test frontend health
  const frontendResponse = http.get(FRONTEND_URL);
  
  const success = check(frontendResponse, {
    'frontend status is 200': (r) => r.status === 200,
    'frontend response time < 3000ms': (r) => r.timings.duration < 3000,
    'frontend contains expected content': (r) => r.body.includes('Marketing AI') || r.body.includes('html'),
  });

  errorRate.add(!success);
  responseTime.add(frontendResponse.timings.duration);
  requestCount.add(1);
}

export function teardown(data) {
  // Cleanup if needed
  console.log('Load test completed');
}
