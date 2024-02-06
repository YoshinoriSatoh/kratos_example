#!/bin/sh

publicEndpoint=http://localhost:4433
adminEndpoint=http://localhost:4434

echo "------------- [check session] -------------"
responseSessionWhoami=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H 'X-Session-Token: '$sessionToken \
  $publicEndpoint/sessions/whoami)
echo $responseSessionWhoami | jq 