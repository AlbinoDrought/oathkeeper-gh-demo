# token-exchange

This service listens on port :3001 and performs all of your authentication by converting tokens into user details.

## Expected Output

`curl -H "Authorization: Bearer github 1234" http://localhost:3001`

```
{
  "sub": "1234",
  "extra": {
    "provider": "github",
    "username": "ghost",
    "email": "ghost@github.com"
  }
}
```
