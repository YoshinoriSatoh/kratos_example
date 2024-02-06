#!/bin/sh

publicEndpoint=http://localhost:4433
adminEndpoint=http://localhost:4434

sessionToken=$(cat .session_token)
echo $sessionToken

if [ -z "$sessionToken" ]; then
  echo "already logged out"
  exit 1
fi

echo "\n\n\n------------- [perform logout] -------------"
responsePerformLogoutFlow=$(curl -v -s -X DELETE \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"session_token": "'$sessionToken'"}' \
  $publicEndpoint/self-service/logout/api)
echo $responsePerformLogoutFlow | jq 

echo "" > .session_token
