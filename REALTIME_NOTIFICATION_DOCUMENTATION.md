# Real-time Notification System Documentation

## Overview
Complete implementation of WebSocket-powered real-time notifications, instant chat system, push notifications, and live booking updates.

## Architecture

### Components
1. **WebSocket Hub** - Central connection manager
2. **Client Manager** - WebSocket client lifecycle
3. **Chat Service** - Instant messaging business logic
4. **Notification Service** - Push notification management
5. **FCM Service** - Firebase Cloud Messaging integration
6. **Presence System** - Online user tracking

## Features Implemented

### ✅ Live Booking Status Updates (WebSocket)
- Real-time booking status changes
- Payment confirmation notifications
- Delivery status updates
- Equipment dispatch notifications
- Auto-refund processing alerts

### ✅ Instant Chat System
- One-on-one vendor-customer messaging
- Group chat for booking discussions
- Real-time typing indicators
- Read receipts and message status
- File attachments support
- Message search functionality
- Conversation archiving

### ✅ Real-time Equipment Availability
- Live inventory tracking
- Booking conflict detection
- Available quantity updates
- Real-time reservation status
- Vendor inventory notifications

### ✅ Push Notifications (Mobile)
- Firebase Cloud Messaging (FCM) integration
- Apple Push Notification Service (APNS) support
- Device token management
- Quiet hours configuration
- Notification preferences

### ✅ Notification Preferences
- Email notification toggles
- SMS notification controls
- Push notification settings
- In-app notification preferences
- Quiet hours configuration
- Per-notification type settings

## WebSocket Protocol

### Connection
```javascript
// Connect to WebSocket
const ws = new WebSocket('wss://api.genrent.com/ws');

// Authenticate (send token after connection)
ws.send(JSON.stringify({
  type: 'authenticate',
  data: { token: 'your_jwt_token' }
}));
```

### Message Types

#### Client → Server
```json
// Join a room
{
  "type": "join_room",
  "data": { "room_id": "conversation_123" }
}

// Leave a room
{
  "type": "leave_room",
  "data": { "room_id": "conversation_123" }
}

// Send message to room
{
  "type": "room_message",
  "data": {
    "room_id": "conversation_123",
    "message": "Hello!"
  }
}

// Send message to user
{
  "type": "send_message",
  "data": {
    "recipient_id": 456,
    "message": "Private message"
  }
}

// Typing indicator
{
  "type": "typing_status",
  "data": {
    "conversation_id": 123,
    "is_typing": true
  }
}
```

#### Server → Client
```json
// New message notification
{
  "type": "new_message",
  "data": {
    "conversation_id": 123,
    "message_id": 456,
    "sender_id": 789,
    "message": "Hello!",
    "sent_at": 1691234567
  },
  "timestamp": "2024-08-05T10:30:45Z"
}

// Booking update
{
  "type": "booking_update",
  "data": {
    "booking_id": 123,
    "status": "confirmed",
    "timestamp": 1691234567
  },
  "timestamp": "2024-08-05T10:30:45Z"
}

// Typing status
{
  "type": "typing_status",
  "data": {
    "conversation_id": 123,
    "user_id": 456,
    "is_typing": true
  },
  "timestamp": "2024-08-05T10:30:45Z"
}

// User joined room
{
  "type": "user_joined",
  "data": {
    "user_id": 456,
    "room_id": "conversation_123"
  },
  "timestamp": "2024-08-05T10:30:45Z"
}
```

## API Endpoints

### WebSocket Connection

#### WS /ws
**Description**: Establish WebSocket connection

**Authentication**: Bearer token in query parameter or header

**Example**:
```javascript
const ws = new WebSocket('wss://api.genrent.com/ws?token=your_jwt_token');
```

### Online Status

#### GET /api/websocket/online-users
**Description**: Get list of online users

**Response**:
```json
{
  "online_users": [123, 456, 789],
  "count": 3
}
```

#### GET /api/websocket/online-count
**Description**: Get count of online users

**Response**:
```json
{
  "online_count": 15
}
```

### Chat Operations

#### POST /api/chat/conversations
**Description**: Create a new conversation

**Request**:
```json
{
  "vendor_id": 123,
  "booking_id": 456
}
```

#### GET /api/chat/conversations?role=vendor
**Description**: Get all conversations for current user

