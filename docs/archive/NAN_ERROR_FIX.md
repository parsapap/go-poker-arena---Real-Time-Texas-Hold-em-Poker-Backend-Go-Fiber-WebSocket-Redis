# NaN Error Fix

## Problem Fixed ✅
**Error:** "The specified value 'NaN' cannot be parsed, or is out of range"

This error was caused by `currentBet` being `undefined` or `NaN`, which made the range input's `min` attribute invalid.

## Changes Made

### 1. Fixed Range Input
```typescript
// Before
<input
  type="range"
  min={currentBet * 2}  // ❌ NaN * 2 = NaN
  max={user?.chips || 5000}
  ...
/>

// After
<input
  type="range"
  min={Math.max((currentBet || 0) * 2, 10)}  // ✅ Always a valid number
  max={user?.chips || 5000}
  ...
/>
```

### 2. Fixed Quick Bet Buttons
```typescript
// Before
<button onClick={() => setRaiseAmount(currentBet * 2)}>2x</button>  // ❌ NaN * 2

// After
<button onClick={() => setRaiseAmount(Math.max((currentBet || 0) * 2, 20))}>2x</button>  // ✅
```

### 3. Fixed Call Button
```typescript
// Before
{currentBet > 0 ? `Call $${currentBet}` : 'Check'}  // ❌ Shows "Call $NaN"

// After
{(currentBet || 0) > 0 ? `Call $${currentBet || 0}` : 'Check'}  // ✅ Shows "Call $0" or "Check"
```

### 4. Fixed Action Handler
```typescript
// Before
onClick={() => handleAction(currentBet > 0 ? 'call' : 'check')}  // ❌ NaN > 0 is false

// After
onClick={() => handleAction((currentBet || 0) > 0 ? 'call' : 'check')}  // ✅
```

## Why This Happened

When the game page loads, `currentBet` starts as `0` (from the store's initial state). However, if the WebSocket doesn't send game state immediately, or if there's a timing issue, `currentBet` can be `undefined`, which causes:

- `undefined * 2` = `NaN`
- `NaN` as a `min` attribute = Invalid input error

## Solution Pattern

Always use the nullish coalescing operator (`||`) with a default value:
```typescript
(currentBet || 0)  // If currentBet is undefined/null/NaN, use 0
```

And use `Math.max()` to ensure minimum values:
```typescript
Math.max((currentBet || 0) * 2, 10)  // At least 10
```

## Files Modified

- ✅ `frontend/app/game/[roomId]/page.tsx` - Fixed all NaN issues

## Result

✅ No more NaN errors  
✅ Range input always has valid min/max  
✅ Buttons always set valid raise amounts  
✅ Call button shows correct amount  

The game page should now load without the NaN error!
