import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/layout/Layout';
import StandardsPage from './pages/StandardsPage';
import OperationsPage from './pages/OperationsPage';
import ResultsPage from './pages/ResultsPage';
import OperationSummary from './components/operations/OperationSummary';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<StandardsPage />} />
          <Route path="/operations" element={<OperationsPage />} />
          <Route path="/operations/:id/summary" element={<OperationSummary />} />
          <Route path="/results" element={<ResultsPage />} />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App;
