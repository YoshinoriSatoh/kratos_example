#!/bin/sh

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

sessionToken=$(cat .session_token)
echo $sessionToken

if [ -z "$sessionToken" ]; then
  echo "not logged in"
  exit 1
fi

echo "------------- [check session] -------------"
responseSessionWhoami=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H 'X-Session-Token: '$sessionToken \
  $publicEndpoint/sessions/whoami)
echo $responseSessionWhoami | jq 