**Response**:
```json
{
  "conversations": [
    {
      "id": 1,
      "booking_id": 456,
      "vendor_id": 123,
      "customer_id": 789,
      "last_message": "When can you deliver?",
      "last_message_at": "2024-08-05T10:30:00Z",
      "vendor_read": false,
      "customer_read": true,
      "status": "active"
    }
  ]
}
```

#### GET /api/chat/conversations/:id/messages
**Description**: Get messages for a conversation

**Query Parameters**: `page`, `per_page`

**Response**:
```json
{
  "messages": [
    {
      "id": 1,
      "conversation_id": 1,
      "sender_id": 123,
      "receiver_id": 456,
      "message": "Hello!",
      "message_type": "text",
      "is_read": true,
      "read_at": "2024-08-05T10:31:00Z",
      "sent_at": "2024-08-05T10:30:00Z"
    }
  ],
  "total": 25,
  "page": 1,
  "per_page": 50
}
```

#### POST /api/chat/messages
**Description**: Send a message

**Request**:
```json
{
  "conversation_id": 1,
  "receiver_id": 456,
  "message": "Hello!",
  "message_type": "text"
}
```

#### PUT /api/chat/conversations/:id/messages/read
**Description**: Mark messages as read

#### GET /api/chat/unread-count
**Description**: Get unread message count

**Response**:
```json
{
  "unread_count": 5
}
```

#### POST /api/chat/conversations/:id/typing
**Description**: Set typing status

**Request**:
```json
{
  "is_typing": true
}
```

### Notification Operations

#### GET /api/notifications
**Description**: Get user notifications

**Response**:
```json
{
  "notifications": [
    {
      "id": 1,
      "user_id": 123,
      "type": "booking_update",
      "title": "Booking Confirmed!",
      "message": "Your booking has been confirmed",
      "read": false,
      "created_at": "2024-08-05T10:30:00Z"
    }
  ],
  "unread_count": 5
}
```

#### PUT /api/notifications/:id/read
**Description**: Mark notification as read

#### PUT /api/notifications/read-all
**Description**: Mark all notifications as read

#### DELETE /api/notifications/:id
**Description**: Delete a notification

### Notification Preferences

#### GET /api/notifications/preferences
**Description**: Get user notification preferences

**Response**:
```json
{
  "email_enabled": true,
  "sms_enabled": false,
  "push_enabled": true,
  "in_app_enabled": true,
  "booking_updates": true,
  "message_alerts": true,
  "promotions": false,
  "reviews": true,
  "availability": true,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "08:00",
  "timezone": "Asia/Kolkata"
}
```

#### PUT /api/notifications/preferences
**Description**: Update notification preferences

**Request**:
```json
{
  "push_enabled": false,
  "quiet_hours_start": "23:00",
  "quiet_hours_end": "07:00"
}
```

### Device Management

#### POST /api/devices/register
**Description**: Register a device for push notifications

**Request**:
```json
{
  "device_token": "firebase_device_token",
  "platform": "ios",
  "device_name": "iPhone 13",
  "app_version": "1.0.0",
  "os_version": "iOS 16.0"
}
```

#### DELETE /api/devices/:id
**Description**: Unregister a device

## Database Schema

### Tables

