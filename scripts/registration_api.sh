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

publicEndpoint=http://localhost:4433
adminEndpoint=http://localhost:4434

echo "------------- [create registration flow] -------------"
responseCreateRegistrationFlow=$(curl -v -s -X GET \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/registration/api)
echo $responseCreateRegistrationFlow | jq 

actionUrl=$(echo $responseCreateRegistrationFlow | jq -r '.ui.action')

echo "\n\n\n------------- [update registration flow] -------------"
responseUpdateRegistrationFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"traits.email": "'$email'", "password": "'$password'", "method": "password"}' \
  "$actionUrl") 
echo $responseUpdateRegistrationFlow | jq

read -p "please input code emailed to you: " code

verificationFlowId=$(echo $responseUpdateRegistrationFlow | jq -r -c '.continue_with[] | select(.action=="show_verification_ui") | .flow.id')
echo $verificationFlowId 

echo "\n\n\n------------- [complete verification flow (send verification flow)] -------------"
responseUpdateVerificationFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"code": "'$code'", "method": "code"}' \
  "$publicEndpoint/self-service/verification?flow=$verificationFlowId")
echo $responseUpdateVerificationFlow | jq 

