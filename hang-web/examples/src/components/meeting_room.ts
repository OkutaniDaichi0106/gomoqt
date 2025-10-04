import { VideoRenderer,AudioOffloader,VideoPreviewer } from "@okutanidaichi/hang";
import type { RoomElement,BroadcastPublisher,BroadcastSubscriber } from "@okutanidaichi/hang";
import { Camera, Microphone } from "@okutanidaichi/hang/media";
import type {
    ProfileTrackDescriptor,
    VideoTrackDescriptor,
    AudioTrackDescriptor,
TrackDescriptor,
} from "@okutanidaichi/hang/catalog";
import {
    VideoConfigSchema,
    AudioConfigSchema,
VideoTrackSchema,
AudioTrackSchema
} from "@okutanidaichi/hang/catalog";
import {
    VideoTrackEncoder,
    AudioTrackProcessor,
    VideoTrackProcessor,
    AudioTrackEncoder,
    videoEncoderConfig,
    audioEncoderConfig,
    VideoTrackDecoder,
    AudioTrackDecoder,
} from "@okutanidaichi/hang/internal";

export async function setupMeetingRoom(room: RoomElement, local: BroadcastPublisher): Promise<void> {
    await Promise.allSettled([
        setCameraTrack(local, room).then(() => console.log("Camera track setup complete")).catch((e) => console.warn("Camera track setup failed:", e)),
        setMicrophoneTrack(local).then(() => console.log("Microphone track setup complete")).catch((e) => console.warn("Microphone track setup failed:", e)),
    ])

    console.debug("Local tracks set up");

    room.onjoin = async (member) => {
        if (!member.remote) {
            // Local member joined
            // Ignore local member joins
            return;
        }

        console.debug("Remote member joined:", member.name);

        // // Create participant container for this member
        // const participantDiv = document.createElement("div");
        // participantDiv.className = "participant";
        // participantDiv.style.cssText = `
        //     margin: 10px;
        //     padding: 10px;
        //     border: 1px solid #ccc;
        //     border-radius: 8px;
        //     display: inline-block;
        //     text-align: center;
        // `;

        // // Add participant name
        // const nameDiv = document.createElement("div");
        // nameDiv.textContent = `${member.name}`;
        // nameDiv.style.cssText = "margin-bottom: 10px; font-weight: bold;";
        // participantDiv.appendChild(nameDiv);

        // // Add participant to room's remote participants section
        // const remoteParticipantsContainer = room.querySelector('.remote-participants');
        // if (remoteParticipantsContainer && !remoteParticipantsContainer.querySelector(`[data-member-name="${member.name}"]`)) {
        //     participantDiv.setAttribute('data-member-name', member.name);
        //     participantDiv.setAttribute('data-member-type', 'remote');
        //     remoteParticipantsContainer.appendChild(participantDiv);
        // }

        while(true) {
            const [desc, err] = await member.broadcast.nextTrack();
            if (err) {
                console.error("Error receiving track:", err);
                break;
            }

            console.debug("Remote member track description:", desc);

            switch (desc.name) {
                case "camera":
                    if (desc.schema !== "video") {
                        console.error("Expected video schema for camera track");
                        break;
                    }
                    await getCameraTrack(VideoTrackSchema.parse(desc.config), member.broadcast);
                    break;
                case "microphone":
                    if (desc.schema !== "audio") {
                        console.error("Expected audio schema for microphone track");
                        break;
                    }
                    await getMicrophoneTrack(AudioTrackSchema.parse(desc.config), member.broadcast);
                    break;
                default:
                    break;
            }
        }
    };
    room.onleave = (member) => {
        // Remove participant's canvas when they leave
        if (member.remote) {
            const remoteParticipantsContainer = room.querySelector('.remote-participants');
            if (remoteParticipantsContainer) {
                const participantDiv = remoteParticipantsContainer.querySelector(`[data-member-name="${member.name}"]`);
                if (participantDiv) {
                    participantDiv.remove();
                }
            }
        }
    };
    return;
}

