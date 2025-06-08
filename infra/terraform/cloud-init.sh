#!/bin/bash
set -e
exec > /tmp/startup.log 2>&1

sudo apt-get update
sudo apt-get install -y nodejs npm git

APP_DIR="$HOME/app"
git clone ${repo_url} "$APP_DIR"
cd "$APP_DIR"
git checkout ${repo_branch}
npm install
npm start &