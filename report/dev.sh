#!/bin/bash
cd frontend

nvm use

npm run dev &

cd ../backend

go run . &