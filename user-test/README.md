# user-test

This is the main service. It listens on port :3000, receives auth from Oathkeeper's X-User-* headers, and echos some information back.

## Expected Output

`curl -H "X-User-ID=1234" -H "X-User-Provider=github" -H "X-User-Username=ghost" -H "X-User-Email=ghost@github.com" http://localhost:3000

```
Hello user 1234
You logged in with github using the username ghost and the email ghost@github.com
```
