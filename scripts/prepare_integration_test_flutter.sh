#!/bin/bash
#
# Copyright Avast Software. All Rights Reserved.
# Copyright Gen Digital Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running $0"

ROOT=`pwd`

echo "starting containers..."
cd $ROOT/test/integration/fixtures
docker pull jaegertracing/all-in-one:1.62.0
docker pull devopsfaith/krakend:2.1.3
docker pull aholovko/cognito-local:0.2.2
docker images
(source .env && docker-compose -f docker-compose.yml up --force-recreate -d)

sleep 120

echo "running healthcheck..."

healthCheckFailed=0

# healthCheck function
RED=$(tput setaf 1)
GREEN=$(tput setaf 2)
AQUA=$(tput setaf 6)
NONE=$(tput sgr0)
healthCheck() {
	sleep 1
    n=0
    maxAttempts=60
    if [ "" != "$4" ]
    then
	   maxAttempts=$4
    fi

	echo "running health check : app=$1 url=$2 timeout=$maxAttempts seconds"

	until [ $n -ge $maxAttempts ]
	do

    docker ps -a

	  response=$(curl -H 'Cache-Control: no-cache' -o response.txt -s -w "%{http_code}" "$2")
	  echo "running health check : httpResponseCode=$response"

	  if [ "$response" == "$3" ]
	  then
	    echo "${GREEN}$1 $2 is up ${NONE}"
		  break
	  fi

	  n=$((n+1))
	  if [ $n -eq $maxAttempts ]
	  then
	     echo "${RED}failed health check : app=$1 url=$2 responseCode=$response ${NONE}"
	     healthCheckFailed=1
	     docker-compose -f docker-compose.yml logs --no-color >& docker-compose.log
	     cat ./docker-compose.log
	  fi
	  sleep 1
	 done
}

# healthcheck
healthCheck vc-rest http://localhost:8075/version 200 180
healthCheck did-resolver http://did-resolver.trustbloc.local:8072/healthcheck 200 180

if [ $healthCheckFailed == 1 ]
then
  echo "${RED}some health checks failed, see logs above"
  exit 1
fi
