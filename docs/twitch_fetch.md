# Twitch Fetch Pattern

Documentation for API calls in `services/twitch/fetch/`

## Overview

The `services/twitch/fetch/` package contains functions for making Twitch Helix API calls. All functions should follow a consistent pattern for URL building, request execution, and response handling.

## 1. URL Building Pattern

### Correct Pattern: `twitch.BuildURL()`

```go
opts := map[string]any{
    "param_name": value,
}
url := twitch.BuildURL(twitch.HelixBaseURL+"/endpoint", opts)
```

**Key Benefits:**
- Type-safe parameter handling
- Consistent URL construction
- Handles `?` vs `&` automatically
- Works with pagination cursors

### Incorrect Pattern: `fmt.Sprintf()`

```go
// WRONG - Avoid this pattern
url := fmt.Sprintf("%s?param=%s", urlAPI, value)
```

## 2. Request Execution Pattern

### GET/DELETE Requests (No Body)

```go
result, err := twitch.ExecuteRequest[ResponseType]("GET", url, 200)
if err != nil {
    helpers.Logf(helpers.DEBUG, "[TWITCH] FunctionName: params=%v", values)
    return nil, err
}
```

### POST/PATCH/PUT Requests (With Body)

```go
body := map[string]any{"field": value}
// OR use a request struct
result, err := twitch.ExecuteJSONRequest[ResponseType, map[string]any]("POST", url, body, 201)
if err != nil {
    helpers.Logf(helpers.DEBUG, "[TWITCH] FunctionName: params=%v", values)
    return nil, err
}
```

### Request with Custom Struct

```go
result, err := twitch.ExecuteJSONRequest[ResponseType, RequestStruct]("POST", url, request, 201)
```

## 3. Token-Match Functions

Functions where the Twitch API requires IDs to match the auth token must use `globals.GetState().GetTwitchUser()` directly:

```go
func FunctionName(requiredParam string) (*ReturnType, error) {
    user := globals.GetState().GetTwitchUser()

    opts := map[string]any{
        "required_id": user.UserID,
        "param":      requiredParam,
    }

    url := twitch.BuildURL(twitch.HelixBaseURL+"/endpoint", opts)
    // ... rest of function
}
```

**When to Enforce Token Match:**
- API docs say "ID must match the user ID in the OAuth token"
- Function operates on the authenticated user's channel/data

**When to Allow Flexible IDs:**
- Query endpoints (e.g., `GetChannelInformation`, `GetStreams`)
- Lookup functions (e.g., `GetUserData`)
- Pagination-based queries

## 4. Pagination Pattern

### Response Struct

```go
type GetXxxResponse struct {
    Data       []Xxx      `json:"data"`
    Pagination Pagination `json:"pagination"`
}
```

### Function with Pagination

```go
func GetXxx(req *twitch.PaginationRequest) (*twitch.PaginationData[Xxx], error) {
    opts := map[string]any{}

    if req != nil {
        if req.Cursor != "" {
            opts["after"] = req.Cursor
        }
        if req.Quantity > 0 {
            opts["first"] = req.Quantity
        }
    }

    url := twitch.BuildURL(twitch.HelixBaseURL+"/endpoint", opts)

    result, err := twitch.ExecuteRequest[twitch.PaginationData[Xxx]]("GET", url, 200)
    if err != nil {
        helpers.Logf(helpers.DEBUG, "[TWITCH] GetXxx: params=%v", ...)
        return nil, err
    }

    result.GetNext = func() *twitch.PaginationData[Xxx] {
        GetXxx(&twitch.PaginationRequest{
            Cursor:   result.Pagination.Cursor,
            Quantity: req.Quantity,
        })
        return result
    }

    return result, nil
}
```

## 5. Error Handling

### Logging Pattern

```go
if err != nil {
    helpers.Logf(helpers.DEBUG, "[TWITCH] FunctionName: relevant_params=%v", values)
    return nil, err  // or return err
}
```

### Error Messages

- Use `helpers.DEBUG` for expected failures (API errors, not found)
- Use `helpers.ERROR` for unexpected failures (marshaling, network)
- Include relevant parameters for debugging

## 6. File Status

