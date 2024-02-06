import http from 'k6/http';
import {check, sleep} from 'k6';

const base_url = 'http://app:3000'

export const options = {
  // Key configurations for breakpoint in this section
  executor: 'ramping-arrival-rate', //Assure load increase if the system slows
  stages: [
    { duration: '1m', target: 20000 }, // just slowly ramp-up to a HUGE load
    { duration: '1m', target: 20000 }, // maintain VUs for another 1m
  ],
  thresholds: {
    http_req_failed: [{ threshold: 'rate<0.01', abortOnFail: true, delayAbortEval: "10s" }], // http errors should be less than 1%
    http_req_duration: [{ threshold: 'p(95)<1000', abortOnFail: true, delayAbortEval: "10s" }], // 95% of requests should be below 1s
  },
};

export default () => {
  const res = http.get(base_url + '/api');

  check(res, { 'status was 200': (r) => r.status == 200 });
};

