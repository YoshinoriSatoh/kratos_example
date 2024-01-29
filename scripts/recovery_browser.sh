#!/bin/sh

IFS='
'

email=$1
if [ -z "$email" ]; then
  email=1@local
fi

password=$2
if [ -z "$password" ]; then
  password=overwatch2023
fi

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

echo "------------- [create recovery flow] -------------"
responseToChoozeMethod=$(curl -v -s -X GET \
  -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/recovery/browser)
echo $responseToChoozeMethod | jq 

actionUrlToSentEmail=$(echo $responseToChoozeMethod | jq -r '.ui.action')
csrfToken=$(echo $responseToChoozeMethod | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete recovery flow (send recovery email)] -------------"
responseToSentEmail=$(curl -v -s -X POST \
  -b .session_cookie -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "email": "'$email'", "method": "code"}' \
  "$actionUrlToSentEmail")
actionUrlToComplete=$(echo $responseToSentEmail | jq -r '.ui.action')
echo $responseToSentEmail | jq 
echo $actionUrlToComplete

read -p "please input code emailed to you: " code

echo "\n\n\n------------- [complete recovery flow (send recovery code)] -------------"
responseToPassedChallenge=$(curl -v -s -X POST \
  -b .session_cookie -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "code": "'$code'", "method": "code"}' \
  "$actionUrlToComplete")
echo $responseToPassedChallenge | jq 

redirectBrowserTo=$(echo $responseToPassedChallenge | jq -r '.redirect_browser_to')
echo $redirectBrowserTo
settingsFlowId=$(echo $redirectBrowserTo | rev | cut -c -36 | rev)
echo $settingsFlowId

echo "\n\n\n------------- [get settings flow] -------------"
responseGetSettingsFlow=$(curl -v -s -X GET \
  -b .session_cookie -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  "$publicEndpoint/self-service/settings/flows?id=$settingsFlowId")
echo $responseGetSettingsFlow | jq 

csrfToken=$(echo $responseGetSettingsFlow | jq -r '.ui.nodes[] | select(.attributes.name=="csrf_token") | .attributes.value') 

echo "\n\n\n------------- [complete settings flow] -------------"
responseCompleteSettingsFlow=$(curl -v -s -X POST \
  -b .session_cookie -c .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"csrf_token": "'$csrfToken'", "password": "'$password'", "method": "password"}' \
  "$publicEndpoint/self-service/settings?flow=$settingsFlowId")
echo $responseCompleteSettingsFlow | jq 
