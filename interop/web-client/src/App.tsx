import type { Component } from 'solid-js';
import { Router, Route } from '@solidjs/router';

import Home from './pages/Home';

const App: Component = () => {
  return (
    <Router>
      <Route path="/" component={Home} />
    </Router>
  );
};

export default App;
