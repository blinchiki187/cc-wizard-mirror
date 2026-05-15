#!/bin/bash

export PATH=/home/node/.local/bin/:$PATH

/home/node/wizard/gobackend &

npm run dev -- --host
