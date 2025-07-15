import type { Component } from 'solid-js';
import { createSignal, onMount, onCleanup } from 'solid-js';
import { A } from '@solidjs/router';
import styles from '../App.module.css';
import { MOQ, Session } from '@okutanidaichi/moqt';

const Subscribe: Component = () => {
  const [isConnected, setIsConnected] = createSignal(false);
  const [messages, setMessages] = createSignal<string[]>([]);
  const [connectionStatus, setConnectionStatus] = createSignal('Connecting...');
  let intervalId: NodeJS.Timeout | undefined;

  // MOQT client implementation using the @okutanidaichi/moqt library
  const startSubscribe = async () => {
    try {
      setConnectionStatus('Connecting...');

      // Test WebTransport API availability
      if (!('WebTransport' in window)) {
        throw new Error('WebTransport is not supported in this browser');
      }

      // Create MOQT client using the library
      const moq = new MOQ();

      const session = await moq.dial("https://moqt.example.com:9000/subscribe")

      // Wait for session to be ready
      await session.ready;

      // For now, just test the library import
      setIsConnected(true);
      setConnectionStatus('Connected (MOQT library loaded)');

      const annstr = await session.openAnnounceStream("/");

      for (;;) {
        const announcement = await annstr.receive()
        if (!announcement.isActive()) {
            continue
        }

        console.log("Announcement received:", announcement)

        const subscription = await session.openTrackStream(announcement.broadcastPath, "")
        for (;;) {
            const [group, err] = await subscription.trackReader.acceptGroup()
            if (err || !group) {
                continue;
            }

            let frameSequence = 0;
            for (;;) {
                const [frame, err] = await group.readFrame();
                if (err || !frame) {
                    break;
                }

                // Display frame data with Group and Frame sequence
                const frameText = new TextDecoder().decode(frame);
                const timestamp = new Date().toLocaleTimeString();
                const message = `[${timestamp}] Group: ${group.groupSequence}, Frame: ${frameSequence} - ${frameText}`;

                setMessages([message]);
                frameSequence++;
            }
        }
      }

    } catch (error) {
      console.error('Subscribe error:', error);
      let errorMessage = 'Unknown error';

      if (error instanceof Error) {
        errorMessage = error.message;
        console.error('Error details:', {
          name: error.name,
          message: error.message,
          stack: error.stack,
          cause: error.cause
        });
      }

      setConnectionStatus(`Error: ${errorMessage}`);
      setIsConnected(false);
    }
  };

  const stopSubscribe = () => {
    setIsConnected(false);
    setConnectionStatus('Disconnected');
    setMessages([]);

    if (intervalId) {
      clearInterval(intervalId);
      intervalId = undefined;
    }

    // TODO: Clean up WebTransport connection
    // - Close streams
    // - End session
  };

  // Auto-start subscription when component mounts
  onMount(() => {
    startSubscribe();
  });

  // Cleanup on component unmount
  onCleanup(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });

  return (
    <div class={styles.App}>
      <header class={styles.header}>
        <h1>MOQT Subscribe</h1>
        <A href="/" class={styles.link} style="margin-bottom: 20px; display: inline-block;">
          ← Back to Home
        </A>

        <div style="margin: 20px 0;">
          <p>Status: <span style={`color: ${isConnected() ? 'green' : 'red'}`}>{connectionStatus()}</span></p>
        </div>

        <div style="margin: 20px 0;">
          <button
            onClick={stopSubscribe}
            style="padding: 10px 20px; font-size: 16px; background: #f56565; border: none; border-radius: 5px; cursor: pointer; color: white;"
            disabled={!isConnected()}
          >
            Stop Subscribe
          </button>
        </div>

        <div style="margin-top: 30px; height: 150px; width: 600px; margin-left: auto; margin-right: auto; display: flex; flex-direction: column;">
          <h3 style="margin-bottom: 15px; color: #61dafb;">Received Messages:</h3>
          <div style="background: #1e1e1e; padding: 15px; border-radius: 8px; font-family: 'Consolas', 'Monaco', 'Courier New', monospace; border: 1px solid #333; flex: 1; overflow: hidden;">
            {messages().length === 0 ? (
              <p style="color: #888; margin: 8px 0; font-size: 17px; text-align: left; line-height: 1.2;">No messages received yet...</p>
            ) : (
              messages().map((message, index) => (
                <div style="margin: 2px 0; color: #61dafb; font-size: 17px; line-height: 1.2; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; text-align: left;">
                  {message}
                </div>
              ))
            )}
          </div>
        </div>
      </header>
    </div>
  );
};

export default Subscribe;
