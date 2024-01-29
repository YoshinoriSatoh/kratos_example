#!/bin/sh

publicEndpoint=http://localhost:4533
adminEndpoint=http://localhost:4534

echo "------------- [create logout flow] -------------"
responseCreateLogoutFlow=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  $publicEndpoint/self-service/logout/browser)
echo $responseCreateLogoutFlow | jq 

actionUrlUpdateLogout=$(echo $responseCreateLogoutFlow | jq -r '.logout_url')

echo "\n\n\n------------- [submit logout flow] -------------"
responseSubmitLogoutFlow=$(curl -v -s -X GET \
  -c .session_cookie -b .session_cookie \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  "$actionUrlUpdateLogout")
echo $responseSubmitLogoutFlow | jq 

