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

echo "------------- [create login flow] -------------"
responseCreateLoginFlow=$(curl -v -s -X GET \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/login/api)
echo $responseCreateLoginFlow | jq 

actionUrlSubmitLogin=$(echo $responseCreateLoginFlow | jq -r '.ui.action')

echo "\n\n\n------------- [submit login flow] -------------"
responseSubmitLoginFlow=$(curl -v -s -X POST \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d '{"identifier": "'$email'", "method": "password", "password": "'$password'"}' \
  "$actionUrlSubmitLogin")
echo $responseSubmitLoginFlow | jq 

sessionToken=$(echo $responseSubmitLoginFlow | jq -r '.session_token')
echo $sessionToken > .session_token

