# Login/Register Fix - JSON Parse Error

## Issue
When trying to login or register, you got the error:
```
Unexpected token 'I', "Internal S"... is not valid JSON
```

## Root Cause
The frontend was calling `/api/auth/login` and `/api/auth/signup`, which are Next.js API routes that don't exist. Next.js was returning an HTML error page ("Internal Server Error"), which the frontend tried to parse as JSON, causing the error.

## Solution
Changed the frontend to call the backend API directly at `http://localhost:8080/auth/login` and `http://localhost:8080/auth/signup`.

## Changes Made

### 1. Login Page (`frontend/app/login/page.tsx`)
```typescript
// Before
const response = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(formData),
})
const data = await response.json()

// After
const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
const response = await fetch(`${apiUrl}/auth/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(formData),
})

// Better error handling
if (!response.ok) {
  const text = await response.text()
  let errorMessage = 'Login failed'
  try {
    const data = JSON.parse(text)
    errorMessage = data.error || errorMessage
  } catch {
    errorMessage = text || errorMessage
  }
  throw new Error(errorMessage)
}

const data = await response.json()
```

### 2. Register Page (`frontend/app/register/page.tsx`)
Same changes as login page, but for `/auth/signup` endpoint.

## Benefits

1. ✅ **Correct API endpoint** - Now calls backend directly
2. ✅ **Better error handling** - Handles both JSON and text responses
3. ✅ **Environment variable support** - Uses `NEXT_PUBLIC_API_URL` if set
4. ✅ **Clearer error messages** - Shows actual error from backend

## Testing

### 1. Register a New User
```bash
# Make sure backend is running
cd backend
go run cmd/server/main.go
```

Then in browser:
1. Go to `http://localhost:3000/register`
2. Fill in:
   - Username: `testuser`
   - Email: `test@example.com`
   - Password: `password123`
3. Click "Create Account"
4. Should redirect to lobby with 1,000 chips

### 2. Login with Existing User
1. Go to `http://localhost:3000/login`
2. Enter username and password
3. Click "Login"
4. Should redirect to lobby

## Backend Endpoints

The backend provides these auth endpoints:

### POST /auth/signup
```json
// Request
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}

// Response (201 Created)
{
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "chips": 1000,
    "wins": 0,
    "losses": 0
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

// Error Response (400 Bad Request)
{
  "error": "username or email already exists"
}
```

### POST /auth/login
```json
// Request
{
  "username": "testuser",
  "password": "password123"
}

// Response (200 OK)
{
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "chips": 1000,
    "wins": 0,
    "losses": 0
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

// Error Response (401 Unauthorized)
{
  "error": "invalid credentials"
}
```

## Environment Variables

Make sure your `.env.local` file in the frontend directory has:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

If the file doesn't exist, create it:
```bash
cd frontend
cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
EOF
```

## Common Issues

### Issue: "Failed to fetch"
**Cause:** Backend not running
**Solution:**
```bash
cd backend
go run cmd/server/main.go
```

### Issue: "CORS error"
**Cause:** Backend CORS not configured for frontend URL
**Solution:** Check backend CORS config in `backend/cmd/server/main.go`:
```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "http://localhost:3000",
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
}))
```

### Issue: "username or email already exists"
**Cause:** User already registered
**Solution:** Either:
1. Login with existing credentials
2. Use a different username/email
3. Delete user from database:
```sql
-- Connect to database
docker exec -it poker-postgres psql -U poker -d poker_arena

-- Delete user
DELETE FROM users WHERE username = 'testuser';
```

### Issue: "invalid credentials"
**Cause:** Wrong username or password
**Solution:** 
1. Check username spelling
2. Check password
3. Try registering a new account

## Verification

After the fix, you should see:

1. ✅ No JSON parse errors
2. ✅ Clear error messages from backend
3. ✅ Successful login/register redirects to lobby
4. ✅ Token and user data saved to localStorage
5. ✅ Can see your chips in the lobby

## Testing Commands

```bash
# Test signup endpoint directly
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# Test login endpoint directly
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Check if backend is running
curl http://localhost:8080/healthz
```

## Files Modified

- ✅ `frontend/app/login/page.tsx` - Fixed API endpoint and error handling
- ✅ `frontend/app/register/page.tsx` - Fixed API endpoint and error handling

## Summary

The login/register functionality now works correctly by:
1. Calling the correct backend endpoints
2. Handling both JSON and text error responses
3. Showing clear error messages to users
4. Using environment variables for API URL configuration

Try logging in or registering now - it should work without the JSON parse error!
