import React from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';

const container = document.getElementById('root');
if (container) {
  let root = (container as any)._reactRoot;
  if (!root) {
    root = createRoot(container);
    (container as any)._reactRoot = root;
  }
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}
