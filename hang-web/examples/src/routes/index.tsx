import { A } from "@solidjs/router";
import { createSignal, onMount } from "solid-js";
import { clientOnly } from "@solidjs/start";

const HangRoom = clientOnly(() => import("../components/HangRoom"));

export default function Home() {
  return (
    <main class="text-center mx-auto text-gray-700 p-4">
      <h1 class="max-6-xs text-6xl text-sky-700 font-thin uppercase my-16">Room Demo</h1>
      <div class="mb-8">
        <HangRoom roomId="demo-room" localName="User1" fallback={
          <div class="hang-room-container" data-fallback></div>
        }/>
      </div>
      <p class="mt-8">
        Visit{" "}
        <a href="https://solidjs.com" target="_blank" class="text-sky-600 hover:underline">
          solidjs.com
        </a>{" "}
        to learn how to build Solid apps.
      </p>
      <p class="my-4">
        <span>Home</span>
        {" - "}
        <a href="/about" class="text-sky-600 hover:underline">
          About Page
        </a>{" "}
      </p>
    </main>
  );
}
