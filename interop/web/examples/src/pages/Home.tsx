import type { Component } from 'solid-js';
import { A } from '@solidjs/router';
import styles from '../App.module.css';

const Home: Component = () => {
  return (
    <div class={styles.App}>
      <header class={styles.header}>
        <h1>MOQT Web Client</h1>
        <p>
          MOQT (Media over QUIC Transport) client implementation using WebTransport
        </p>
        <div style="display: flex; flex-direction: column; align-items: center; gap: 8px; margin-top: 40px; max-width: 320px; margin-left: auto; margin-right: auto;">
          <A 
            href="/subscribe" 
            class={styles.link}
            style="
              display: block;
              width: 100%;
              padding: 8px 12px;
              background: #0d1117;
              border: 1px solid #30363d;
              border-radius: 6px;
              text-decoration: none;
              font-family: 'SFMono-Regular', 'Consolas', 'Liberation Mono', 'Menlo', monospace;
              font-size: 14px;
              color: #e6edf3;
              font-weight: 400;
              transition: background-color 0.2s;
            "
          >
            /subscribe
          </A>
          <A 
            href="/publish" 
            class={styles.link}
            style="
              display: block;
              width: 100%;
              padding: 8px 12px;
              background: #0d1117;
              border: 1px solid #30363d;
              border-radius: 6px;
              text-decoration: none;
              font-family: 'SFMono-Regular', 'Consolas', 'Liberation Mono', 'Menlo', monospace;
              font-size: 14px;
              color: #e6edf3;
              font-weight: 400;
              transition: background-color 0.2s;
            "
          >
            /publish
          </A>
          <style>{`
            a:hover {
              background-color: #161b22 !important;
            }
          `}</style>
        </div>
      </header>
    </div>
  );
};

export default Home;
