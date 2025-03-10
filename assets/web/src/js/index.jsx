import React from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './ui/App';

import '../scss/main.scss';

// Normally, the 'varnishmon' variable (defined in the 'index.html' template) is
// hydrated server-side by the varnishmon service to avoid a round-trip during
// the initial page load. However, when using Vite's development server, the
// template is rendered directly by Vite using mock data, and then the
// 'varnishmon' variable is populated client-side using an API call.
async function initConfig() {
  if (varnishmon == null) {
    return fetch('/config')
      .then((response) => response.json())
      .then((data) => {
        varnishmon = data;
      })
      .catch((error) => {
        console.error('Failed to fetch configuration!', error);
      });
  }
  return Promise.resolve();
}

initConfig().then(() => {
  createRoot(document.getElementById('root')).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  );
});
