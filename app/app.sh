#!/bin/bash
CHROME_URL=`curl HTTP://192.168.31.5:9222/json|grep -o -E 'ws:[^"]*'`
echo $CHROME_URL
/go/bin/app -devtools-ws-url=$CHROME_URL
