#!/bin/sh

email=$1
if [ -z "$email" ]; then
  email=1@local
fi

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

echo "------------- [create verification flow] -------------"
responseToChoozeMethod=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/verification/browser)
echo $responseToChoozeMethod | jq 

actionUrlToSentEmail=$(echo $responseToChoozeMethod | jq -r '.ui.action')
csrfToken=$(echo $responseToChoozeMethod | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete verification flow (send verification email)] -------------"
responseToSentEmail=$(curl -v -s -X POST \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "email": "'$email'", "method": "code"}' \
  "$actionUrlToSentEmail")
actionUrlToComplete=$(echo $responseToSentEmail | jq -r '.ui.action')
echo $responseToSentEmail | jq 

read -p "please input code emailed to you: " code

echo "\n\n\n------------- [complete verification flow (send verification code)] -------------"
responseToPassedChallenge=$(curl -v -s -X POST \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "code": "'$code'", "method": "code"}' \
  "$actionUrlToComplete&code=$code")
echo $responseToPassedChallenge | jq 

