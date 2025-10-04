import { createSignal, createResource } from 'solid-js';
import { Session } from '@okutanidaichi/moqt';
import { defineRoom,BroadcastPublisher } from '@okutanidaichi/hang';
import type { RoomElement, JoinedMember } from '@okutanidaichi/hang';
import { setupMeetingRoom } from "./meeting_room";
import { connect } from "./connection"

defineRoom();

// Local re-declare (runtime shape) to avoid depending on library export yet
interface RoomLifecycleStatus { type: 'idle' | 'connecting' | 'success' | 'error' | 'left' | 'disconnected'; message: string }

interface HangRoomProps {
  roomId: string;
}

export default function HangRoom(props: HangRoomProps) {
  let room!: RoomElement;
  const [isHover, setIsHover] = createSignal(false);
  const [roomId, setRoomId] = createSignal(props.roomId);
  const [localName, setLocalName] = createSignal("Guest");
  const [joined, setJoined] = createSignal(false);
  const [session] = createResource(async () => {
    try {
      const moqUrl = import.meta.env.VITE_MOQ_URL;
      const conn = await connect(moqUrl);
      const sess = new Session({conn});
      await sess.ready;
      return sess;
    } catch (e) {
      throw e;
    }
  });

  const handleJoin = async () => {
    if (joined() || session.loading || !session()) return;
    try {
      const local = new BroadcastPublisher(localName())
      await setupMeetingRoom(room, local);
      await room.join(session()!, local);
      console.debug("Joined room")
      setJoined(true);
    } catch (e) {
      // Optionally handle join error (e.g., show alert or reset state)
    }
  };

  const handleLeave = () => {
    if (!joined()) return;
    room.leave();
    setJoined(false);
  };

  return (
    <div class='hang-room-container relative w-full aspect-video bg-gray-100 rounded-lg overflow-hidden' style="min-height: 400px; max-height: 80vh;">
      <hang-room
        ref={room}
        attr:room-id={roomId()}
        // attr:local-name={localName()}
        attr:description="Demo room"
        class="block w-full h-full absolute inset-0"
        style="width: 100%; height: 100%;"
      />
      
      {/* Pre-join form display */}
      {!joined() && (
        <div class="absolute inset-0 flex items-center justify-center p-6 bg-gradient-to-br from-slate-50/95 to-slate-100/95 z-10 backdrop-blur-sm">
          <div class="bg-white/90 backdrop-blur-lg p-6 rounded-xl shadow-xl max-w-sm w-full border border-white/30">
            <div class="text-center mb-6">
              <h2 class="text-lg font-bold text-slate-800 mb-2">Join Room</h2>
              <div class="w-8 h-0.5 bg-gradient-to-r from-blue-500 to-purple-500 mx-auto rounded-full"></div>
            </div>
            
            <div class='mb-4'>
              <label class='block text-sm font-semibold text-slate-700 mb-1.5'>Room ID</label>
              <input
                type='text'
                value={roomId()}
                disabled={session.loading || !session()}
                onInput={(e) => setRoomId(e.target.value)}
                class='w-full px-3 py-2 border border-slate-200 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 disabled:bg-slate-50 transition-all duration-200 text-sm'
                placeholder="Enter room ID"
              />
            </div>
            
            <div class='mb-4'>
              <label class='block text-sm font-semibold text-slate-700 mb-1.5'>Display Name</label>
              <input
                type='text'
                value={localName()}
                disabled={session.loading || !session()}
                onInput={(e) => setLocalName(e.target.value)}
                class='w-full px-3 py-2 border border-slate-200 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 disabled:bg-slate-50 transition-all duration-200 text-sm'
                placeholder="Your name"
              />
            </div>
            
            <button
              onClick={handleJoin}
              disabled={session.loading || !session()}
              class='w-full mb-3 px-4 py-2.5 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-lg hover:from-blue-600 hover:to-purple-700 disabled:from-slate-400 disabled:to-slate-500 shadow-lg transition-all duration-200 font-semibold text-sm'
            >
              {session.loading ? 'Connecting...' : 'Join Room'}
            </button>
            
            <p class='text-center text-xs text-slate-500'>
              {session.loading ? 'Establishing connection...' : session.error ? 'Connection failed' : 'Ready to join'}
            </p>
          </div>
        </div>
      )}

      {/* Post-join hover overlay */}
      {joined() && (
        <div
          class="absolute inset-0 z-10 w-full h-full"
          onMouseEnter={() => setIsHover(true)}
          onMouseLeave={() => setIsHover(false)}
        >
          {/* Hover overlay */}
          {isHover() && (
            <div class="absolute inset-0 bg-black/40 transition-opacity duration-300 backdrop-blur-sm flex items-center justify-center">
              <button
                onClick={handleLeave}
                class='px-8 py-4 bg-red-500/90 text-white rounded-lg hover:bg-red-600 shadow-xl backdrop-blur-sm border border-red-400/30 transition-all duration-200 font-medium text-lg'
              >
                Leave Room
              </button>
            </div>
          )}
          
          {/* ON AIR indicator - only visible when not hovering */}
          {!isHover() && (
            <div class='absolute right-6 top-6 flex items-center gap-2'>
              <div class='w-3 h-3 bg-red-500 rounded-full animate-pulse'></div>
              <span class='text-white text-sm font-medium drop-shadow-lg'>ON AIR</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