#### conversations
```sql
- id: BIGSERIAL PRIMARY KEY
- booking_id: BIGINT (FK to bookings)
- vendor_id: BIGINT (FK to users)
- customer_id: BIGINT (FK to users)
- last_message: TEXT
- last_message_at: TIMESTAMP
- vendor_read: BOOLEAN
- customer_read: BOOLEAN
- status: VARCHAR(20) - 'active', 'archived', 'blocked'
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

#### messages
```sql
- id: BIGSERIAL PRIMARY KEY
- conversation_id: BIGINT (FK to conversations)
- sender_id: BIGINT (FK to users)
- receiver_id: BIGINT (FK to users)
- message: TEXT
- message_type: VARCHAR(20) - 'text', 'image', 'file', 'system'
- attachment_url: VARCHAR
- is_read: BOOLEAN
- read_at: TIMESTAMP
- sent_at: TIMESTAMP
- created_at: TIMESTAMP
```

#### notification_preferences
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT UNIQUE (FK to users)
- email_enabled: BOOLEAN
- sms_enabled: BOOLEAN
- push_enabled: BOOLEAN
- in_app_enabled: BOOLEAN
- booking_updates: BOOLEAN
- message_alerts: BOOLEAN
- promotions: BOOLEAN
- reviews: BOOLEAN
- availability: BOOLEAN
- quiet_hours_start: VARCHAR(5)
- quiet_hours_end: VARCHAR(5)
- timezone: VARCHAR(50)
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

#### device_registrations
```sql
- id: BIGSERIAL PRIMARY KEY
- user_id: BIGINT (FK to users)
- device_token: VARCHAR UNIQUE
- platform: VARCHAR(20) - 'ios', 'android', 'web'
- device_name: VARCHAR(100)
- app_version: VARCHAR(20)
- os_version: VARCHAR(20)
- is_active: BOOLEAN
- last_used_at: TIMESTAMP
- created_at: TIMESTAMP
- updated_at: TIMESTAMP
```

#### real_time_inventory
```sql
- id: BIGSERIAL PRIMARY KEY
- equipment_id: BIGINT UNIQUE (FK to equipment)
- available_qty: INT
- reserved_qty: INT
- last_updated: TIMESTAMP
- updated_by: BIGINT (FK to users)
- created_at: TIMESTAMP
```

#### typing_status
```sql
- id: BIGSERIAL PRIMARY KEY
- conversation_id: BIGINT (FK to conversations)
- user_id: BIGINT (FK to users)
- is_typing: BOOLEAN
- updated_at: TIMESTAMP
```

#### user_presence
```sql
- user_id: BIGINT PRIMARY KEY (FK to users)
- is_online: BOOLEAN
- last_seen: TIMESTAMP
- device_info: JSONB
- updated_at: TIMESTAMP
```

## Usage Examples

### Client-side WebSocket Implementation

```javascript
class GenRentWebSocket {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.reconnectInterval = null;
    this.messageHandlers = {};
  }

  connect() {
    const ws = new WebSocket(`wss://api.genrent.com/ws?token=${this.token}`);

    ws.onopen = () => {
      console.log('WebSocket connected');
      this.clearReconnect();
    };

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.scheduleReconnect();
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws = ws;
  }

  handleMessage(message) {
    const handler = this.messageHandlers[message.type];
    if (handler) {
      handler(message.data);
    }
  }

  on(type, handler) {
    this.messageHandlers[type] = handler;
  }

  send(type, data) {
    this.ws.send(JSON.stringify({ type, data }));
  }

  joinRoom(roomId) {
    this.send('join_room', { room_id: roomId });
  }

  leaveRoom(roomId) {
    this.send('leave_room', { room_id: roomId });
  }

  sendMessage(roomId, message) {
    this.send('room_message', {
      room_id: roomId,
      message
    });
  }

  scheduleReconnect() {
    if (!this.reconnectInterval) {
      this.reconnectInterval = setInterval(() => {
        this.connect();
      }, 5000);
    }
  }

  clearReconnect() {
    if (this.reconnectInterval) {
      clearInterval(this.reconnectInterval);
      this.reconnectInterval = null;
    }
  }
}

// Usage
const ws = new GenRentWebSocket('your_jwt_token');
ws.connect();

// Handle new messages
ws.on('new_message', (data) => {
  console.log('New message:', data);
  updateChatUI(data);
});

// Handle booking updates
ws.on('booking_update', (data) => {
  console.log('Booking update:', data);
  updateBookingStatus(data);
});

// Handle typing indicators
ws.on('typing_status', (data) => {
  showTypingIndicator(data.user_id, data.is_typing);
});

// Join conversation room
ws.joinRoom('conversation_123');
```

### React Hook Example

```javascript
function useWebSocket(token) {
  const [ws, setWS] = useState(null);
  const [connected, setConnected] = useState(false);
  const [messages, setMessages] = useState([]);
  const [typingUsers, setTypingUsers] = useState(new Set());

  useEffect(() => {
    const websocket = new GenRentWebSocket(token);

    websocket.on('new_message', (data) => {
      setMessages(prev => [...prev, data]);
    });

    websocket.on('typing_status', (data) => {
      setTypingUsers(prev => {
        const next = new Set(prev);
        if (data.is_typing) {
          next.add(data.user_id);
        } else {
          next.delete(data.user_id);
        }
        return next;
      });
    });

    websocket.connect();
    setWS(websocket);
    setConnected(true);

    return () => {
      websocket.ws.close();
    };
  }, [token]);

  return { ws, connected, messages, typingUsers };
}
```

## Configuration

### Environment Variables

```env
# WebSocket Configuration
WEBSOCKET_ENABLED=true
WEBSOCKET_PATH=/ws
WEBSOCKET_READ_BUFFER=1024
WEBSOCKET_WRITE_BUFFER=1024
WEBSOCKET_PING_INTERVAL=54
WEBSOCKET_PONG_TIMEOUT=60

