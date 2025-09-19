import { createSignal, createEffect, createResource } from 'solid-js';
import { Session } from '@okutanidaichi/moqt';
import { defineRoom } from '@okutanidaichi/hang';
import type { RoomElement,JoinedMember } from '@okutanidaichi/hang';

defineRoom();

// Local re-declare (runtime shape) to avoid depending on library export yet
interface RoomLifecycleStatus { type: 'idle' | 'connecting' | 'success' | 'error' | 'left' | 'disconnected'; message: string }

interface HangRoomProps {
  roomId: string;
  localName: string;
}

export default function HangRoom(props: HangRoomProps) {
  let roomEl: RoomElement | undefined;
  const [isHover, setIsHover] = createSignal(false);
  const [roomId, setRoomId] = createSignal(props.roomId);
  const [localName, setLocalName] = createSignal(props.localName);
  const [joined, setJoined] = createSignal(false);
  const [session] = createResource(async () => {
    try {
      const moqUrl = import.meta.env.VITE_MOQ_URL;
      const transport = new WebTransport(moqUrl);
      console.log('Initializing session...');
      const sess = new Session(transport);
      console.log('Waiting for session to be ready...');
      await sess.ready;
      console.log('Session initialized:', moqUrl);
      return sess;
    } catch (e) {
      console.error('Failed to initialize session:', e);
      throw e;
    }
  });

  const handleJoin = async () => {
    if (joined() || session.loading || !session()) return;
    try {
      if (roomEl && session()) {
        roomEl.setAttribute('room-id', roomId());
        roomEl.setAttribute('local-name', localName());
        roomEl.join(session()!);
      }
      setJoined(true);
    } catch (e) {
      console.error(e);
      // Optionally handle join error (e.g., show alert or reset state)
    }
  };

  const handleLeave = () => {
    if (!joined()) return;
    roomEl?.leave();
    setJoined(false);
  };

  createEffect(() => {
    if (!roomEl) return;
    if (!joined()) {
      roomEl.setAttribute('room-id', roomId());
      roomEl.setAttribute('local-name', localName());
    }
  });

  return (
    <div class='hang-room-container'>
      <hang-room ref={(el: any) => (roomEl = el as RoomElement)} attr:room-id={roomId()} attr:local-name={localName()} attr:description="Demo room" />
      <div
        class={`absolute inset-0 flex flex-col items-stretch justify-center p-4 bg-white/92 transition-opacity duration-300 z-10 ${
          joined() ? 'opacity-0 pointer-events-none' : 'opacity-100 pointer-events-auto'
        }`}
        onMouseEnter={() => setIsHover(true)}
        onMouseLeave={() => setIsHover(false)}
      >
        <div class='mb-4'>
          <label class='block text-sm font-medium text-gray-700'>Room ID:</label>
          <input
            type='text'
            value={roomId()}
            disabled={joined() || session.loading || !session()}
            onInput={(e) => setRoomId(e.target.value)}
            class='mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 disabled:bg-gray-100'
          />
        </div>
        <div class='mb-4'>
          <label class='block text-sm font-medium text-gray-700'>Local Name:</label>
          <input
            type='text'
            value={localName()}
            disabled={joined() || session.loading || !session()}
            onInput={(e) => setLocalName(e.target.value)}
            class='mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 disabled:bg-gray-100'
          />
        </div>
        <button
          onClick={handleJoin}
          disabled={joined() || session.loading || !session()}
          class='mx-auto mb-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-400'
        >
          {joined() ? 'Joined' : 'Join Room'}
        </button>
        <p class='text-center text-sm text-gray-600 mt-1'>
          {session.loading ? 'Connecting to server...' : session.error ? 'Failed to connect' : joined() ? 'Connected to room' : 'Enter details to join'}
        </p>
      </div>
      {joined() && isHover() && (
        <button
          onClick={handleLeave}
          class='absolute right-3 top-3 z-20 px-3 py-1 bg-red-500 text-white rounded hover:bg-red-600'
        >
          Leave
        </button>
      )}
      {joined() && (
        <div class='absolute right-3 top-3 z-20 ml-2 inline-block px-2 py-1 bg-green-600 text-white rounded text-xs'>
          Joined
        </div>
      )}
    </div>
  );
}