| Status | Files | Notes |
|--------|-------|-------|
| ✓ Correct | `polls.go` | GetPolls, CreatePoll, EndPoll |
| ✓ Correct | `predictions.go` | GetPredictions, CreatePrediction, EndPrediction |
| ✓ Correct | `channel_points.go` | All 7 functions |
| ✓ Correct | `chat.go` | GetChatters, GetChannelEmotes, GetChatSettings, etc. |
| ✓ Correct | `streams.go` | GetStreamKey, GetStreams, GetFollowedStreams |
| ✓ Correct | `charity.go` | GetCharityCampaign, GetCharityCampaignDonations |
| ✓ Correct | `games.go` | GetTopGames, GetGames |
| ✓ Correct | `search.go` | SearchCategories, SearchChannels |
| ✓ Correct | `subscriptions.go` | GetBroadcasterSubscriptions, CheckUserSubscription |
| ✓ Correct | `videos.go` | GetVideos (with request struct), DeleteVideos |
| ✓ Correct | `moderation.go` | GetBannedUsers, GetBlockedTerms, GetModerators, etc. |
| ✓ Correct | `channels.go` | ModifyChannelInformation, GetChannelEditors, etc. |
| ✓ Correct | `whispers.go` | SendWhisper |
| ✓ Correct | `users.go` | All functions refactored |

## 7. Code Examples

### GET Request Example (polls.go)

```go
func GetPolls(pollIDs []string) ([]Poll, error) {
    user := globals.GetState().GetTwitchUser()

    opts := map[string]any{
        "broadcaster_id": user.UserID,
    }
    for _, id := range pollIDs {
        opts["id"] = id
    }

    url := twitch.BuildURL(twitch.HelixBaseURL+"/polls", opts)

    result, err := twitch.ExecuteRequest[GetPollsResponse]("GET", url, 200)
    if err != nil {
        helpers.Logf(helpers.DEBUG, "[TWITCH] GetPolls: broadcasterID=%v", user.UserID)
        return nil, err
    }

    return result.Data, nil
}
```

### POST Request with Body (predictions.go)

```go
func CreatePrediction(req CreatePredictionRequest) (*Prediction, error) {
    user := globals.GetState().GetTwitchUser()

    opts := map[string]any{
        "broadcaster_id": user.UserID,
    }
    url := twitch.BuildURL(twitch.HelixBaseURL+"/predictions", opts)

    result, err := twitch.ExecuteJSONRequest[GetPredictionsResponse, CreatePredictionRequest]("POST", url, req, 201)
    if err != nil {
        helpers.Logf(helpers.DEBUG, "[TWITCH] CreatePrediction: title=%v", req.Title)
        return nil, err
    }

    if len(result.Data) == 0 {
        return nil, nil
    }

    return &result.Data[0], nil
}
```

### Request Struct Pattern (videos.go)

```go
type GetVideosRequest struct {
    VideoIDs []string
    UserID   string  // Required
    GameID   string  // Required
    Period   string
    Sort     string
    Type     string
}

func GetVideos(req GetVideosRequest) (*twitch.PaginationData[Video], error) {
    opts := map[string]any{}

    if req.UserID != "" {
        opts["user_id"] = req.UserID
    }
    if req.GameID != "" {
        opts["game_id"] = req.GameID
    }
    if req.Period != "" {
        opts["period"] = req.Period
    }
    // ... rest of function
}
```

## 8. Adding New Functions

When adding a new Twitch API endpoint:

1. **Check API docs** - Does the endpoint require token match?
2. **Create response struct** - Use `GetXxxResponse` naming convention
3. **Build URL** - Use `twitch.BuildURL()` with `map[string]any`
4. **Execute request** - Use `ExecuteRequest` or `ExecuteJSONRequest`
5. **Handle response** - Check `len(result.Data)` for empty results
6. **Log errors** - Include relevant parameters
7. **Add tests** - If applicable

### Checklist

- [ ] URL built with `BuildURL`
- [ ] Request uses `ExecuteRequest` or `ExecuteJSONRequest`
- [ ] Token-match enforced where required
- [ ] Error logging includes debug info
- [ ] Response struct follows naming convention
- [ ] Build passes