# Firebase Cloud Messaging
FCM_ENABLED=true
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json
FCM_API_KEY=your_fcm_server_key

# Push Notification Settings
PUSH_NOTIFICATION_ENABLED=true
PUSH_QUIET_HOURS_START=22:00
PUSH_QUIET_HOURS_END=08:00
PUSH_DEFAULT_TIMEZONE=Asia/Kolkata

# Notification Settings
NOTIFICATION_RETENTION_DAYS=90
NOTIFICATION_CLEANUP_HOUR=2
```

## Firebase Setup

### 1. Create Firebase Project
1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Create a new project
3. Add Android/iOS app
4. Download `google-services.json` / `GoogleService-Info.plist`

### 2. Generate Server Key
1. Go to Project Settings → Cloud Messaging
2. Generate Server Key
3. Copy credentials to environment variables

### 3. Configure FCM Service
```go
fcmService, err := notification.NewFCMService(
    os.Getenv("FCM_CREDENTIALS_PATH"),
)
```

## Implementation Status

### ✅ Completed Features
- [x] WebSocket server with connection management
- [x] Live booking status updates
- [x] Instant chat system
- [x] Real-time availability tracking
- [x] Push notification infrastructure
- [x] Notification preferences
- [x] Typing indicators
- [x] Read receipts
- [x] Online presence tracking
- [x] Message search
- [x] Conversation archiving
- [x] Quiet hours configuration

### 🔧 Configuration Required
1. **WebSocket Server**
   ```env
   WEBSOCKET_ENABLED=true
   ```

2. **Firebase Setup**
   - Create Firebase project
   - Download credentials
   - Configure FCM service

3. **Database Migration**
   ```bash
   psql -U genrent -d genrent_db -f internal/migrate/005_realtime_system.sql
   ```

## Monitoring & Analytics

### Key Metrics
- **WebSocket Connections**: Active connection count
- **Message Volume**: Messages sent/received per day
- **Chat Engagement**: Active conversations
- **Push Delivery**: Push notification success rate
- **Online Users**: Real-time online count

### Analytics Queries
```sql
-- Daily message statistics
SELECT
    DATE(sent_at) as date,
    COUNT(*) as message_count,
    COUNT(DISTINCT conversation_id) as conversation_count
FROM messages
WHERE sent_at > CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(sent_at)
ORDER BY date DESC;

-- User engagement
SELECT
    u.id,
    u.name,
    COUNT(DISTINCT c.id) as conversation_count,
    COUNT(m.id) as message_count
FROM users u
LEFT JOIN conversations c ON (u.id = c.vendor_id OR u.id = c.customer_id)
LEFT JOIN messages m ON m.conversation_id = c.id
GROUP BY u.id
ORDER BY message_count DESC;

-- Online users by hour
SELECT
    EXTRACT(HOUR FROM updated_at) as hour,
    COUNT(*) as online_count
FROM user_presence
WHERE is_online = true
GROUP BY hour
ORDER BY hour;
```

## Troubleshooting

### Common Issues

#### 1. WebSocket Connection Failed
**Solution**: Check CORS settings and authentication
```go
upgrader.CheckOrigin = func(r *http.Request) bool {
    return true // Configure based on your domain
}
```

#### 2. Push Notifications Not Working
**Solution**: Verify FCM credentials and device tokens
```bash
# Test FCM connection
curl -X POST https://fcm.googleapis.com/v1/projects/YOUR_PROJECT/messages:send
```

#### 3. Typing Indicators Not Working
**Solution**: Ensure clients are in the same room
```javascript
ws.joinRoom('conversation_123');
```

## Future Enhancements

### Planned Features
- [ ] Voice messages support
- [ ] Video calling integration
- [ ] Group chat for events
- [ ] Message reactions and emojis
- [ ] File sharing and document preview
- [ ] Message encryption
- [ ] Chat analytics dashboard
- [ ] Automated chat responses
- [ ] Chat bot integration
- [ ] Multi-language support

---

**Last Updated**: 2026-07-14
**Version**: 1.0
**Status**: Production Ready
