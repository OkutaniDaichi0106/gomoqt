import type {
    JoinedMember,
    LeftMember,
    BroadcastPublisher,
} from "../";
import {
    Room,
} from "../";
import type {
    Session
} from "@okutanidaichi/moqt";

// Extended status type includes lifecycle states
export type RoomLifecycleStatus = { type: 'idle' | 'connecting' | 'success' | 'error' | 'left' | 'disconnected'; message: string };

export class RoomElement extends HTMLElement {
    // Private properties
    room?: Room; // Made public for testing
    #statusState: RoomLifecycleStatus = { type: 'idle', message: 'Idle' };

    // Public callback properties
    onjoin?: (meta: JoinedMember) => void;
    onleave?: (meta: LeftMember) => void;
    onstatus?: (status: RoomLifecycleStatus) => void;
    // localHandler?: (publisher: BroadcastPublisher) => void;

    // Static methods
    static get observedAttributes(): string[] {
        return ['room-id', 'description'];
    }

    // Constructor
    constructor() {
        super();
        // Remove Shadow DOM - use Light DOM instead
    }

    // Lifecycle callbacks
    connectedCallback(): void {
        this.render();
    }

    disconnectedCallback(): void {
        // Leave the room if connected
        if (this.room) {
            this.room.leave();
        }
        this.#setStatus({ type: 'disconnected', message: 'Disconnected' });
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string): void {
        // Only leave room if we're connected and the change is significant
        if (oldValue !== newValue && this.room) {
            // Only leave if room-id changes, not for description
            if (name === 'room-id') {
                this.leave();
            }
        }
    }

    // Public methods
    async join(session: Session, local: BroadcastPublisher): Promise<void> {
        const roomId = this.getAttribute('room-id');
        // const localName = this.getAttribute('local-name');
        const description = this.getAttribute('description');

        if (!roomId) {
            this.#setStatus({ type: 'error', message: 'room-id is missing' });
            return;
        }

        try {
            this.#setStatus({ type: 'connecting', message: `Connecting to room ${roomId}...` });

            // Clear any existing member elements by re-rendering the DOM structure
            this.render();

            const room = new Room({
                roomID: roomId,
                description: description || undefined,
                onmember: {
                    onJoin: this.#onJoin.bind(this),
                    onLeave: this.#onLeave.bind(this)
                }
            });

            await room.join(session, local);

            this.room = room;
            this.#setStatus({ type: 'success', message: `✓ Joined room ${room.roomID} as ${local.name}` });
        } catch (e) {
            this.#setStatus({ type: 'error', message: `Failed to join: ${e instanceof Error ? e.message : String(e)}` });
        }
    }

    leave(): void {
        if (!this.room) {
            return;
        }
        const id = this.room.roomID;

        this.room.leave();
        this.room = undefined;
        this.#setStatus({ type: 'left', message: `Left room ${id}` });

        // Clear all member elements by re-rendering the DOM structure
        this.render();
    }

    render(): void {
        // Minimal Light DOM structure - no styling
        this.innerHTML = `
            <div class="room-status-display status-${this.#statusState.type}">${this.#statusState.message}</div>
            <div class="local-participant"></div>
            <div class="remote-participants"></div>
        `;
    }

    // Private methods
    #setStatus(status: RoomLifecycleStatus): void {
        this.#statusState = status;

        // Update status display content and class
        const statusDisplay = this.querySelector('.room-status-display') as HTMLElement;
        if (statusDisplay) {
            statusDisplay.textContent = status.message;
            statusDisplay.className = `room-status-display status-${status.type}`;
        }

        // Call callback and dispatch event
        this.onstatus?.(status);
        this.dispatchEvent(new CustomEvent('statuschange', { detail: status, bubbles: true, composed: true }));
    }

    #onJoin(member: JoinedMember): void {
        // Add participant to Light DOM (idempotent by name+type)
        const container = member.remote ? this.querySelector('.remote-participants') : this.querySelector('.local-participant');
        if (!container) {
        } else {
            const participantDiv = document.createElement('div');
            participantDiv.className = member.remote ? `remote-member remote-member-${member.name}` : `local-member local-member-${member.name}`;
            participantDiv.setAttribute('data-member-name', member.name);
            participantDiv.setAttribute('data-member-type', member.remote ? 'remote' : 'local');
            participantDiv.textContent = member.name;
            container.appendChild(participantDiv);
        }

        // Dispatch join event
        this.dispatchEvent(new CustomEvent('join', { detail: member, bubbles: true, composed: true }));

        // Call user-provided handler
        if (this.onjoin) {
            try {
                this.onjoin(member);
            } catch (e) {
                this.#setStatus({ type: 'error', message: `onjoin handler failed: ${e instanceof Error ? e.message : String(e)}` });
            }
        }
    }

    #onLeave(member: LeftMember): void {
        // Remove participant from Light DOM by matching name+type
        const container = member.remote ? this.querySelector('.remote-participants') : this.querySelector('.local-participant');
        if (!container) {
            return;
        }

        const selector = `[data-member-name="${member.name}"][data-member-type="${member.remote ? 'remote' : 'local'}"]`;
        const participantDiv = container.querySelector(selector);
        if (participantDiv) {
            participantDiv.remove();
        } else {

        }

        // Dispatch leave event
        this.dispatchEvent(new CustomEvent('leave', { detail: member, bubbles: true, composed: true }));

        // Call user-provided handler
        if (this.onleave) {
            try {
                this.onleave(member);
            } catch (e) {
                this.#setStatus({ type: 'error', message: `onleave handler failed: ${e instanceof Error ? e.message : String(e)}` });
            }
        }
    }
}

export function defineRoom(name: string = 'hang-room'): void {
    if (!customElements.get(name)) {
        customElements.define(name, RoomElement);
    } else {
        console.warn(`Custom element with name ${name} is already defined.`);
    }
}