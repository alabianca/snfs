import React from 'react';
import logo from './logo.svg';
import './App.css';
import '../node_modules/bulma/css/bulma.css'

import Layout from './components/layout/Layout';
import Sidebar from './components/sidebar/sidebar';
import Content from './components/content/content';

function App() {
  return (
    <div className="App">
      <Layout>
        <Sidebar/>
        <Content/>
      </Layout>
    </div>
  );
}

export default App;
