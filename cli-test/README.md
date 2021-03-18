# cli-test

This command line application retrieves a token using the [Github OAuth](https://github.com/settings/applications/new) device flow and sends it to the API.

## Expected Output

`OATHKEEPER_GH_DEMO_CLIENT_ID=<your github client ID> ./cli-test`

```
INFO[0000] OATHKEEPER_GH_DEMO_API_URL was not set, using default  default="http://localhost/"
WARN[0000] failed reading persisted access token, will regenerate  error="open access-token: no such file or directory"
INFO[0000] please enter this code :)                     code=1234-ABCD url="https://github.com/login/device"
WARN[0007] poll failed with expected error               error-code=authorization_pending error-description="The authorization request is still pending." error-uri="https://docs.github.com/developers/apps/authorizing-oauth-apps#error-codes-for-the-device-flow"
INFO[0014] retrieved new token!                          access-token=12345
INFO[0014] hit our API :)                                output="Hello user 1234\nYou logged in with github using the username ghost and the email ghost@github.com\n"
```
