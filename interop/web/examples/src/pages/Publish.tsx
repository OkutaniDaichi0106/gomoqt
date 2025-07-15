import type { Component } from 'solid-js';
import { createSignal, onCleanup } from 'solid-js';
import { A } from '@solidjs/router';
import styles from '../App.module.css';

const Publish: Component = () => {
  const [isConnected, setIsConnected] = createSignal(false);
  const [publishCount, setPublishCount] = createSignal(0);
  const [connectionStatus, setConnectionStatus] = createSignal('Disconnected');
  const [publishMessage, setPublishMessage] = createSignal('Hello from web client');

  // MOQT client implementation using WebTransport API directly  
  // This function will be updated to use the MOQT library once import issues are resolved
  const startPublish = async () => {
    try {
      setConnectionStatus('Connecting...');
      
      // For now, test WebTransport API availability
      if (!('WebTransport' in window)) {
        throw new Error('WebTransport is not supported in this browser');
      }
      
      console.log('WebTransport API is available');
      setIsConnected(true);
      setConnectionStatus('Connected (WebTransport ready)');
      
      // TODO: Once @okutanidaichi/moqt import is fixed:
      // const client = new Client();
      // const session = await client.dial("https://localhost:4469/publish");
      // Set up TrackMux and publication...
      
      // Simulate publishing for testing
      const interval = setInterval(() => {
        const count = publishCount() + 1;
        setPublishCount(count);
        const message = `Publishing frame ${count}: ${publishMessage()}`;
        console.log(message);
        // Log message (no UI display needed for publish)
      }, 2000);
      
      onCleanup(() => {
        clearInterval(interval);
        // TODO: session.close();
      });
      
    } catch (error) {
      console.error('Publish error:', error);
      setConnectionStatus(`Error: ${error}`);
    }
  };

  const stopPublish = () => {
    setIsConnected(false);
    setConnectionStatus('Disconnected');
    setPublishCount(0);
    
    // TODO: Clean up WebTransport connection
    // - Close TrackWriter
    // - End session
    // - Stop transmission loop
  };

  return (
    <div class={styles.App}>
      <header class={styles.header}>
        <h1>MOQT Publish</h1>
        <A href="/" class={styles.link} style="margin-bottom: 20px; display: inline-block;">
          ← Back to Home
        </A>
        
        <div style="margin: 20px 0;">
          <p>Status: <span style={`color: ${isConnected() ? 'green' : 'red'}`}>{connectionStatus()}</span></p>
          {isConnected() && (
            <p>Published frames: <span style="color: #61dafb;">{publishCount()}</span></p>
          )}
        </div>

        <div style="margin: 20px 0;">
          <label style="display: block; margin-bottom: 10px;">
            Message to publish:
            <br />
            <input 
              type="text" 
              value={publishMessage()} 
              onInput={(e) => setPublishMessage(e.currentTarget.value)}
              disabled={isConnected()}
              style="margin-top: 5px; padding: 8px; width: 300px; font-size: 14px; border: 1px solid #ccc; border-radius: 4px;"
            />
          </label>
        </div>

        <div style="margin: 20px 0;">
          {!isConnected() ? (
            <button 
              onClick={startPublish}
              style="padding: 10px 20px; font-size: 16px; background: #61dafb; border: none; border-radius: 5px; cursor: pointer;"
            >
              Start Publish
            </button>
          ) : (
            <button 
              onClick={stopPublish}
              style="padding: 10px 20px; font-size: 16px; background: #f56565; border: none; border-radius: 5px; cursor: pointer; color: white;"
            >
              Stop Publish
            </button>
          )}
        </div>

        <div style="margin-top: 30px; text-align: left;">
          <h3>Publish Log:</h3>
          <div style="background: #1e1e1e; padding: 10px; border-radius: 5px; font-family: monospace; max-height: 200px; overflow-y: auto;">
            {publishCount() === 0 ? (
              <p style="color: #888;">No frames published yet...</p>
            ) : (
              <div style="color: #61dafb;">
                Publishing "{publishMessage()}" every 2 seconds...
                <br />
                Total frames sent: {publishCount()}
              </div>
            )}
          </div>
        </div>
      </header>
    </div>
  );
};

export default Publish;
