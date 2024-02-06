#!/bin/sh

IFS='
'

updatePassword=$2
if [ -z "$updatePassword" ]; then
  updatePassword=updated-overwatch2024
fi

publicEndpoint=http://localhost:4433
adminEndpoint=http://localhost:4434

sessionToken=$(cat .session_token)
echo $sessionToken

if [ -z "$sessionToken" ]; then
  echo "not logged in"
  exit 1
fi

echo "------------- [create settings flow (method: password)] -------------"
responseCreateSettingsFlow=$(curl -v -s -X GET \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: $sessionToken" \
  $publicEndpoint/self-service/settings/api)
echo $responseCreateSettingsFlow | jq 

actionUrl=$(echo $responseCreateSettingsFlow | jq -r '.ui.action')

echo "\n\n\n------------- [complete settings flow (method: password)] -------------"
responseCompleteSettingsFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: $sessionToken" \
  -d '{"password": "'$updatePassword'", "method": "password"}' \
  "$actionUrl")
echo $responseCompleteSettingsFlow | jq 

read -p "please input code emailed to you: " code

verificationFlowId=$(echo $responseCompleteSettingsFlow | jq -r -c '.continue_with[] | select(.action=="show_verification_ui") | .flow.id')
echo $verificationFlowId 

echo "\n\n\n------------- [complete verification flow (send verification flow)] -------------"
responseUpdateVerificationFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"code": "'$code'", "method": "code"}' \
  "$publicEndpoint/self-service/verification?flow=$verificationFlowId")
echo $responseUpdateVerificationFlow | jq 

