#!/usr/bin/env bash
export DEBIAN_FRONTEND=noninteractive

if ! dpkg -s busybox-static software-properties-common >/dev/null 2>&1; then
    sudo apt-get -y update && sudo apt-get -y install busybox-static software-properties-common
fi

if ! dpkg -s golang-go >/dev/null 2>&1; then
    sudo add-apt-repository -y ppa:longsleep/golang-backports
    sudo apt-get -y update &&  sudo apt-get -y install ca-certificates curl gcc git golang-go
fi

if ! dpkg -s docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin >/dev/null 2>&1; then
    sudo install -m 0755 -d /etc/apt/keyrings && sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc && sudo chmod a+r /etc/apt/keyrings/docker.asc
    echo deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
    sudo apt-get -y update && sudo apt-get -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin docker-compose
fi

sudo usermod -aG docker ubuntu
sudo chgrp ubuntu /var/run/docker.sock
sudo chmod +rw /var/run/docker.sock
