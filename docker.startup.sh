#!/bin/bash

nohup /headless-shell/headless-shell --no-sandbox --headless --disable-gpu --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 &>/dev/null &
/gql-server