import { Channel } from 'golikejs/sync';

async function testChannel() {
    const ch = new Channel(2);
    
    console.log('Creating channel...');
    console.log('Channel closed?', ch.closed);
    
    // Send some data
    await ch.send(1);
    await ch.send(2);
    console.log('Sent 1 and 2');
    
    // Close the channel
    ch.close();
    console.log('Channel closed?', ch.closed);
    
    // Try to receive after closing
    try {
        console.log('Trying to receive from closed channel...');
        const val = await ch.receive();
        console.log('Received:', val);
    } catch (e) {
        console.log('Error receiving from closed channel:', e);
    }
    
    // Try to receive when empty and closed
    try {
        console.log('Trying to receive again from empty closed channel...');
        const val = await ch.receive();
        console.log('Received:', val);
    } catch (e) {
        console.log('Error receiving from empty closed channel:', e);
    }
    
    // Try to receive one more time to see what happens
    try {
        console.log('Trying to receive AGAIN from empty closed channel...');
        const val = await ch.receive();
        console.log('Received:', val);
    } catch (e) {
        console.log('Error on third receive:', e);
    }
}

testChannel().catch(console.error);
