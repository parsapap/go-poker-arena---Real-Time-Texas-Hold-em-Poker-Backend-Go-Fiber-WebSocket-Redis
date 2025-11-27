# Login Credentials

## Test User Created Successfully! ✅

Your backend is working perfectly. A test user has been created for you.

### Login Credentials:

**Username:** `testuser_1764254810`  
**Password:** `password123`

### How to Login:

1. Go to: http://localhost:3000/login
2. Enter the username and password above
3. Click "Login"
4. You should be redirected to the lobby with 1,000 chips!

---

## Create Your Own Account

If you want to create your own account:

1. Go to: http://localhost:3000/register
2. Fill in:
   - **Username:** Choose any username (e.g., `myusername`)
   - **Email:** Any email (e.g., `my@email.com`)
   - **Password:** At least 6 characters (e.g., `mypassword123`)
3. Click "Create Account"
4. You'll be automatically logged in with 1,000 starting chips!

---

## Why "Invalid Credentials" Error Happened

The "invalid credentials" error happens when:
1. ❌ User doesn't exist in database yet
2. ❌ Wrong password
3. ❌ Wrong username

**Solution:** Either:
- Use the test credentials above
- Register a new account first, then login

---

## Quick Test Commands

### Test Registration (creates new user):
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"myuser","email":"my@email.com","password":"password123"}'
```

### Test Login:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"myuser","password":"password123"}'
```

### Run Full Test Suite:
```bash
./test-auth.sh
```

This will:
- Check if backend is running
- Create a new test user
- Test login
- Test protected endpoints
- Give you credentials to use

---

## Troubleshooting

### "Failed to fetch"
**Problem:** Backend not running  
**Solution:**
```bash
cd backend
go run cmd/server/main.go
```

### "username or email already exists"
**Problem:** User already registered  
**Solution:** Just login with existing credentials or use a different username

### "invalid credentials"
**Problem:** User doesn't exist or wrong password  
**Solution:** 
1. Register first at http://localhost:3000/register
2. Or use the test credentials above

### Check if backend is running:
```bash
curl http://localhost:8080/healthz
# Should return: {"status":"ok","service":"go-poker-arena","version":"1.0.0"}
```

---

## Database Users

To see all users in the database:
```bash
# Connect to PostgreSQL
docker exec -it poker-postgres psql -U poker -d poker_arena

# List all users
SELECT id, username, email, chips, wins, losses FROM users;

# Exit
\q
```

To delete a user (if needed):
```sql
DELETE FROM users WHERE username = 'testuser';
```

---

## Summary

✅ Backend is working  
✅ Test user created  
✅ Login endpoint working  
✅ Registration endpoint working  

**Just use the credentials above to login, or register a new account!**

---

## Next Steps

1. Login with test credentials
2. Explore the lobby
3. Create or join a room
4. Start playing poker!

Enjoy! 🎰🃏