async function setCameraTrack(local: BroadcastPublisher, room: RoomElement): Promise<void> {
    const camera = new Camera({enabled: true});

    // Create VideoPreviewer with custom virtual content
    const videoPreviewer = new VideoPreviewer({
        width: 320,
        height: 240,
        source: camera.getVideoTrack(),
        virtualContent: {
            backgroundColor: '#1a1a1a',
            textColor: '#4CAF50',
            title: 'Camera Loading',
            subtitle: 'Please wait...',
            fontSize: 18
        }
    });

    // Add preview canvas to local participant area
    const localParticipantContainer = room.querySelector('.local-participant') ||
                                      room.querySelector('.participants-container') ||
                                      room;
    if (localParticipantContainer) {
        const localPreviewDiv = document.createElement("div");
        localPreviewDiv.className = "local-preview";
        localPreviewDiv.style.cssText = `
            margin: 10px;
            text-align: center;
        `;

        const labelDiv = document.createElement("div");
        labelDiv.textContent = "You (Preview)";
        labelDiv.style.cssText = "margin-bottom: 5px; font-weight: bold; color: #4CAF50;";

        localPreviewDiv.appendChild(labelDiv);
        localPreviewDiv.appendChild(videoPreviewer.canvas);
        localParticipantContainer.appendChild(localPreviewDiv);
    }

    // Get encoder from VideoPreviewer
    const cameraEncoder = await videoPreviewer.encoder();

    // Configure encoder
    const cameraEncoderConfig = await videoEncoderConfig({
        width: 640,
        height: 480,
        frameRate: 30,
    });

    console.debug("Camera encoder config:", cameraEncoderConfig);

    const cameraDecoderConfig = await cameraEncoder.configure(cameraEncoderConfig);

    console.debug("Camera decoder config:", cameraDecoderConfig);

    // Create a video track descriptor
    const videoDesc: VideoTrackDescriptor = {
        name: "camera",
        priority: 1,
        schema: "video",
        config: VideoConfigSchema.parse({
            ...cameraDecoderConfig,
            // Convert ArrayBuffer to Uint8Array if present
            description: cameraDecoderConfig.description 
                ? (cameraDecoderConfig.description instanceof ArrayBuffer 
                    ? new Uint8Array(cameraDecoderConfig.description)
                    : cameraDecoderConfig.description)
                : undefined,
            container: "loc"
        }),
    };

    // Close existing track if any
    const gotCamera = local.getTrack(videoDesc.name);
    if (gotCamera) {
        await gotCamera.close();
    }

    // Set the new track
    local.setTrack(videoDesc, cameraEncoder);

    console.debug("Camera track set successfully with VideoPreviewer");
}

async function setMicrophoneTrack(local: BroadcastPublisher): Promise<void> {
    // Get real microphone track (no virtual fallback)
    const microphone = new Microphone({enabled: true});
    const microphoneTrack = await microphone.getAudioTrack();
    console.log("Using real microphone track");

    const settings = microphoneTrack.getSettings();

    // Determine encoder config first
    const microphoneEncoderConfig = await audioEncoderConfig({
        sampleRate: settings.sampleRate || 48000,
        channels: settings.channelCount ?? 2,
    });

    // Create processor with target channel configuration
    const microphoneProcessor = new AudioTrackProcessor(microphoneTrack);

    const microphoneEncoder = new AudioTrackEncoder({
        source: microphoneProcessor.readable,
    });

    const microphoneDecoderConfig = await microphoneEncoder.configure(microphoneEncoderConfig);

    // Create an audio track descriptor
    const audioDesc: AudioTrackDescriptor = {
        name: "microphone",
        priority: 2,
        schema: "audio",
        config: AudioConfigSchema.parse({
            ...microphoneDecoderConfig,
            // Convert ArrayBuffer to Uint8Array if present
            description: microphoneDecoderConfig.description 
                ? (microphoneDecoderConfig.description instanceof ArrayBuffer 
                    ? new Uint8Array(microphoneDecoderConfig.description)
                    : microphoneDecoderConfig.description)
                : undefined,
            container: "loc"
        }),
    };

    // Close existing track if any
    const got = local.getTrack(audioDesc.name);
    if (got) {
        await got.close();
    }

    // Set the microphone track
    local.setTrack(audioDesc, microphoneEncoder);
}

async function getCameraTrack(desc: VideoTrackDescriptor, subscriber: BroadcastSubscriber): Promise<void> {
    // VideoRenderer creates canvas internally
    const videoRenderer = new VideoRenderer({
        width: 320,
        height: 240
    });

    const decoder = await videoRenderer.decoder();

    decoder.configure(desc.config);

    subscriber.subscribeTrack(desc.name, decoder);

    console.log(`Camera track set up`);
}

async function getMicrophoneTrack(desc: AudioTrackDescriptor, subscriber: BroadcastSubscriber): Promise<void> {
    const audioOffloader = new AudioOffloader({
        sampleRate: desc.config.sampleRate,
        numberOfChannels: desc.config.numberOfChannels,
    });

    const decoder = await audioOffloader.decoder();

    decoder.configure(desc.config);

    subscriber.subscribeTrack(desc.name, decoder);

    console.log(`Microphone track set up`);
}