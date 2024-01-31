#!/bin/sh

IFS='
'

updateEmail=$1
if [ -z "$updateEmail" ]; then
  updateEmail=updated-1@local
fi

updateNickname=$2
if [ -z "$updateNickname" ]; then
  updateNickname=updated-nickname
fi

updateBirthdate=$3
if [ -z "$updateBirthdate" ]; then
  updateBirthdate=2000-01-01
fi

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

sessionToken=$(cat .session_token)
echo $sessionToken

if [ -z "$sessionToken" ]; then
  echo "not logged in"
  exit 1
fi

echo "------------- [create settings flow (method: profile)] -------------"
responseCreateSettingsFlow=$(curl -v -s -X GET \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: $sessionToken" \
  $publicEndpoint/self-service/settings/api)
echo $responseCreateSettingsFlow | jq 

actionUrl=$(echo $responseCreateSettingsFlow | jq -r '.ui.action')

echo "\n\n\n------------- [complete settings flow (method: profile)] -------------"
responseCompleteSettingsFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: $sessionToken" \
  -d '{"method": "profile", "traits": { "email": "'$updateEmail'", "nickname": "'$updateNickname'", "birthdate": "'$updateBirthdate'" }}' \
  "$actionUrl")
echo $responseCompleteSettingsFlow | jq 
