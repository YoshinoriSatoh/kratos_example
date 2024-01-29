#!/bin/sh

IFS='
'

updateEmail=$1
if [ -z "$updateEmail" ]; then
  updateEmail=updated-1@local
fi

updatePassword=$2
if [ -z "$updatePassword" ]; then
  updatePassword=updated-overwatch2024
fi

updateNickname=$3
if [ -z "$updateNickname" ]; then
  updateNickname=updated-nickname
fi

updateBirthdate=$4
if [ -z "$updateBirthdate" ]; then
  updateBirthdate=2000-01-01
fi

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534


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


echo "------------- [create settings flow (method: profile)] -------------"
responseCreateSettingsFlow=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/settings/browser)
echo $responseCreateSettingsFlow | jq 

actionUrl=$(echo $responseCreateSettingsFlow | jq -r '.ui.action')
csrfToken=$(echo $responseCreateSettingsFlow | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete settings flow (method: profile)] -------------"
responseCompleteSettingsFlow=$(curl -v -s -X POST \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "method": "profile", "traits": { "email": "'$updateEmail'", "nickname": "'$updateNickname'", "birthdate": "'$updateBirthdate'" }}' \
  "$actionUrl")
echo $responseCompleteSettingsFlow | jq 
