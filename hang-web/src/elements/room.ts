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
    #room?: Room;
    #statusState: RoomLifecycleStatus = { type: 'idle', message: 'Idle' };

    // Public callback properties
    onjoin?: (meta: JoinedMember) => void;
    onleave?: (meta: LeftMember) => void;
    onstatus?: (status: RoomLifecycleStatus) => void;
    localHandler?: (publisher: BroadcastPublisher) => void;

    // Static methods
    static get observedAttributes(): string[] {
        return ['room-id', 'local-name', 'description'];
    }

    // Constructor
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
    }

    // Lifecycle callbacks
    connectedCallback(): void {
        this.render();
    }

    disconnectedCallback(): void {
        // Leave the room if connected
        if (this.#room) {
            this.#room.leave();
        }
        this.#setStatus({ type: 'disconnected', message: 'Disconnected' });
    }

    attributeChangedCallback(name: string, oldValue: string, newValue: string): void {
        // Leave room if attributes change while connected
        if (oldValue !== newValue && this.#room) {
            this.leave();
        }
    }

    // Public methods
    join(session: Session): void {
        const roomId = this.getAttribute('room-id');
        const localName = this.getAttribute('local-name');
        const description = this.getAttribute('description');

        if (!roomId || !localName) {
            this.#setStatus({ type: 'error', message: 'room-id or local-name is missing' });
            return;
        }

        try {
            this.#setStatus({ type: 'connecting', message: `Connecting to room ${roomId}...` });
            const room = new Room({
                roomID: roomId,
                description: description || undefined,
                onmember: {
                    onJoin: this.#onJoin.bind(this),
                    onLeave: this.#onLeave.bind(this)
                }
            });

            room.join(session, localName).then((local) => {
                this.localHandler?.(local);
            });

            this.#room = room;
            this.#setStatus({ type: 'success', message: `✓ Joined room ${room.roomID} as ${localName}` });
        } catch (e) {
            this.#setStatus({ type: 'error', message: `Failed to join: ${e instanceof Error ? e.message : String(e)}` });
        }
    }

    leave(): void {
        if (!this.#room) return;
        const id = this.#room.roomID;
        this.#room.leave();
        this.#room = undefined;
        this.#setStatus({ type: 'left', message: `Left room ${id}` });
    }

    render(): void {
        if (!this.shadowRoot) return;
        this.shadowRoot.innerHTML = `
            <style>
                :host { display: block; }
                .room-info { margin-bottom: 1rem; }
                .room-status-display {
                    padding: 0.75rem; margin: 0.5rem 0; border-radius: 0.375rem;
                    font-size: 0.875rem; line-height: 1.25rem; border: 1px solid transparent;
                }
                .status-connecting { background-color: #fef3c7; color: #92400e; border-color: #fcd34d; }
                .status-success { background-color: #d1fae5; color: #065f46; border-color: #a7f3d0; }
                .status-error { background-color: #fee2e2; color: #991b1b; border-color: #fca5a5; }
                .status-left { background-color: #e0e7ff; color: #3730a3; border-color: #c7d2fe; }
                .status-disconnected { background-color: #e5e7eb; color: #374151; border-color: #d1d5db; }
                .status-idle { background-color: #f3f4f6; color: #374151; border-color: #e5e7eb; }
            </style>
            <div class="room-info">
                <h3>Room: ${this.getAttribute('room-id') || 'Not specified'}</h3>
                <p>Local Name: ${this.getAttribute('local-name') || 'Not specified'}</p>
            </div>
            <slot name="status">
                <div class="room-status-display status-${this.#statusState.type}">${this.#statusState.message}</div>
            </slot>
            <div id="local-participant"></div>
            <div id="remote-participants"></div>
        `;
        // Ensure status is set after render
        this.#setStatus(this.#statusState);
    }

    // Private methods
    #setStatus(status: RoomLifecycleStatus): void {
        this.#statusState = status;

        // Update slotted status element if present
        const slotted = this.querySelector('[slot="status"]') as HTMLElement | null;
        if (slotted) {
            slotted.textContent = status.message;
            slotted.className = `room-status-display status-${status.type}`;
            this.onstatus?.(status);
            this.dispatchEvent(new CustomEvent('statuschange', { detail: status, bubbles: true, composed: true }));
            return;
        }

        // Update shadow DOM status display
        if (!this.shadowRoot) return;
        const statusDisplay = this.shadowRoot.querySelector('.room-status-display') as HTMLElement;
        if (statusDisplay) {
            statusDisplay.textContent = status.message;
            statusDisplay.className = `room-status-display status-${status.type}`;
        }

        // Call callback and dispatch event
        this.onstatus?.(status);
        this.dispatchEvent(new CustomEvent('statuschange', { detail: status, bubbles: true, composed: true }));
    }

    #onJoin(member: JoinedMember): void {
        // Add slot for the member
        const container = member.remote ? this.shadowRoot?.querySelector('#remote-participants') : this.shadowRoot?.querySelector('#local-participant');
        if (container) {
            const slot = document.createElement('slot');
            slot.name = member.remote ? `remote-${member.name}` : `local-${member.name}`;
            container.appendChild(slot);
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

    async #onLeave(member: LeftMember): Promise<void> {
        // Remove slot for the member
        const container = member.remote ? this.shadowRoot?.querySelector('#remote-participants') : this.shadowRoot?.querySelector('#local-participant');
        if (container) {
            const slot = container.querySelector(`slot[name="${member.remote ? `remote-${member.name}` : `local-${member.name}`}"]`);
            if (slot) slot.remove();
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