import type { Component } from 'solid-js';
import { createSignal, onMount, onCleanup } from 'solid-js';
import styles from '../App.module.css';
import { DefaultTrackMux, MOQ, TrackWriter } from '@okutanidaichi/moqt';
import { background } from '@okutanidaichi/moqt/internal';

const Home: Component = () => {
  // Publish state
  const [isPublishConnected, setIsPublishConnected] = createSignal(false);
  const [publishCount, setPublishCount] = createSignal(0);
  const [publishConnectionStatus, setPublishConnectionStatus] = createSignal('Disconnected');
  const [publishMessage, setPublishMessage] = createSignal('Hello from interop web client');

  // Subscribe state
  const [isSubscribeConnected, setIsSubscribeConnected] = createSignal(false);
  const [messages, setMessages] = createSignal<string[]>([]);
  const [subscribeConnectionStatus, setSubscribeConnectionStatus] = createSignal('Connecting...');
  let intervalId: NodeJS.Timeout | undefined;

  // Publish functionality
  const startPublish = async () => {
    try {
      setPublishConnectionStatus('Connecting...');
      
      if (!('WebTransport' in window)) {
        throw new Error('WebTransport is not supported in this browser');
      }

      DefaultTrackMux.handleTrack(background(), "/interop.client", {serveTrack: async (trackWriter: TrackWriter)=>{
        const encoder = new TextEncoder()
        let sequence = 1n
        for (let i = 0; i < 10; i++) {
          const [group, err] = await trackWriter.openGroup(sequence)
          if (err || !group) {
            console.log("Failed to open group")
            trackWriter.closeWithError(0, "unexpected error")
            return
          }

          const err2 = await group.writeFrame(encoder.encode("Hello from interop web client!"))
          if (err2) {
            console.log("Failed to write frame")
            trackWriter.closeWithError(0, "unexpected error")
            return
          }

          group.close()

          sequence++
          await new Promise(resolve => setTimeout(resolve, 1000))
        }
        trackWriter.close()
      }})

      const moq = new MOQ()
      const session = await moq.dial("https://moqt.example.com:9000/publish")
      
      setIsPublishConnected(true);
      setPublishConnectionStatus('Connected (WebTransport ready)');
      
      const interval = setInterval(() => {
        const count = publishCount() + 1;
        setPublishCount(count);
        const message = `Publishing frame ${count}: ${publishMessage()}`;
        console.log(message);
      }, 2000);
      
      onCleanup(() => {
        clearInterval(interval);
      });
      
    } catch (error) {
      console.error('Publish error:', error);
      setPublishConnectionStatus(`Error: ${error}`);
    }
  };

  const stopPublish = () => {
    setIsPublishConnected(false);
    setPublishConnectionStatus('Disconnected');
    setPublishCount(0);
  };

  // Subscribe functionality
  const startSubscribe = async () => {
    try {
      setSubscribeConnectionStatus('Connecting...');

      if (!('WebTransport' in window)) {
        throw new Error('WebTransport is not supported in this browser');
      }

      const moq = new MOQ();
      const session = await moq.dial("https://moqt.example.com:9000/subscribe")

      await session.ready;

      setIsSubscribeConnected(true);
      setSubscribeConnectionStatus('Connected (MOQT library loaded)');

      const annstr = await session.openAnnounceStream("/");

      for (;;) {
        const announcement = await annstr.receive()
        if (!announcement.isActive()) {
            continue
        }

        console.log("Announcement received:", announcement)

        const trackReader = await session.openTrackStream(announcement.broadcastPath, "")
        for (;;) {
            const [group, err] = await trackReader.acceptGroup()
            if (err || !group) {
                continue;
            }

            let frameSequence = 0;
            for (;;) {
                const [frame, err] = await group.readFrame();
                if (err || !frame) {
                    break;
                }

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

      setSubscribeConnectionStatus(`Error: ${errorMessage}`);
      setIsSubscribeConnected(false);
    }
  };

  const stopSubscribe = () => {
    setIsSubscribeConnected(false);
    setSubscribeConnectionStatus('Disconnected');
    setMessages([]);

    if (intervalId) {
      clearInterval(intervalId);
      intervalId = undefined;
    }
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
        <h1>MOQT Web Client</h1>
        <p>MOQT (Media over QUIC Transport) client implementation using WebTransport</p>
        
        <div style="display: flex; gap: 40px; margin-top: 40px; max-width: 1200px; margin-left: auto; margin-right: auto;">
          {/* Publish Section */}
          <div style="flex: 1; border: 1px solid #30363d; border-radius: 8px; padding: 20px; background: #0d1117;">
            <h2 style="color: #61dafb; margin-top: 0;">📤 Publish</h2>
            
            <div style="margin: 20px 0;">
              <p>Status: <span style={`color: ${isPublishConnected() ? 'green' : 'red'}`}>{publishConnectionStatus()}</span></p>
              {isPublishConnected() && (
                <p>Published frames: <span style="color: #61dafb;">{publishCount()}</span></p>
              )}
            </div>

            <div style="margin: 20px 0;">
              <label style="display: block; margin-bottom: 10px; color: #e6edf3;">
                Message to publish:
                <br />
                <input 
                  type="text" 
                  value={publishMessage()} 
                  onInput={(e) => setPublishMessage(e.currentTarget.value)}
                  disabled={isPublishConnected()}
                  style="margin-top: 5px; padding: 8px; width: 100%; font-size: 14px; border: 1px solid #30363d; border-radius: 4px; background: #161b22; color: #e6edf3;"
                />
              </label>
            </div>

            <div style="margin: 20px 0;">
              {!isPublishConnected() ? (
                <button 
                  onClick={startPublish}
                  style="padding: 10px 20px; font-size: 16px; background: #61dafb; border: none; border-radius: 5px; cursor: pointer; color: #0d1117; font-weight: bold;"
                >
                  Start Publish
                </button>
              ) : (
                <button 
                  onClick={stopPublish}
                  style="padding: 10px 20px; font-size: 16px; background: #f56565; border: none; border-radius: 5px; cursor: pointer; color: white; font-weight: bold;"
                >
                  Stop Publish
                </button>
              )}
            </div>

            <div style="margin-top: 20px;">
              <h3 style="color: #e6edf3; margin-bottom: 10px;">Publish Log:</h3>
              <div style="background: #1e1e1e; padding: 10px; border-radius: 5px; font-family: monospace; height: 120px; overflow-y: auto; border: 1px solid #30363d;">
                {publishCount() === 0 ? (
                  <p style="color: #888; margin: 0;">No frames published yet...</p>
                ) : (
                  <div style="color: #61dafb; margin: 0;">
                    Publishing "{publishMessage()}" every 2 seconds...
                    <br />
                    Total frames sent: {publishCount()}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Subscribe Section */}
          <div style="flex: 1; border: 1px solid #30363d; border-radius: 8px; padding: 20px; background: #0d1117;">
            <h2 style="color: #61dafb; margin-top: 0;">📥 Subscribe</h2>

            <div style="margin: 20px 0;">
              <p>Status: <span style={`color: ${isSubscribeConnected() ? 'green' : 'red'}`}>{subscribeConnectionStatus()}</span></p>
            </div>

            <div style="margin: 20px 0;">
              <button
                onClick={stopSubscribe}
                style="padding: 10px 20px; font-size: 16px; background: #f56565; border: none; border-radius: 5px; cursor: pointer; color: white; font-weight: bold;"
                disabled={!isSubscribeConnected()}
              >
                Stop Subscribe
              </button>
            </div>

            <div style="margin-top: 20px;">
              <h3 style="color: #e6edf3; margin-bottom: 10px;">Received Messages:</h3>
              <div style="background: #1e1e1e; padding: 10px; border-radius: 5px; font-family: monospace; height: 120px; overflow-y: auto; border: 1px solid #30363d;">
                {messages().length === 0 ? (
                  <p style="color: #888; margin: 0;">No messages received yet...</p>
                ) : (
                  messages().map((message, index) => (
                    <div style="margin: 2px 0; color: #61dafb; font-size: 14px; line-height: 1.2; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">
                      {message}
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </header>
    </div>
  );
};

export default Home;
