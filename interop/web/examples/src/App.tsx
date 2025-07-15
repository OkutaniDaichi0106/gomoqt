import type { Component } from 'solid-js';
import { Router, Route } from '@solidjs/router';

import Home from './pages/Home';
import Subscribe from './pages/Subscribe';
import Publish from './pages/Publish';

const App: Component = () => {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/subscribe" component={Subscribe} />
      <Route path="/publish" component={Publish} />
    </Router>
  );
};

export default App;
