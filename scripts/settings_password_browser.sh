#!/bin/sh

IFS='
'

updatePassword=$2
if [ -z "$updatePassword" ]; then
  updatePassword=updated-overwatch2024
fi

publicEndpoint=http://localhost:4433
adminEndpoint=http://localhost:4434

echo "------------- [create settings flow (method: password)] -------------"
responseCreateSettingsFlow=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/settings/browser)
echo $responseCreateSettingsFlow | jq 

actionUrl=$(echo $responseCreateSettingsFlow | jq -r '.ui.action')
csrfToken=$(echo $responseCreateSettingsFlow | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete settings flow (method: password)] -------------"
responseCompleteSettingsFlow=$(curl -v -s -X POST \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "password": "'$updatePassword'", "method": "password"}' \
  "$actionUrl")
echo $responseCompleteSettingsFlow | jq 


