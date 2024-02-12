import React, { useState, useEffect } from 'react';
import './App.css';

function App() {
  const [time, setTime] = useState(new Date());

  useEffect(() => {
    const timer = setInterval(() => {
      // Get the current UTC time
      const utcTime = new Date().toUTCString();
      setTime(new Date(utcTime));
    }, 1000);

    // Clean up the interval on component unmount
    return () => clearInterval(timer);
  }, []);

  return (
    <div className="App">
      <header className="App-header">
        <p>Current Time in UTC</p>
        <h2>{time.toLocaleTimeString('en-GB', { timeZone: 'UTC' })}</h2>
      </header>
    </div>
  );
}

export default App;
