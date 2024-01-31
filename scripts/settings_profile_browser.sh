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

read -p "please input code emailed to you: " code

verificationFlowId=$(echo $responseCompleteSettingsFlow | jq -r -c '.continue_with[] | select(.action=="show_verification_ui") | .flow.id')
echo $verificationFlowId 

echo "\n\n\n------------- [get verification flow] -------------"
responseGetVerificationFlow=$(curl -v -s -X GET \
  -b .session_cookie -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  "$publicEndpoint/self-service/verification/flows?id=$verificationFlowId")
echo $responseGetVerificationFlow | jq 

csrfToken=$(echo $responseGetVerificationFlow | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete verification flow (send verification flow)] -------------"
responseUpdateVerificationFlow=$(curl -v -s -X POST \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "code": "'$code'", "method": "code"}' \
  "$publicEndpoint/self-service/verification?flow=$verificationFlowId")
echo $responseUpdateVerificationFlow | jq 

