#!/usr/bin/env node

// Direct WebSocket connection test
const WebSocket = require('ws');

console.log('🧪 Testing WebSocket Connection');
console.log('================================\n');

// Test configuration
const USER_ID = 1;
const USERNAME = 'testuser';
const ROOM_ID = 1;
const WS_URL = `ws://localhost:8080/ws?user_id=${USER_ID}&username=${USERNAME}&room_id=${ROOM_ID}`;

console.log('Configuration:');
console.log(`  User ID: ${USER_ID}`);
console.log(`  Username: ${USERNAME}`);
console.log(`  Room ID: ${ROOM_ID}`);
console.log(`  URL: ${WS_URL}\n`);

console.log('Connecting...\n');

const ws = new WebSocket(WS_URL);

let messageCount = 0;
let connected = false;

ws.on('open', () => {
  connected = true;
  console.log('✅ WebSocket CONNECTED\n');
  
  // Send join message
  const joinMessage = {
    type: 'join',
    room_id: ROOM_ID.toString(),
    user_id: USER_ID,
    username: USERNAME
  };
  
  console.log('📤 Sending join message:', JSON.stringify(joinMessage, null, 2));
  ws.send(JSON.stringify(joinMessage));
  
  // Keep connection alive for 10 seconds
  setTimeout(() => {
    console.log('\n⏰ Test duration complete (10 seconds)');
    console.log(`📊 Total messages received: ${messageCount}`);
    console.log('🔌 Closing connection...');
    ws.close(1000, 'Test complete');
  }, 10000);
});

ws.on('message', (data) => {
  messageCount++;
  try {
    const message = JSON.parse(data.toString());
    console.log(`\n📨 Message ${messageCount} received:`, JSON.stringify(message, null, 2));
  } catch (e) {
    console.log(`\n📨 Message ${messageCount} (raw):`, data.toString());
  }
});

ws.on('close', (code, reason) => {
  console.log(`\n❌ WebSocket CLOSED`);
  console.log(`   Code: ${code}`);
  console.log(`   Reason: ${reason.toString() || 'No reason provided'}`);
  
  const closeReasons = {
    1000: 'Normal closure',
    1001: 'Going away',
    1002: 'Protocol error',
    1003: 'Unsupported data',
    1006: 'Abnormal closure (no close frame)',
    1011: 'Server error',
    1012: 'Service restart'
  };
  
  console.log(`   Meaning: ${closeReasons[code] || 'Unknown'}`);
  
  if (code === 1006) {
    console.log('\n⚠️  ISSUE DETECTED: Close code 1006');
    console.log('   This means the server closed the connection abnormally.');
    console.log('   Check backend logs for errors.');
  } else if (code === 1000 && !connected) {
    console.log('\n⚠️  ISSUE: Connection closed before opening');
  } else if (code === 1000 && messageCount === 0) {
    console.log('\n⚠️  ISSUE: No messages received before close');
  } else if (code === 1000) {
    console.log('\n✅ Connection closed normally');
  }
  
  process.exit(code === 1000 ? 0 : 1);
});

ws.on('error', (error) => {
  console.log('\n❌ WebSocket ERROR:', error.message);
  console.log('   Full error:', error);
  process.exit(1);
});

// Timeout safety
setTimeout(() => {
  if (!connected) {
    console.log('\n❌ TIMEOUT: Could not connect after 5 seconds');
    console.log('   Check if backend is running on port 8080');
    process.exit(1);
  }
}, 5000);
