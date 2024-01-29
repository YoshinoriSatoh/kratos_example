#!/bin/sh

email=$1
if [ -z "$email" ]; then
  email=1@local
fi

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

sessionToken=$(cat .session_token)
echo $sessionToken

if [ -z "$sessionToken" ]; then
  echo "not logged in"
  exit 1
fi

echo "------------- [create verification flow] -------------"
responseToChoozeMethod=$(curl -v -s -X GET \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/verification/api)
echo $responseToChoozeMethod | jq 

actionUrlToSentEmail=$(echo $responseToChoozeMethod | jq -r '.ui.action')

echo "\n\n\n------------- [complete verification flow (send verification email)] -------------"
responseToSentEmail=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"email": "'$email'", "method": "code"}' \
  "$actionUrlToSentEmail")
actionUrlToComplete=$(echo $responseToSentEmail | jq -r '.ui.action')
echo $responseToSentEmail | jq 

read -p "please input code emailed to you: " code

echo "\n\n\n------------- [complete verification flow (send verification code)] -------------"
responseToPassedChallenge=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"code": "'$code'", "method": "code"}' \
  "$actionUrlToComplete&code=$code")
echo $responseToPassedChallenge | jq 

